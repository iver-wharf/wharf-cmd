package wharfyml

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testVarSource struct{}

func (testVarSource) Lookup(name string) (interface{}, bool) {
	switch name {
	case "REPO_GROUP":
		return "iver-wharf", true
	case "REPO_NAME":
		return "wharf-cmd", true
	case "REG_URL":
		return "http://harbor.example.com", true
	case "CHART_REPO":
		return "http://harbor.example.com", true
	default:
		return nil, false
	}
}

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
