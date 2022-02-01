package resultstore

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
)

func (s *store) AddArtifact(stepID uint64, artifactName string) (io.WriteCloser, error) {
	// TODO: implement this
	//artifactID := atomic.AddUint64(&s.lastArtifactID, 1)
	return nil, nil
}

func (s *store) resolveArtifactListMetaPath(stepID uint64) string {
	return fmt.Sprintf("steps/%d/artifacts.json", stepID)
}

func (s *store) resolveArtifactPath(stepID uint64, artifactID uint64) (string, error) {
	listMeta, err := s.readArtifactListMeta(stepID)
	if err != nil {
		return "", err
	}
	for _, meta := range listMeta.Artifacts {
		if meta.ArtifactID == artifactID {
			return meta.Path, nil
		}
	}
	return "", fmt.Errorf("step %d: artifact by ID %d not found in artifacts.json", stepID, artifactID)
}

func (s *store) readArtifactListMeta(stepID uint64) (ArtifactListMeta, error) {
	file, err := s.fs.OpenRead(s.resolveArtifactListMetaPath(stepID))
	if errors.Is(err, fs.ErrNotExist) {
		return ArtifactListMeta{}, nil
	}
	if err != nil {
		return ArtifactListMeta{}, fmt.Errorf("step %d: read artifact.json: %w", stepID, err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	var listMeta ArtifactListMeta
	if err := decoder.Decode(&listMeta); err != nil {
		return ArtifactListMeta{}, err
	}
	return listMeta, nil
}
