package aggregator

import (
	"bufio"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/iver-wharf/wharf-api-client-go/v2/pkg/model/request"
	v1 "k8s.io/api/core/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type k8sRawLogSource struct {
	ctx context.Context

	buildID uint
	pod     v1.Pod

	pods corev1.PodInterface
}

func (s k8sRawLogSource) pushInto(dst chan<- request.Log) error {
	dst <- s.stringToLog(fmt.Sprintf("The %s pod failed to start.", s.pod.Name))
	dst <- s.stringToLog(fmt.Sprintf("Logs from %s:\n", s.pod.Name))
	req := s.pods.GetLogs(s.pod.Name, &v1.PodLogOptions{})
	readCloser, err := req.Stream(s.ctx)
	if err != nil {
		return err
	}
	defer readCloser.Close()
	scanner := bufio.NewScanner(readCloser)
	for scanner.Scan() {
		txt := scanner.Text()
		log.Debug().Message(txt)
		idx := strings.LastIndexByte(txt, '\r')
		if idx != -1 {
			txt = txt[idx+1:]
		}
		dst <- s.stringToLog(txt)
	}
	return nil
}

func (s k8sRawLogSource) stringToLog(str string) request.Log {
	return request.Log{
		BuildID: s.buildID,
		// WorkerLogID:  0,
		// WorkerStepID: 0,
		Timestamp: time.Now(),
		Message:   str,
	}
}
