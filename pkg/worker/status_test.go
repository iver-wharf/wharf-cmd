package worker

import (
	"encoding/json"
	"testing"

	"github.com/iver-wharf/wharf-cmd/pkg/worker/workermodel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatusMarshalJSON(t *testing.T) {
	got, err := json.Marshal(struct {
		MyStatus workermodel.Status
	}{
		MyStatus: workermodel.StatusSuccess,
	})
	require.NoError(t, err)
	want := `{"MyStatus": "Success"}`
	assert.JSONEq(t, want, string(got))
}

func TestStatusUnmarshalJSON(t *testing.T) {
	var myStruct struct {
		MyStatus workermodel.Status
	}
	b := []byte(`{"MyStatus": "Success"}`)
	err := json.Unmarshal(b, &myStruct)
	require.NoError(t, err)
	assert.Equal(t, workermodel.StatusSuccess, myStruct.MyStatus)
}

func TestStatusPtrUnmarshalJSON(t *testing.T) {
	var myStruct struct {
		MyStatus *workermodel.Status
	}
	b := []byte(`{"MyStatus": "Success"}`)
	err := json.Unmarshal(b, &myStruct)
	require.NoError(t, err)
	assert.Equal(t, workermodel.StatusSuccess, *myStruct.MyStatus)
}
