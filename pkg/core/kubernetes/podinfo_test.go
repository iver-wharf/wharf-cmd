package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestMapToPodInfo(t *testing.T) {
	namespace := "default"
	name := "super-pod"
	uid := "1234567890-poiuytr"

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			UID:       types.UID(uid),
		},
	}

	podInfo := mapToPodInfo(pod)
	assert.Equal(t, namespace, podInfo.namespace)
	assert.Equal(t, name, podInfo.name)
	assert.Equal(t, uid, podInfo.podUID)
}

func TestNilMapToPodInfo(t *testing.T) {
	podInfo := mapToPodInfo(nil)
	assert.Equal(t, "", podInfo.namespace)
	assert.Equal(t, "", podInfo.name)
	assert.Equal(t, "", podInfo.podUID)
}
