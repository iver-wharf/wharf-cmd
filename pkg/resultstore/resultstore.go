package resultstore

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"time"
)

type LogLine struct {
	StepID    uint64
	Line      string
	Timestamp time.Time
}

type ArtifactListMeta struct {
	Artifacts []ArtifactMeta `json:"artifacts"`
}

type ArtifactMeta struct {
	ArtifactID uint64 `json:"artifactId"`
	Name       string `json:"name"`
	Path       string `json:"path"`
}

type Store interface {
	AddLogLine(stepID uint64, line string) error
	AddStatusUpdate(stepID uint64, newStatus Status) error
	AddArtifact(stepID uint64, artifactName string) (io.WriteCloser, error)

	ReadAllLogLines(stepID uint64) ([]LogLine, error)
}

func NewStore(fs fs.FS) Store {
	return &store{
		fs: fs,
	}
}

type store struct {
	fs             fs.FS
	lastLogID      uint64
	lastStatusID   uint64
	lastArtifactID uint64
}

type stepID struct {
	stageID, stepID uint64
}

func (s *store) AddLogLine(stepID uint64, line string) error {
	//logID := atomic.AddUint64(&s.lastLogID, 1)
	// TODO: Open file with os.O_APPEND|os.O_WRONLY|os.O_CREATE
	file, err := s.fs.Open(s.resolveLogPath(stepID))
	return nil
}

func (s *store) AddStatusUpdate(stepID uint64, newStatus Status) error {
	//statusID := atomic.AddUint64(&s.lastStatusID, 1)
	return nil
}

func (s *store) AddArtifact(stepID uint64, artifactName string) (io.WriteCloser, error) {
	//artifactID := atomic.AddUint64(&s.lastArtifactID, 1)
	return nil, nil
}

func (s *store) ReadAllLogLines(stepID uint64) ([]LogLine, error) {
	file, err := s.fs.Open(s.resolveLogPath(stepID))
	return nil, nil
}

func (s *store) resolveLogPath(stepID uint64) string {
	return fmt.Sprintf("steps/%d/logs.log", stepID)
}

func (s *store) resolveStatusPath(stepID uint64) string {
	return fmt.Sprintf("steps/%d/status.json", stepID)
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
	file, err := s.fs.Open(s.resolveArtifactListMetaPath(stepID))
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
