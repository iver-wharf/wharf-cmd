package watchdog

import (
	"fmt"
	"time"

	"github.com/iver-wharf/wharf-api-client-go/v2/pkg/model/response"
	"github.com/iver-wharf/wharf-api-client-go/v2/pkg/wharfapi"
	"github.com/iver-wharf/wharf-cmd/pkg/provisioner"
	"github.com/iver-wharf/wharf-cmd/pkg/provisionerclient"
	"github.com/iver-wharf/wharf-cmd/pkg/worker"
	"github.com/iver-wharf/wharf-core/pkg/logger"
)

var log = logger.NewScoped("WATCHDOG")

// Watch will start the watchdog and wait.
func Watch() error {
	var wd watchdog
	wd.startTicker()
	return nil
}

type watchdog struct {
	wharfapi wharfapi.Client
	prov     provisionerclient.Client
	ticker   *time.Ticker
}

func (wd *watchdog) startTicker() {
	wd.ticker = time.NewTicker(5 * time.Minute)
}

func (wd *watchdog) listenForTicks() {
	for range wd.ticker.C {
		if err := wd.performCheck(); err != nil {
			log.Warn().
				WithError(err).
				WithDuration("interval", 5*time.Minute).
				Message("Failed to perform check. Will try again in next tick.")
		}
	}
}

func (wd *watchdog) performCheck() error {
	builds, err := wd.getRunningBuilds()
	if err != nil {
		return fmt.Errorf("get running builds from wharf-api: %w", err)
	}
	workers, err := wd.getRunningWorkers()
	if err != nil {
		return fmt.Errorf("get running workers from wharf-cmd-provisioner: %w", err)
	}
	log.Debug().
		WithInt("builds", len(builds)).
		WithInt("workers", len(workers)).
		Message("Found running builds and workers.")
	// TODO: diff the builds x workers
	return nil
}

func getBuildsToKill(builds []response.Build, workers []provisioner.Worker, safeAfter time.Time) []uint {
	workersMap := mapWorkersOnID(workers)
	var toKill []uint
	for _, b := range builds {
		if b.WorkerID == "" {
			continue
		}
		if _, ok := workersMap[b.WorkerID]; ok {
			continue
		}
		if !b.ScheduledOn.Valid {
			continue
		}
		if b.ScheduledOn.Time.After(safeAfter) {
			continue
		}
		toKill = append(toKill, b.BuildID)
	}
	return toKill
}

func getWorkersToKill(builds []response.Build, workers []provisioner.Worker, safeAfter time.Time) []string {
	buildsMap := mapBuildsOnWorkerID(builds)
	var toKill []string
	for _, w := range workers {
		if _, ok := buildsMap[w.ID]; ok {
			continue
		}
		if w.CreatedAt.After(safeAfter) {
			continue
		}
		toKill = append(toKill, w.ID)
	}
	return toKill
}

func mapBuildsOnWorkerID(builds []response.Build) map[string]response.Build {
	m := make(map[string]response.Build, len(builds))
	for _, b := range builds {
		m[b.WorkerID] = b
	}
	return m
}

func mapWorkersOnID(workers []provisioner.Worker) map[string]provisioner.Worker {
	m := make(map[string]provisioner.Worker, len(workers))
	for _, w := range workers {
		m[w.ID] = w
	}
	return m
}

func (wd *watchdog) getRunningBuilds() ([]response.Build, error) {
	limit := 0
	page, err := wd.wharfapi.GetBuildList(wharfapi.BuildSearch{
		Limit: &limit,
		StatusID: []int{
			int(wharfapi.BuildRunning),
			int(wharfapi.BuildScheduling),
		},
	})
	if err != nil {
		return nil, err
	}
	return page.List, nil
}

func (wd *watchdog) getRunningWorkers() ([]provisioner.Worker, error) {
	allWorkers, err := wd.prov.ListWorkers()
	if err != nil {
		return nil, err
	}
	var runningWorkers []provisioner.Worker
	for _, w := range allWorkers {
		switch w.Status {
		case worker.StatusInitializing,
			worker.StatusScheduling,
			worker.StatusRunning:
			allWorkers = append(allWorkers, w)
		}
	}
	return runningWorkers, nil
}
