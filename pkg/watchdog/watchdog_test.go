package watchdog

import (
	"testing"
	"time"

	"github.com/iver-wharf/wharf-api-client-go/v2/pkg/model/response"
	"github.com/iver-wharf/wharf-cmd/pkg/provisioner"
	"github.com/stretchr/testify/assert"
	"gopkg.in/guregu/null.v4"
)

var (
	testTimeNow       = time.Date(2022, 3, 1, 10, 00, 00, 00, time.UTC)
	testTimeOld       = testTimeNow.Add(-2 * time.Minute)
	testTimeSafeAfter = testTimeNow.Add(-1 * time.Minute)
)

func TestGetBuildsToKill(t *testing.T) {
	var tests = []struct {
		name    string
		builds  []response.Build
		workers []provisioner.Worker
		want    []uint
	}{
		{
			name:    "empty",
			builds:  []response.Build{},
			workers: []provisioner.Worker{},
			want:    []uint{},
		},
		{
			name: "kills old",
			builds: []response.Build{
				{BuildID: 1, WorkerID: "abc", ScheduledOn: null.TimeFrom(testTimeOld)},
				{BuildID: 2, WorkerID: "def", ScheduledOn: null.TimeFrom(testTimeOld)},
				{BuildID: 3, WorkerID: "ghi", ScheduledOn: null.TimeFrom(testTimeOld)},
				{BuildID: 4, WorkerID: "jkl", ScheduledOn: null.TimeFrom(testTimeOld)},
			},
			workers: []provisioner.Worker{
				{ID: "abc"},
				{ID: "ghi"},
			},
			want: []uint{2, 4},
		},
		{
			name: "keeps new",
			builds: []response.Build{
				{BuildID: 1, WorkerID: "abc", ScheduledOn: null.TimeFrom(testTimeNow)},
				{BuildID: 2, WorkerID: "def", ScheduledOn: null.TimeFrom(testTimeNow)},
				{BuildID: 3, WorkerID: "ghi", ScheduledOn: null.TimeFrom(testTimeOld)},
				{BuildID: 4, WorkerID: "jkl", ScheduledOn: null.TimeFrom(testTimeNow)},
			},
			workers: []provisioner.Worker{
				{ID: "abc"},
			},
			want: []uint{3},
		},
		{
			name: "ignores empty worker IDs",
			builds: []response.Build{
				{BuildID: 1, WorkerID: "", ScheduledOn: null.TimeFrom(testTimeOld)},
				{BuildID: 2, WorkerID: "", ScheduledOn: null.TimeFrom(testTimeOld)},
			},
			workers: []provisioner.Worker{},
			want:    []uint{},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			builds := getBuildsToKill(tc.builds, tc.workers, testTimeSafeAfter)
			if len(builds) == 0 && len(tc.want) == 0 {
				return
			}
			got := getBuildIDs(builds)
			assert.Equal(t, tc.want, got)
		})
	}
}

func getBuildIDs(builds []response.Build) []uint {
	ids := make([]uint, len(builds))
	for i, b := range builds {
		ids[i] = b.BuildID
	}
	return ids
}

func TestGetWorkersToKill(t *testing.T) {
	var tests = []struct {
		name    string
		builds  []response.Build
		workers []provisioner.Worker
		want    []string
	}{
		{
			name:    "empty",
			builds:  []response.Build{},
			workers: []provisioner.Worker{},
			want:    []string{},
		},
		{
			name: "kills old",
			builds: []response.Build{
				{BuildID: 2, WorkerID: "def"},
				{BuildID: 3, WorkerID: "ghi"},
			},
			workers: []provisioner.Worker{
				{ID: "abc", CreatedAt: testTimeOld},
				{ID: "def", CreatedAt: testTimeOld},
				{ID: "ghi", CreatedAt: testTimeOld},
				{ID: "jkl", CreatedAt: testTimeOld},
			},
			want: []string{"abc", "jkl"},
		},
		{
			name: "keeps new",
			builds: []response.Build{
				{BuildID: 1, WorkerID: "abc"},
			},
			workers: []provisioner.Worker{
				{ID: "abc", CreatedAt: testTimeNow},
				{ID: "def", CreatedAt: testTimeNow},
				{ID: "ghi", CreatedAt: testTimeOld},
				{ID: "jkl", CreatedAt: testTimeNow},
			},
			want: []string{"ghi"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			workers := getWorkersToKill(tc.builds, tc.workers, testTimeSafeAfter)
			if len(workers) == 0 && len(tc.want) == 0 {
				return
			}
			got := getWorkerIDs(workers)
			assert.Equal(t, tc.want, got)
		})
	}
}

func getWorkerIDs(workers []provisioner.Worker) []string {
	ids := make([]string, len(workers))
	for i, w := range workers {
		ids[i] = w.ID
	}
	return ids
}
