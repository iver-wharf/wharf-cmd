package worker

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatusMarshalJSON(t *testing.T) {
	got, err := json.Marshal(struct {
		MyStatus Status
	}{
		MyStatus: StatusSuccess,
	})
	require.NoError(t, err)
	want := `{"MyStatus": "Success"}`
	assert.JSONEq(t, want, string(got))
}

func TestStatusUnmarshalJSON(t *testing.T) {
	var myStruct struct {
		MyStatus Status
	}
	b := []byte(`{"MyStatus": "Success"}`)
	err := json.Unmarshal(b, &myStruct)
	require.NoError(t, err)
	assert.Equal(t, StatusSuccess, myStruct.MyStatus)
}

func TestStatusPtrUnmarshalJSON(t *testing.T) {
	var myStruct struct {
		MyStatus *Status
	}
	b := []byte(`{"MyStatus": "Success"}`)
	err := json.Unmarshal(b, &myStruct)
	require.NoError(t, err)
	assert.Equal(t, StatusSuccess, *myStruct.MyStatus)
}
