package provisioner

import (
	"fmt"
	"testing"

	"github.com/iver-wharf/wharf-cmd/pkg/worker"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestConvertPodToWorker_NilPodReturnsEmptyWorker(t *testing.T) {
	wantWorker := Worker{
		ID:     "",
		Name:   "",
		Status: 0,
	}
	gotWorker := convertPodToWorker(nil)

	assert.Equal(t, wantWorker.ID, gotWorker.ID)
	assert.Equal(t, wantWorker.Name, gotWorker.Name)
	assert.Equal(t, wantWorker.Status, gotWorker.Status)
}

func TestConvertPodToWorker_Success(t *testing.T) {
	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			UID:       "some-uid-420",
			Name:      "some-name",
			Namespace: "some-namespace",
		},
	}
	wantWorker := Worker{
		ID:     "some-uid-420",
		Name:   fmt.Sprintf("%s/%s", "some-namespace", "some-name"),
		Status: worker.StatusUnknown,
	}
	gotWorker := convertPodToWorker(&pod)

	assert.Equal(t, wantWorker.ID, gotWorker.ID)
	assert.Equal(t, wantWorker.Name, gotWorker.Name)
	assert.Equal(t, wantWorker.Status, gotWorker.Status)
}

func TestConvertPodsToWorkers_NilPodsReturnsEmptySlice(t *testing.T) {
	wantWorkers := []Worker{}
	gotWorkers := convertPodsToWorkers(nil)
	assert.ElementsMatch(t, wantWorkers, gotWorkers)
}

func TestConvertPodsToWorkers_Success(t *testing.T) {
	pods := []v1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{
				UID:       "first-uid-420",
				Name:      "first-name",
				Namespace: "first-namespace",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				UID:       "second-uid-420",
				Name:      "second-name",
				Namespace: "first-namespace",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				UID:       "third-uid-420",
				Name:      "third-name",
				Namespace: "second-namespace",
			},
		},
	}

	wantWorkers := []Worker{
		{
			ID:     "first-uid-420",
			Name:   fmt.Sprintf("%s/%s", "first-namespace", "first-name"),
			Status: worker.StatusUnknown,
		},
		{
			ID:     "second-uid-420",
			Name:   fmt.Sprintf("%s/%s", "first-namespace", "second-name"),
			Status: worker.StatusUnknown,
		},
		{
			ID:     "third-uid-420",
			Name:   fmt.Sprintf("%s/%s", "second-namespace", "third-name"),
			Status: worker.StatusUnknown,
		},
	}
	gotWorkers := convertPodsToWorkers(pods)

	assert.ElementsMatch(t, wantWorkers, gotWorkers)
}

func TestConvertPodListToWorkerList_NilPodListReturnsEmptyWorkerList(t *testing.T) {
	wantWorkerList := WorkerList{}
	gotWorkerList := convertPodListToWorkerList(nil)
	assert.ElementsMatch(t, wantWorkerList.Items, gotWorkerList.Items)
	assert.Equal(t, wantWorkerList.Count, gotWorkerList.Count)
}

func TestConvertPodListToWorkerList_Success(t *testing.T) {
	podList := v1.PodList{
		Items: []v1.Pod{
			{
				ObjectMeta: metav1.ObjectMeta{
					UID:       "first-uid-420",
					Name:      "first-name",
					Namespace: "first-namespace",
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					UID:       "second-uid-420",
					Name:      "second-name",
					Namespace: "first-namespace",
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					UID:       "third-uid-420",
					Name:      "third-name",
					Namespace: "second-namespace",
				},
			},
		},
	}

	wantWorkerList := WorkerList{
		Items: []Worker{
			{
				ID:     "first-uid-420",
				Name:   fmt.Sprintf("%s/%s", "first-namespace", "first-name"),
				Status: worker.StatusUnknown,
			},
			{
				ID:     "second-uid-420",
				Name:   fmt.Sprintf("%s/%s", "first-namespace", "second-name"),
				Status: worker.StatusUnknown,
			},
			{
				ID:     "third-uid-420",
				Name:   fmt.Sprintf("%s/%s", "second-namespace", "third-name"),
				Status: worker.StatusUnknown,
			},
		},
	}
	wantWorkerList.Count = len(wantWorkerList.Items)

	gotWorkerList := convertPodListToWorkerList(&podList)

	assert.ElementsMatch(t, wantWorkerList.Items, gotWorkerList.Items)
	assert.Equal(t, wantWorkerList.Count, gotWorkerList.Count)
}
