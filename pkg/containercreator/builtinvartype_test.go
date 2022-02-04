package containercreator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToSafeBranchName(t *testing.T) {
	repoBranch := "feature/fix-error-on-submit"
	expectedSafeBranchName := "feature-fix-error-on-submit"
	result := ToSafeBranchName(repoBranch)
	assert.Equal(t, expectedSafeBranchName, result)
}
