package wharfyml

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListParentDirsPossibleBuiltinVarsFiles(t *testing.T) {
	currentDir := "/home/root/repos/my-repo"
	got := listParentDirsPossibleBuiltinVarsFiles(currentDir)
	want := []string{
		"/.wharf-vars.yml",
		"/home/.wharf-vars.yml",
		"/home/root/.wharf-vars.yml",
		"/home/root/repos/.wharf-vars.yml",
		"/home/root/repos/my-repo/.wharf-vars.yml",
	}
	assert.Equal(t, want, got)
}
