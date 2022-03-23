package resultstore

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/iver-wharf/wharf-cmd/pkg/worker/workermodel"
)

var (
	fileNameArtifactEvents = "artifacts.json"
)

func (s *store) AddArtifactEvent(stepID uint64, artifactMeta workermodel.ArtifactMeta) error {
	s.artifactMutex.LockKey(stepID)
	defer s.artifactMutex.UnlockKey(stepID)
	if s.frozen {
		return ErrFrozen
	}
	list, err := s.readArtifactEventsFile(stepID)
	if err != nil {
		return err
	}
	list.LastID++
	artifactEvent := ArtifactEvent{
		ArtifactID: list.LastID,
		StepID:     stepID,
		Name:       artifactMeta.Name,
	}
	list.ArtifactEvents = append(list.ArtifactEvents, artifactEvent)
	if err := s.writeArtifactEventsFile(stepID, list); err != nil {
		return err
	}
	s.pubArtifactEvent(artifactEvent)
	return nil
}

func (s *store) readArtifactEventsFile(stepID uint64) (ArtifactEventList, error) {
	file, err := s.fs.OpenRead(s.resolveArtifactEventsPath(stepID))
	if errors.Is(err, fs.ErrNotExist) {
		return ArtifactEventList{}, nil
	}
	if err != nil {
		return ArtifactEventList{}, fmt.Errorf("open artifact events file for reading: %w", err)
	}
	defer file.Close()
	dec := json.NewDecoder(file)
	var list ArtifactEventList
	if err := dec.Decode(&list); err != nil {
		return ArtifactEventList{}, fmt.Errorf("decode artifact events: %w", err)
	}
	for i := range list.ArtifactEvents {
		list.ArtifactEvents[i].StepID = stepID
	}
	return list, nil
}

func (s *store) writeArtifactEventsFile(stepID uint64, list ArtifactEventList) error {
	file, err := s.fs.OpenWrite(s.resolveArtifactEventsPath(stepID))
	if err != nil {
		return fmt.Errorf("open artifact events file for writing: %w", err)
	}
	defer file.Close()
	enc := json.NewEncoder(file)
	if err := enc.Encode(&list); err != nil {
		return fmt.Errorf("encode artifact events: %w", err)
	}
	return nil
}

func (s *store) resolveArtifactEventsPath(stepID uint64) string {
	return filepath.Join(dirNameSteps, fmt.Sprint(stepID), fileNameArtifactEvents)
}

func (s *store) SubAllArtifactEvents(buffer int) (<-chan ArtifactEvent, error) {
	s.artifactSubMutex.RLock()
	defer s.artifactSubMutex.RUnlock()
	ch := s.artifactPubSub.SubBuf(buffer)
	events, err := s.listAllArtifactEvents()
	if err != nil {
		return nil, fmt.Errorf("read all existing artifact events: %w", err)
	}
	go s.artifactPubSub.WithOnly(ch).PubSliceSync(events)
	return ch, nil
}

func (s *store) listAllArtifactEvents() ([]ArtifactEvent, error) {
	stepIDs, err := s.listAllStepIDs()
	if err != nil {
		return nil, err
	}
	var artifacts []ArtifactEvent
	for _, stepID := range stepIDs {
		list, err := s.readArtifactEventsFile(stepID)
		if err != nil {
			return nil, err
		}
		artifacts = append(artifacts, list.ArtifactEvents...)
	}
	return artifacts, nil
}

func (s *store) UnsubAllArtifactEvents(artifactCh <-chan ArtifactEvent) error {
	return s.artifactPubSub.Unsub(artifactCh)
}

func (s *store) pubArtifactEvent(artifactEvent ArtifactEvent) {
	// Locking to prevent new data being added during fetching existing data
	// part of when a new subscription is made.
	s.artifactSubMutex.RLock()
	s.artifactPubSub.PubSync(artifactEvent)
	s.artifactSubMutex.RUnlock()
}
