package httptests

import "github.com/iver-wharf/wharf-cmd/pkg/workerapi/workerserver/model/response"

type mockArtifactLister struct{}

const (
	unusedArtifactID   = uint(351)
	validArtifactID1   = uint(34)
	validArtifactID2   = uint(35)
	validArtifactName1 = "valid-artifact-1"
	validArtifactName2 = "valid-artifact-2"
)

func (a *mockArtifactLister) ListArtifacts() []response.Artifact {
	return []response.Artifact{
		{
			ArtifactID: validArtifactID1,
			StepID:     1,
			Name:       validArtifactName1,
		},
		{
			ArtifactID: validArtifactID2,
			StepID:     1,
			Name:       validArtifactName2,
		},
	}
}
