package aggregator

import (
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

	assert.Equal(t, wantWorker, gotWorker)
}

func TestConvertPodToWorker_BasicConversionIgnoringStatusInfo(t *testing.T) {
	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			UID:       "some-uid-420",
			Name:      "some-name",
			Namespace: "some-namespace",
		},
	}
	wantWorker := Worker{
		ID:   "some-uid-420",
		Name: "some-namespace/some-name",
	}
	gotWorker := convertPodToWorker(&pod)

	assert.Equal(t, wantWorker.ID, gotWorker.ID)
	assert.Equal(t, wantWorker.Name, gotWorker.Name)
}

func TestConvertPodToWorker_Status(t *testing.T) {
	testCases := []struct {
		name string
		pod  v1.Pod
		want worker.Status
	}{
		{
			name: "scheduling",
			pod: makeTestPod(v1.PodStatus{
				Conditions: []v1.PodCondition{{
					Type:   v1.PodScheduled,
					Status: v1.ConditionTrue,
				}}}),
			want: worker.StatusScheduling,
		},
		{
			name: "initializing",
			pod: makeTestPod(v1.PodStatus{
				InitContainerStatuses: []v1.ContainerStatus{{
					State: v1.ContainerState{
						Running: &v1.ContainerStateRunning{},
					},
				}}}),
			want: worker.StatusInitializing,
		},
		{
			name: "running",
			pod: makeTestPod(v1.PodStatus{
				ContainerStatuses: []v1.ContainerStatus{{
					State: v1.ContainerState{
						Running: &v1.ContainerStateRunning{},
					},
				}}}),
			want: worker.StatusRunning,
		},
		{
			name: "success",
			pod:  makeTestPodWithPhase(v1.PodSucceeded),
			want: worker.StatusSuccess,
		},
		{
			name: "failed",
			pod:  makeTestPodWithPhase(v1.PodFailed),
			want: worker.StatusFailed,
		},
		{
			name: "unknown_explicit",
			pod:  makeTestPodWithPhase(v1.PodUnknown),
			want: worker.StatusUnknown,
		},
		{
			name: "unknown_implicit",
			pod:  makeTestPod(v1.PodStatus{}),
			want: worker.StatusUnknown,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := convertPodToWorker(&tc.pod).Status
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestConvertPodsToWorkers_NilPodsReturnsEmptySlice(t *testing.T) {
	wantWorkers := []Worker{}
	gotWorkers := convertPodsToWorkers(nil)
	assert.ElementsMatch(t, wantWorkers, gotWorkers)
}

func TestConvertPodsToWorkers(t *testing.T) {
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
			Name:   "first-namespace/first-name",
			Status: worker.StatusUnknown,
		},
		{
			ID:     "second-uid-420",
			Name:   "first-namespace/second-name",
			Status: worker.StatusUnknown,
		},
		{
			ID:     "third-uid-420",
			Name:   "second-namespace/third-name",
			Status: worker.StatusUnknown,
		},
	}
	gotWorkers := convertPodsToWorkers(pods)

	assert.ElementsMatch(t, wantWorkers, gotWorkers)
}

func makeTestPod(status v1.PodStatus) v1.Pod {
	return v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			UID:       "some-uid-420",
			Name:      "some-name",
			Namespace: "some-namespace",
		},
		Status: status,
	}
}

func makeTestPodWithPhase(phase v1.PodPhase) v1.Pod {
	return v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			UID:       "some-uid-420",
			Name:      "some-name",
			Namespace: "some-namespace",
		},
		Status: v1.PodStatus{
			Phase: phase,
		},
	}
}
