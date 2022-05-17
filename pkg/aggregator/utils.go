package aggregator

import (
	"context"
	"runtime"

	"github.com/iver-wharf/wharf-api-client-go/v2/pkg/wharfapi"
	"github.com/iver-wharf/wharf-cmd/pkg/workerapi/workerclient"
)

type scopedLogStream struct {
	wharfapi.CreateBuildLogStream
}

func newScopedLogStream(ctx context.Context, wharfapi wharfapi.Client) (scopedLogStream, error) {
	s, err := wharfapi.CreateBuildLogStream(ctx)
	if err != nil {
		return scopedLogStream{}, err
	}
	logStream := scopedLogStream{s}
	runtime.SetFinalizer(logStream, func() {
		resp, err := logStream.CloseAndRecv()
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
	return logStream, nil
}

type scopedWorker struct {
	workerclient.Client
}

func newScopedWorker(a k8sAggr, podName string, buildID uint) (scopedWorker, error) {
	portConn, err := newPortForwarding(a, a.namespace, podName)
	if err != nil {
		return scopedWorker{}, err
	}
	runtime.SetFinalizer(portConn, func() {
		portConn.Close()
	})
	worker, err := a.newWorkerClient(portConn, buildID)
	if err != nil {
		return scopedWorker{}, err
	}
	runtime.SetFinalizer(worker, func() {
		worker.Close()
	})
	return scopedWorker{worker}, nil
}
