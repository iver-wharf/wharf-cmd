package aggregator

import (
	"context"
	"runtime"

	"github.com/iver-wharf/wharf-api-client-go/v2/pkg/model/request"
	"github.com/iver-wharf/wharf-api-client-go/v2/pkg/wharfapi"
	"github.com/iver-wharf/wharf-cmd/internal/parallel"
	"github.com/iver-wharf/wharf-cmd/pkg/workerapi/workerclient"
	v1 "k8s.io/api/core/v1"
)

type scopedLogStream struct {
	wharfapi.CreateBuildLogStream
}

func newScopedLogStream(ctx context.Context, wharfapi wharfapi.Client) (*scopedLogStream, error) {
	s, err := wharfapi.CreateBuildLogStream(ctx)
	if err != nil {
		return &scopedLogStream{}, err
	}
	scoped := &scopedLogStream{s}
	runtime.SetFinalizer(scoped, func(s *scopedLogStream) {
		resp, err := scoped.CloseAndRecv()
		if err != nil {
			log.Warn().
				WithError(err).
				Message("Unexpected error when closing log writer stream to wharf-api.")
			return
		}
		log.Debug().
			WithUint("inserted", resp.LogsInserted).
			Message("Inserted logs into wharf-api.")
	})
	return scoped, nil
}

type scopedWorker struct {
	workerclient.Client
}

func newScopedWorker(a k8sAggr, podName string, buildID uint) (*scopedWorker, error) {
	portConn, err := newPortForwarding(a, a.namespace, podName)
	if err != nil {
		return &scopedWorker{}, err
	}
	worker, err := a.newWorkerClient(portConn.Local, buildID)
	if err != nil {
		return &scopedWorker{}, err
	}
	scoped := new(scopedWorker)
	scoped.Client = worker
	runtime.SetFinalizer(scoped, func(s *scopedWorker) {
		s.Close()
		portConn.Close()
	})
	return scoped, nil
}

func handleRunningPod(ctx context.Context, namespace string, wharfapi wharfapi.Client, apiLogStream wharfapi.CreateBuildLogStream, worker workerclient.Client, pod v1.Pod) error {
	if err := worker.Ping(ctx); err != nil {
		log.Debug().
			WithStringf("pod", "%s/%s", namespace, pod.Name).
			Message("Failed to ping worker pod. Assuming it's not running yet. Skipping.")
		return nil
	}
	pg := parallel.Group{}
	pg.AddFunc("logs", func(pod v1.Pod) parallel.Func {
		return func(ctx context.Context) error {
			err := relay[request.Log](logLineSource{ctx, worker}, func(v request.Log) error {
				return apiLogStream.Send(v)
			})
			if err != nil {
				log.Error().WithError(err).
					WithStringf("pod", "%s/%s", pod.Namespace, pod.Name).
					Message("Relay logs error.")
			}
			return err
		}
	}(pod))
	pg.AddFunc("status events", func(pod v1.Pod) parallel.Func {
		return func(ctx context.Context) error {
			previousStatus := request.BuildScheduling
			updateStatus := func(newStatus request.BuildStatus) error {
				statusUpdate := request.LogOrStatusUpdate{Status: newStatus}
				if _, err := wharfapi.UpdateBuildStatus(worker.BuildID(), statusUpdate); err != nil {
					return err
				}
				log.Info().
					WithString("new", string(newStatus)).
					WithString("previous", string(previousStatus)).
					Message("Updated build status.")
				previousStatus = newStatus
				return nil
			}
			err := relay[request.BuildStatus](statusSource{ctx, worker}, func(reqNewStatus request.BuildStatus) error {
				if reqNewStatus == request.BuildFailed || reqNewStatus == request.BuildCompleted {
					updateStatus(reqNewStatus)
					// return updateStatus(reqNewStatus)
				}
				if reqNewStatus == request.BuildRunning && previousStatus == request.BuildScheduling {
					updateStatus(request.BuildRunning)
					// return updateStatus(request.BuildRunning)
				}
				return nil
			})
			if previousStatus != request.BuildCompleted && previousStatus != request.BuildFailed {
				updateStatus(request.BuildFailed)
			}
			if err != nil {
				log.Error().WithError(err).
					WithStringf("pod", "%s/%s", pod.Namespace, pod.Name).
					Message("Relay statuses error.")
			}
			return err
		}
	}(pod))
	pg.AddFunc("artifact events", func(pod v1.Pod) parallel.Func {
		return func(ctx context.Context) error {
			err := relay[*workerclient.ArtifactEvent](artifactSource{ctx, worker}, func(v *workerclient.ArtifactEvent) error {
				// No way to send to wharf DB through stream currently
				// so we're just logging it here.
				log.Debug().
					WithUint64("step", v.StepID).
					WithString("name", v.Name).
					WithUint64("id", v.ArtifactID).
					Message("Received artifact event.")
				return nil
			})
			return err
		}
	}(pod))
	return pg.RunCancelEarly(ctx)
}
