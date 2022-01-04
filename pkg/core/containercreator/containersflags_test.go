package containercreator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidInitContainer(t *testing.T) {
	type testCase struct {
		name           string
		initContainer  InitContainerType
		expectedResult bool
	}

	tests := []testCase{
		{
			name:           "Valid init container type",
			initContainer:  Git,
			expectedResult: true,
		},
		{
			name:           "Invalid init container type",
			initContainer:  InitContainerType(0x8),
			expectedResult: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.initContainer.IsValid()
			assert.Equal(t, tc.expectedResult, result)
		})
	}
}

func TestInitContainerToPodContainerFlag(t *testing.T) {
	containerFlag := Git.toPodContainersFlag()
	assert.Equal(t, PodContainersFlags(0x1), containerFlag)
}

func TestIsValidContainer(t *testing.T) {
	type testCase struct {
		name           string
		container      ContainerType
		expectedResult bool
	}

	tests := []testCase{
		{
			name:           "Valid container type",
			container:      Kaniko,
			expectedResult: true,
		},
		{
			name:           "Invalid container type",
			container:      ContainerType(0x80),
			expectedResult: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.container.IsValid()
			assert.Equal(t, tc.expectedResult, result)
		})
	}
}

func TestContainerToPodContainerFlag(t *testing.T) {
	containerFlag := KubeApply.toPodContainersFlag()
	assert.Equal(t, PodContainersFlags(0x1000000000), containerFlag)
}

func TestAddContainer(t *testing.T) {
	type testCase struct {
		name         string
		flags        PodContainersFlags
		newContainer ContainerType
		expected     PodContainersFlags
	}

	tests := []testCase{
		{
			name:         "Add valid container type",
			flags:        PodContainersFlags(0x1),
			newContainer: Docker,
			expected:     PodContainersFlags(0x400000001),
		},
		{
			name:         "Add invalid container type",
			flags:        PodContainersFlags(0x1),
			newContainer: ContainerType(0x800000),
			expected:     PodContainersFlags(0x80000000000001),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.flags.AddContainer(tc.newContainer)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestAddInitContainer(t *testing.T) {
	type testCase struct {
		name         string
		flags        PodContainersFlags
		newContainer InitContainerType
		expected     PodContainersFlags
	}

	tests := []testCase{
		{
			name:         "Add valid init container type",
			flags:        PodContainersFlags(0x400000000),
			newContainer: Git,
			expected:     PodContainersFlags(0x400000001),
		},
		{
			name:         "Add invalid container type",
			flags:        PodContainersFlags(0x400000000),
			newContainer: InitContainerType(0x8),
			expected:     PodContainersFlags(0x400000008),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result:= tc.flags.AddInitContainer(tc.newContainer)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestGetContainers(t *testing.T) {
	flags := PodContainersFlags(0x600000001)
	containers := flags.GetContainers()
	assert.Equal(t, ContainerType(0x6), containers)
}

func TestGetInitContainers(t *testing.T) {
	flags := PodContainersFlags(0x600000001)
	containers := flags.GetInitContainers()
	assert.Equal(t, InitContainerType(0x1), containers)
}

func TestBitsCounter(t *testing.T) {
	type testCase struct {
		name     string
		flags    uint
		maxValue uint
		count    uint
	}
	tests := []testCase{
		{
			name:     "Valid flags",
			flags:    uint(Kaniko | Helm | KubeApply),
			maxValue: uint(AllAvailableContainers),
			count:    3,
		},
		{
			name:     "More flags than max value",
			flags:    uint(Kaniko|Helm|KubeApply) | 0x20,
			maxValue: uint(AllAvailableContainers),
			count:    3,
		},
		{
			name:     "No flags below max value",
			flags:    0x20,
			maxValue: uint(AllAvailableContainers),
			count:    0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := bitsCounter(tc.flags, tc.maxValue)
			assert.Equal(t, tc.count, result)
		})
	}
}
