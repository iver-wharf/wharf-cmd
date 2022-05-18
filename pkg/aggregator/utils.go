package aggregator

import (
	"context"

	"github.com/iver-wharf/wharf-api-client-go/v2/pkg/model/request"
	"github.com/iver-wharf/wharf-api-client-go/v2/pkg/wharfapi"
	"github.com/iver-wharf/wharf-cmd/internal/parallel"
	"github.com/iver-wharf/wharf-cmd/pkg/workerapi/workerclient"
	v1 "k8s.io/api/core/v1"
)

type portForwardedWorker struct {
	workerclient.Client
	portConn portConnection
	podName  string
}

func newPortForwardedWorker(a k8sAggr, podName string, buildID uint) (portForwardedWorker, error) {
	portConn, err := newPortForwarding(a, a.namespace, podName)
	if err != nil {
		return portForwardedWorker{}, err
	}
	worker, err := a.newWorkerClient(portConn.Local, buildID)
	if err != nil {
		portConn.Close()
		return portForwardedWorker{}, err
	}
	pfWorker := portForwardedWorker{
		Client:   worker,
		portConn: portConn,
		podName:  podName,
	}
	return pfWorker, nil
}

func (w portForwardedWorker) Close() {
	w.Client.Close()
	w.portConn.Close()
}

type k8sHandler struct {
	workerclient.Client
	wharfapi wharfapi.Client
	s        wharfapi.CreateBuildLogStream
	pod      v1.Pod

	prevStatus request.BuildStatus
}

func newK8sHandler(worker workerclient.Client, wharfapi wharfapi.Client, apiLogStream wharfapi.CreateBuildLogStream, pod v1.Pod) k8sHandler {
	return k8sHandler{
		Client:   worker,
		wharfapi: wharfapi,
		s:        apiLogStream,
		pod:      pod,
	}
}

func (h k8sHandler) handleRunningPod(ctx context.Context) error {
	if err := h.Client.Ping(ctx); err != nil {
		log.Debug().
			WithStringf("pod", "%s/%s", h.pod.Namespace, h.pod.Name).
			Message("Failed to ping worker pod. Assuming it's not running yet. Skipping.")
		return nil
	}
	pg := parallel.Group{}
	pg.AddFunc("logs", h.relayLogs)
	pg.AddFunc("status events", h.relayStatuses)
	pg.AddFunc("artifact events", h.relayArtifacts)
	return pg.RunCancelEarly(ctx)
}

func (h k8sHandler) relayLogs(ctx context.Context) error {
	err := relay[request.Log](logLineSource{ctx, h}, func(v request.Log) error {
		return h.s.Send(v)
	})
	if err != nil {
		log.Error().WithError(err).
			WithStringf("pod", "%s/%s", h.pod.Namespace, h.pod.Name).
			Message("Relay logs error.")
	}
	return err
}

func (h k8sHandler) relayStatuses(ctx context.Context) error {
	h.prevStatus = request.BuildScheduling
	err := relay[request.BuildStatus](statusSource{ctx, h}, func(newStatus request.BuildStatus) error {
		if s, ok := h.getStatusToSet(newStatus); ok {
			h.updateStatus(s)
		}
		return nil
	})

	h.ensureStatusCompletedOrFailed()

	if err != nil {
		log.Error().WithError(err).
			WithStringf("pod", "%s/%s", h.pod.Namespace, h.pod.Name).
			Message("Relay statuses error.")
	}
	return err
}

func (h k8sHandler) updateStatus(s request.BuildStatus) error {
	statusUpdate := request.LogOrStatusUpdate{Status: s}
	if _, err := h.wharfapi.UpdateBuildStatus(h.BuildID(), statusUpdate); err != nil {
		return err
	}
	log.Info().
		WithString("new", string(s)).
		WithString("previous", string(h.prevStatus)).
		Message("Updated build status.")
	h.prevStatus = s
	return nil
}

func (h k8sHandler) getStatusToSet(s request.BuildStatus) (request.BuildStatus, bool) {
	if s == request.BuildFailed || s == request.BuildCompleted {
		return s, true
	}
	if s == request.BuildRunning && h.prevStatus == request.BuildScheduling {
		return request.BuildRunning, true
	}
	return "", false
}

func (h k8sHandler) ensureStatusCompletedOrFailed() {
	if h.prevStatus != request.BuildCompleted && h.prevStatus != request.BuildFailed {
		h.updateStatus(request.BuildFailed)
	}
}

func (h k8sHandler) relayArtifacts(ctx context.Context) error {
	err := relay[*workerclient.ArtifactEvent](artifactSource{ctx, h}, func(v *workerclient.ArtifactEvent) error {
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

func closeLogStream(s wharfapi.CreateBuildLogStream) {
	resp, err := s.CloseAndRecv()
	if err != nil {
		log.Warn().
			WithError(err).
			Message("Unexpected error when closing log writer stream to wharf-api.")
		return
	}
	log.Debug().
		WithUint("inserted", resp.LogsInserted).
		Message("Inserted logs into wharf-api.")
}
