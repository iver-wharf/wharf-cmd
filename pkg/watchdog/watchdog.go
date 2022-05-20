package watchdog

import (
	"fmt"
	"time"

	"github.com/iver-wharf/wharf-api-client-go/v2/pkg/model/request"
	"github.com/iver-wharf/wharf-api-client-go/v2/pkg/model/response"
	"github.com/iver-wharf/wharf-api-client-go/v2/pkg/wharfapi"
	"github.com/iver-wharf/wharf-cmd/pkg/config"
	"github.com/iver-wharf/wharf-cmd/pkg/provisioner"
	"github.com/iver-wharf/wharf-cmd/pkg/provisioner/provisionerclient"
	"github.com/iver-wharf/wharf-cmd/pkg/worker/workermodel"
	"github.com/iver-wharf/wharf-core/v2/pkg/logger"
)

var log = logger.NewScoped("WATCHDOG")

const interval = 5 * time.Minute
const safeAfterDuration = -time.Minute

// Watch will start the watchdog and wait.
func Watch(config config.WatchdogConfig) error {
	wd := watchdog{
		wharfapi: wharfapi.Client{
			APIURL: config.WharfAPIURL,
		},
		prov: provisionerclient.Client{
			APIURL: config.ProvisionerURL,
		},
	}
	wd.startTicker()
	return wd.listenForTicks()
}

type watchdog struct {
	wharfapi wharfapi.Client
	prov     provisionerclient.Client
	ticker   *time.Ticker
}

func (wd *watchdog) startTicker() {
	wd.ticker = time.NewTicker(interval)
}

func (wd *watchdog) listenForTicks() error {
	res, err := wd.performCheck(time.Now())
	if err != nil {
		log.Error().
			WithError(err).
			WithDuration("interval", interval).
			Message("Failed to perform initial check. Need at least one successful check to start. Stopping ticker.")
		wd.ticker.Stop()
		return err
	}
	log.Info().WithDuration("interval", interval).
		WithInt("killedBuilds", res.killedBuilds).
		WithInt("killedWorkers", res.killedWorkers).
		Message("Done with initial check. Waiting for next interval tick.")
	for now := range wd.ticker.C {
		res, err := wd.performCheck(now)
		if err != nil {
			log.Warn().
				WithError(err).
				WithDuration("interval", interval).
				Message("Failed to perform check. Will try again in next tick.")
			continue
		}
		log.Info().WithDuration("interval", interval).
			WithInt("killedBuilds", res.killedBuilds).
			WithInt("killedWorkers", res.killedWorkers).
			Message("Done with check. Waiting for next interval tick.")
	}
	return nil
}

type checkResult struct {
	killedBuilds  int
	killedWorkers int
}

func (wd *watchdog) performCheck(now time.Time) (checkResult, error) {
	builds, err := wd.getRunningBuilds()
	if err != nil {
		return checkResult{}, fmt.Errorf("get running builds from wharf-api: %w", err)
	}
	workers, err := wd.prov.ListWorkers()
	if err != nil {
		return checkResult{}, fmt.Errorf("get running workers from wharf-cmd-provisioner: %w", err)
	}
	if len(builds) == 0 && len(workers) == 0 {
		log.Debug().
			Message("Found no running builds nor workers.")
		return checkResult{}, nil
	}

	log.Debug().
		WithInt("builds", len(builds)).
		WithInt("workers", len(workers)).
		Message("Found running builds and workers.")

	safeAfter := now.Add(safeAfterDuration)
	buildsToKill := getBuildsToKill(builds, workers, safeAfter)
	workersToKill := getWorkersToKill(builds, workers, safeAfter)

	if len(buildsToKill) == 0 && len(workersToKill) == 0 {
		log.Debug().Message("No workers nor builds to kill. Happy day!")
		return checkResult{}, nil
	}
	if err := wd.killBuilds(buildsToKill); err != nil {
		return checkResult{}, fmt.Errorf("kill builds: %w", err)
	}
	if err := wd.killWorkers(workersToKill); err != nil {
		return checkResult{}, fmt.Errorf("kill workers: %w", err)
	}
	return checkResult{
		killedBuilds:  len(buildsToKill),
		killedWorkers: len(workersToKill),
	}, nil
}

func (wd *watchdog) killBuilds(buildsToKill []response.Build) error {
	if len(buildsToKill) == 0 {
		return nil
	}
	log.Debug().WithInt("builds", len(buildsToKill)).
		Message("Killing builds.")
	for _, b := range buildsToKill {
		updated, err := wd.wharfapi.UpdateBuildStatus(b.BuildID, request.LogOrStatusUpdate{
			Status: request.BuildFailed,
		})
		if err != nil {
			return fmt.Errorf("update status on build %d to %s: %w",
				b.BuildID, request.BuildFailed, err)
		}
		log.Info().
			WithUint("buildId", b.BuildID).
			WithString("oldStatus", string(b.Status)).
			WithString("newStatus", string(updated.Status)).
			WithTime("scheduled", b.ScheduledOn.Time).
			Message("Killed stray build.")
	}
	return nil
}

func (wd *watchdog) killWorkers(workersToKill []provisioner.Worker) error {
	if len(workersToKill) == 0 {
		return nil
	}
	log.Debug().WithInt("workers", len(workersToKill)).
		Message("Killing workers.")
	for _, w := range workersToKill {
		if err := wd.prov.DeleteWorker(w.WorkerID); err != nil {
			return fmt.Errorf("kill worker by ID %q: %w", w.WorkerID, err)
		}
	}
	return nil
}

func getBuildsToKill(builds []response.Build, workers []provisioner.Worker, safeAfter time.Time) []response.Build {
	workersMap := mapRunningWorkersOnID(workers)
	var toKill []response.Build
	for _, b := range builds {
		reason := getReasonToNotKillBuild(b, workersMap, safeAfter)
		if reason != "" {
			log.Debug().
				WithString("reason", reason).
				WithUint("buildId", b.BuildID).
				Messagef("Skip killing build.")
			continue
		}
		toKill = append(toKill, b)
	}
	return toKill
}

func getReasonToNotKillBuild(b response.Build, workersMap map[string]provisioner.Worker, safeAfter time.Time) string {
	if b.WorkerID == "" {
		return "build.workerId is empty"
	}
	if _, ok := workersMap[b.WorkerID]; ok {
		return "the build's worker is still running"
	}
	if !b.ScheduledOn.Valid {
		return "build.scheduledOn is not valid"
	}
	if b.ScheduledOn.Time.After(safeAfter) {
		return "build.scheduledOn is not old enough"
	}
	return ""
}

func getWorkersToKill(builds []response.Build, workers []provisioner.Worker, safeAfter time.Time) []provisioner.Worker {
	buildsMap := mapBuildsOnWorkerID(builds)
	var toKill []provisioner.Worker
	for _, w := range workers {
		reason := getReasonNotToKillWorker(w, buildsMap, safeAfter)
		if reason != "" {
			log.Debug().
				WithString("reason", reason).
				WithString("workerId", w.WorkerID).
				Messagef("Skip killing worker.")
			continue
		}
		toKill = append(toKill, w)
	}
	return toKill
}

func getReasonNotToKillWorker(w provisioner.Worker, buildsMap map[string]response.Build, safeAfter time.Time) string {
	if _, ok := buildsMap[w.WorkerID]; ok {
		return "the worker's build is still running"
	}
	if w.CreatedAt.After(safeAfter) {
		return "worker.createdAt is not old enough"
	}
	return ""
}

func mapBuildsOnWorkerID(builds []response.Build) map[string]response.Build {
	m := make(map[string]response.Build, len(builds))
	for _, b := range builds {
		m[b.WorkerID] = b
	}
	return m
}

func mapRunningWorkersOnID(workers []provisioner.Worker) map[string]provisioner.Worker {
	m := make(map[string]provisioner.Worker, len(workers))
	for _, w := range workers {
		switch w.Status {
		case workermodel.StatusInitializing,
			workermodel.StatusScheduling,
			workermodel.StatusRunning,
			workermodel.StatusNone:
			m[w.WorkerID] = w
		}
	}
	return m
}

func (wd *watchdog) getRunningBuilds() ([]response.Build, error) {
	limit := 0
	page, err := wd.wharfapi.GetBuildList(wharfapi.BuildSearch{
		Limit: &limit,
		StatusID: []int{
			int(wharfapi.BuildScheduling),
			int(wharfapi.BuildRunning),
		},
	})
	if err != nil {
		return nil, err
	}
	return page.List, nil
}
