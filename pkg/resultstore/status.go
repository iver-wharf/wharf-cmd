package resultstore

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"sync/atomic"
	"time"

	"github.com/iver-wharf/wharf-cmd/pkg/worker"
)

func (s *store) AddStatusUpdate(stepID uint64, timestamp time.Time, newStatus worker.Status) error {
	updateID := atomic.AddUint64(&s.lastStatusID, 1)
	//s.statusMutex.Lock(stepID)
	//defer s.statusMutex.Unlock(stepID)
	list, err := s.readStatusUpdatesFile(stepID)
	if err != nil {
		return err
	}
	if len(list.StatusUpdates) > 0 &&
		list.StatusUpdates[len(list.StatusUpdates)-1].Status == newStatus.String() {
		return nil
	}
	statusUpdate := StatusUpdate{
		StepID:    stepID,
		UpdateID:  updateID,
		Timestamp: timestamp,
		Status:    newStatus.String(),
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
	var list StatusList
	if err := dec.Decode(&list); err != nil {
		return StatusList{}, fmt.Errorf("decode status updates: %w", err)
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
	return fmt.Sprintf("steps/%d/status.json", stepID)
}

func (s *store) SubAllStatusUpdates(buffer int) <-chan StatusUpdate {
	s.statusSubMutex.Lock()
	defer s.statusSubMutex.Unlock()
	// TODO: Feed all existing status updates into new channel
	ch := make(chan StatusUpdate, buffer)
	s.statusSubs = append(s.statusSubs, ch)
	return ch
}

func (s *store) UnsubAllStatusUpdates(statusCh <-chan StatusUpdate) bool {
	s.statusSubMutex.Lock()
	defer s.statusSubMutex.Unlock()
	for i, ch := range s.statusSubs {
		if ch == statusCh {
			if i != len(s.statusSubs)-1 {
				copy(s.statusSubs[i:], s.statusSubs[i+1:])
			}
			s.statusSubs = s.statusSubs[:len(s.statusSubs)-1]
			close(ch)
			return true
		}
	}
	return false
}

func (s *store) pubStatusUpdate(statusUpdate StatusUpdate) {
	s.statusSubMutex.RLock()
	for _, ch := range s.statusSubs {
		ch <- statusUpdate
	}
	s.statusSubMutex.RUnlock()
}
