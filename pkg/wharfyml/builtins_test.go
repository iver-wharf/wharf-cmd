package wharfyml

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListParentDirsPossibleBuiltinVarsFiles(t *testing.T) {
	currentDir := "/home/root/repos/my-repo"
	varFiles := listParentDirsPossibleVarsFiles(currentDir)
	want := []string{
		"/.wharf-vars.yml",
		"/home/.wharf-vars.yml",
		"/home/root/.wharf-vars.yml",
		"/home/root/repos/.wharf-vars.yml",
		"/home/root/repos/my-repo/.wharf-vars.yml",
	}
	got := make([]string, len(varFiles))
	for i, f := range varFiles {
		got[i] = f.path
	}
	assert.Equal(t, want, got)
}

func TestVarFilePrettyPath(t *testing.T) {
	currentDir := "/home/root/repos/my-repo"
	file := varFile{path: "/home/root/.wharf-vars.yml", source: varFileSourceParentDir}
	want := "../../.wharf-vars.yml"
	got := file.prettyPath(currentDir)
	assert.Equal(t, want, got)
}
