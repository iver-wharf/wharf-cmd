package kubernetes

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainerType_String(t *testing.T) {
	type testCase struct {
		name          string
		containerType ContainerType
		expectedStr   string
	}

	tests := []testCase{
		{
			name:          "init container type to string",
			containerType: ContainerTypeInit,
			expectedStr:   "init-container",
		},
		{
			name:          "app container type to string",
			containerType: ContainerTypeApp,
			expectedStr:   "app-container",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			typeStringName := fmt.Sprintf("%s", tc.containerType)
			assert.Equal(t, tc.expectedStr, typeStringName)
		})
	}
}
