package resultstore

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"time"

	"github.com/iver-wharf/wharf-cmd/pkg/worker"
)

type LogLine struct {
	StepID    uint64
	LogID     uint64
	Line      string
	Timestamp time.Time
}

type StatusList struct {
	StatusUpdates []StatusUpdate `json:"statusUpdates"`
}

type StatusUpdate struct {
	UpdateID  uint64    `json:"updateId"`
	Timestamp time.Time `json:"timestamp"`
	Status    string    `json:"status"`
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
	OpenLogFile(stepID uint64) (LogLineWriteCloser, error)

	AddStatusUpdate(stepID uint64, timestamp time.Time, newStatus worker.Status) error
	AddArtifact(stepID uint64, artifactName string) (io.WriteCloser, error)

	ReadAllLogLines(stepID uint64) ([]LogLine, error)
}

type LogLineWriteCloser interface {
	io.Closer
	WriteLogLine(line string) error
}

func NewStore(fs FS) Store {
	return &store{
		fs: fs,
	}
}

type store struct {
	fs             FS
	lastStatusID   uint64
	lastArtifactID uint64

	statusMutex keyedMutex
}

func (s *store) AddArtifact(stepID uint64, artifactName string) (io.WriteCloser, error) {
	//artifactID := atomic.AddUint64(&s.lastArtifactID, 1)
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
