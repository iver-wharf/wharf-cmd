package resultstore

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"time"

	"github.com/iver-wharf/wharf-cmd/pkg/worker/workermodel"
)

var (
	fileNameStatusUpdates = "status.json"
)

func (s *store) AddStatusUpdate(stepID uint64, timestamp time.Time, newStatus workermodel.Status) error {
	s.statusMutex.LockKey(stepID)
	defer s.statusMutex.UnlockKey(stepID)
	if s.frozen {
		return ErrFrozen
	}
	list, err := s.readStatusUpdatesFile(stepID)
	if err != nil {
		return err
	}
	if len(list.StatusUpdates) > 0 &&
		list.StatusUpdates[len(list.StatusUpdates)-1].Status == newStatus {
		return nil
	}
	list.LastID++
	updateID := list.LastID
	statusUpdate := StatusUpdate{
		StepID:    stepID,
		UpdateID:  updateID,
		Timestamp: timestamp,
		Status:    newStatus,
	}
	list.StatusUpdates = append(list.StatusUpdates, statusUpdate)
	if err := s.writeStatusUpdatesFile(stepID, list); err != nil {
		return err
	}
	s.pubStatusUpdate(statusUpdate)
	return nil
}

func (s *store) readStatusUpdatesFile(stepID uint64) (StatusList, error) {
	file, err := s.fs.OpenRead(s.resolveStatusPath(stepID))
	if errors.Is(err, fs.ErrNotExist) {
		return StatusList{}, nil
	}
	if err != nil {
		return StatusList{}, fmt.Errorf("open status updates file for reading: %w", err)
	}
	defer file.Close()
	dec := json.NewDecoder(file)
	if errors.Is(err, io.EOF) {
		return StatusList{}, nil
	}
	var list StatusList
	if err := dec.Decode(&list); err != nil {
		return StatusList{}, fmt.Errorf("decode status updates: %w", err)
	}
	for i := range list.StatusUpdates {
		list.StatusUpdates[i].StepID = stepID
	}
	return list, nil
}

func (s *store) writeStatusUpdatesFile(stepID uint64, list StatusList) error {
	file, err := s.fs.OpenWrite(s.resolveStatusPath(stepID))
	if err != nil {
		return fmt.Errorf("open status updates file for writing: %w", err)
	}
	defer file.Close()
	enc := json.NewEncoder(file)
	if err := enc.Encode(&list); err != nil {
		return fmt.Errorf("encode status updates: %w", err)
	}
	return nil
}

func (s *store) resolveStatusPath(stepID uint64) string {
	return filepath.Join(dirNameSteps, fmt.Sprint(stepID), fileNameStatusUpdates)
}

func (s *store) SubAllStatusUpdates(buffer int) (<-chan StatusUpdate, error) {
	s.statusSubMutex.RLock()
	defer s.statusSubMutex.RUnlock()
	ch := s.statusPubSub.SubBuf(buffer)
	updates, err := s.listAllStatusUpdates()
	if err != nil {
		return nil, fmt.Errorf("read all existing status updates: %w", err)
	}
	go func() {
		s.statusPubSub.WithOnly(ch).PubSliceSync(updates)
		if s.frozen {
			s.statusPubSub.Unsub(ch)
		}
	}()
	return ch, nil
}

func (s *store) listAllStatusUpdates() ([]StatusUpdate, error) {
	stepIDs, err := s.listAllStepIDs()
	if err != nil {
		return nil, err
	}
	var updates []StatusUpdate
	for _, stepID := range stepIDs {
		list, err := s.readStatusUpdatesFile(stepID)
		if err != nil {
			return nil, err
		}
		updates = append(updates, list.StatusUpdates...)
	}
	return updates, nil
}

func (s *store) UnsubAllStatusUpdates(statusCh <-chan StatusUpdate) error {
	return s.statusPubSub.Unsub(statusCh)
}

func (s *store) pubStatusUpdate(statusUpdate StatusUpdate) {
	// Additional locking as we want to pre-fetch existing data on new
	// subscriptions
	s.statusSubMutex.RLock()
	s.statusPubSub.PubSync(statusUpdate)
	s.statusSubMutex.RUnlock()
}
