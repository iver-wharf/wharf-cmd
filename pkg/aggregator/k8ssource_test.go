package aggregator_test

import (
	"fmt"
	"testing"

	"github.com/iver-wharf/wharf-cmd/pkg/aggregator"
	"github.com/stretchr/testify/assert"
)

type mockSource struct {
	// for example of configurable source
	startNumber int
}

var _ aggregator.Source[string] = &mockSource{}

const mockSourceFinalLength = 5

func (s mockSource) PushInto(dst chan<- string) error {
	for i := s.startNumber; i < s.startNumber+mockSourceFinalLength; i++ {
		dst <- fmt.Sprintf("%d", i)
	}
	return nil
}

func TestSource(t *testing.T) {
	testCases := []struct {
		name       string
		source     mockSource
		wantValues []string
	}{
		{
			name:       "0 to 4",
			source:     mockSource{},
			wantValues: []string{"0", "1", "2", "3", "4"},
		},
		{
			name:       "5 to 9",
			source:     mockSource{startNumber: 5},
			wantValues: []string{"5", "6", "7", "8", "9"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dst := make(chan string)
			go func() {
				tc.source.PushInto(dst)
				close(dst)
			}()
			var result []string
			for v, ok := <-dst; ok; v, ok = <-dst {
				result = append(result, v)
			}
			assert.Len(t, result, mockSourceFinalLength)
			assert.EqualValues(t, tc.wantValues, result)
		})
	}
}
