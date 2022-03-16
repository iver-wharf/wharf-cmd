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
		got[i] = f.Path
	}
	assert.Equal(t, want, got)
}

func TestVarFilePrettyPath(t *testing.T) {
	currentDir := "/home/root/repos/my-repo"
	file := VarFile{Path: "/home/root/.wharf-vars.yml", Kind: VarFileKindParentDir}
	want := "../../.wharf-vars.yml"
	got := file.PrettyPath(currentDir)
	assert.Equal(t, want, got)
}
