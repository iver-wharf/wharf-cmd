package resultstore

import (
	"time"

	"github.com/iver-wharf/wharf-cmd/pkg/worker"
)

func (s *store) AddStatusUpdate(stepID uint64, timestmap time.Time, newStatus worker.Status) error {
	//statusID := atomic.AddUint64(&s.lastStatusID, 1)
	return nil
}
