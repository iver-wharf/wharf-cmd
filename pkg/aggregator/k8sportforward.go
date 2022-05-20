package aggregator

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/iver-wharf/wharf-cmd/pkg/worker/workerclient"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

type portForwardedWorker struct {
	workerclient.Client
	portConn portConnection
	podName  string
}

func newPortForwardedWorker(a k8sAggr, podName string, buildID uint) (portForwardedWorker, error) {
	portConn, err := newPortConnection(a, a.namespace, podName)
	if err != nil {
		return portForwardedWorker{}, err
	}

	worker, err := workerclient.New(fmt.Sprintf("http://localhost:%d", portConn.Local), workerclient.Options{
		// Skipping security because we've already authenticated with Kubernetes
		// and are communicating through a secured port-forwarding tunnel.
		// Don't need to add TLS on top of TLS.
		InsecureSkipVerify: true,
		BuildID:            buildID,
	})

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

func (w portForwardedWorker) Close() error {
	w.Client.Close()
	w.portConn.Close()
	return nil
}

type portConnection struct {
	portforward.ForwardedPort
	stopCh chan struct{}
}

func newPortConnection(a k8sAggr, namespace, podName string) (portConnection, error) {
	portForwardURL, err := newPortForwardURL(a.restConfig.Host, namespace, podName)
	if err != nil {
		return portConnection{}, err
	}

	dialer := spdy.NewDialer(a.upgrader, a.httpClient, http.MethodGet, portForwardURL)
	stopCh, readyCh := make(chan struct{}, 1), make(chan struct{}, 1)
	forwarder, err := portforward.New(dialer,
		// From random unused local port (port 0) to the worker HTTP API port.
		[]string{fmt.Sprintf("0:%d", a.aggrConfig.WorkerAPIExternalPort)},
		stopCh, readyCh, nil, nil)
	if err != nil {
		return portConnection{}, err
	}

	var forwarderErr error
	go func() {
		if forwarderErr = forwarder.ForwardPorts(); forwarderErr != nil {
			close(stopCh)
		}
	}()

	select {
	case <-readyCh:
	case <-stopCh:
	}
	if forwarderErr != nil {
		return portConnection{}, forwarderErr
	}

	ports, err := forwarder.GetPorts()
	if err != nil {
		log.Error().WithError(err).Message("Error getting ports.")
		close(stopCh)
		return portConnection{}, err
	}
	port := ports[0]

	log.Debug().
		WithStringf("pod", "%s/%s", a.namespace, podName).
		WithUint("local", uint(port.Local)).
		WithUint("remote", uint(port.Remote)).
		Message("Connected to worker. Port-forwarding from pod.")

	return portConnection{
		ForwardedPort: port,
		stopCh:        stopCh,
	}, nil
}

func (pc portConnection) Close() error {
	close(pc.stopCh)
	return nil
}

func newPortForwardURL(apiURL, namespace, podName string) (*url.URL, error) {
	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward",
		url.PathEscape(namespace), url.PathEscape(podName))

	portForwardURL, err := url.Parse(apiURL)
	if err != nil {
		return nil, fmt.Errorf("parse URL from kubeconfig: %w", err)
	}
	portForwardURL.Path += path
	return portForwardURL, nil
}
