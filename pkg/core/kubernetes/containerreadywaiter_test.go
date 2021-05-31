package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type containerReadyWaiterSuite struct {
	suite.Suite
	fakeClientset    *fake.Clientset
	fakePodInterface corev1.PodInterface
	sut              *containerWaiter
}

func TestContainerReadyWaiter(t *testing.T) {
	suite.Run(t, new(containerReadyWaiterSuite))
}

func (suite *containerReadyWaiterSuite) SetupSuite() {
	suite.fakeClientset = fake.NewSimpleClientset()

	p := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "my-pod"},
		Spec: v1.PodSpec{
			InitContainers: []v1.Container{{Name: "git-cloner"}, {Name: "prepare-sth"}},
			Containers:     []v1.Container{{Name: "do job"}},
		},
		Status: v1.PodStatus{
			InitContainerStatuses: []v1.ContainerStatus{
				{
					Name:         "git-cloner",
					Ready:        true,
					RestartCount: 0,
				},
				{
					Name:         "prepare-sth",
					Ready:        true,
					RestartCount: 0,
				},
			},
			ContainerStatuses: []v1.ContainerStatus{
				{
					Name:         "do job",
					Ready:        false,
					RestartCount: 2,
				},
			},
		},
	}

	pod, err := suite.fakeClientset.CoreV1().Pods("test-ns").Create(p)
	require.Nil(suite.T(), err)
	require.NotNil(suite.T(), pod)

	suite.fakePodInterface = suite.fakeClientset.CoreV1().Pods("test-ns")
}

func (suite *containerReadyWaiterSuite) SetupTest() {
	var err error
	rw, err := NewContainerReadyWaiter(suite.fakePodInterface, "my-pod")
	suite.sut = rw.(*containerWaiter)
	require.Nil(suite.T(), err)
}

func (suite *containerReadyWaiterSuite) TestFirstRequestAnyRemaining() {
	assert.False(suite.T(), suite.sut.containersFetched)

	more := suite.sut.AnyRemaining()
	assert.True(suite.T(), more)
}

func (suite *containerReadyWaiterSuite) TestAnyRemainingWhenContainersFetched() {
	_, err := suite.sut.WaitNext()
	require.Nil(suite.T(), err)

	more := suite.sut.AnyRemaining()
	assert.True(suite.T(), more)
}

func (suite *containerReadyWaiterSuite) TestFirstRequestSetContainersFetchToTrueWaitNext() {
	container, err := suite.sut.WaitNext()

	assert.True(suite.T(), suite.sut.containersFetched)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "git-cloner", container.Name)
	assert.Equal(suite.T(), 2, len(suite.sut.containers))

	container, err = suite.sut.WaitNext()
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), "prepare-sth", container.Name)
	assert.Equal(suite.T(), 3, len(suite.sut.containers))
}

func (suite *containerReadyWaiterSuite) TestEmptyStatusesPastToNextReadyContainer() {
	var containerStatuses []v1.ContainerStatus
	emptyContainer := AwaitedContainer{
		Namespace: "",
		PodUID:    "",
		PodName:   "",
		Name:      "",
		Type:      false,
		Restarts:  0,
		isReady:   false,
	}

	container, ok := suite.sut.nextContainerInExpectedState(containerStatuses, ContainerTypeInit)
	assert.False(suite.T(), ok)
	assert.Equal(suite.T(), emptyContainer, container)
}

func (suite *containerReadyWaiterSuite) TestNextReadyContainerProperData() {
	containerStatuses := []v1.ContainerStatus{
		{
			Name:                 "test 1",
			State:                v1.ContainerState{},
			LastTerminationState: v1.ContainerState{},
			Ready:                false,
			RestartCount:         0,
			Image:                "",
			ImageID:              "",
			ContainerID:          "",
		},
		{
			Name:                 "test 2",
			State:                v1.ContainerState{},
			LastTerminationState: v1.ContainerState{},
			Ready:                true,
			RestartCount:         2,
			Image:                "",
			ImageID:              "",
			ContainerID:          "",
		},
	}

	container, ok := suite.sut.nextContainerInExpectedState(containerStatuses, ContainerTypeInit)
	assert.True(suite.T(), ok)

	assert.True(suite.T(), container.isReady)
	assert.Equal(suite.T(), "test 2", container.Name)
	assert.Equal(suite.T(), ContainerTypeInit, container.Type)
	assert.Equal(suite.T(), 2, container.Restarts)
	assert.Equal(suite.T(), "", container.PodUID)
	assert.Equal(suite.T(), "my-pod", container.PodName)
	assert.Equal(suite.T(), "test-ns", container.Namespace)

	assert.Equal(suite.T(), 2, len(suite.sut.containers))
}
