package wharfyml

import (
	"testing"

	"github.com/iver-wharf/wharf-cmd/pkg/varsub"
	"github.com/stretchr/testify/assert"
)

var testVarSource = varsub.SourceMap{
	"REPO_GROUP": varsub.Val{Value: "iver-wharf"},
	"REPO_NAME":  varsub.Val{Value: "wharf-cmd"},
	"REG_URL":    varsub.Val{Value: "http://harbor.example.com"},
	"CHART_REPO": varsub.Val{Value: "http://charts.example.com"},
}

func TestListParentDirsPossibleBuiltinVarsFiles(t *testing.T) {
	currentDir := "/home/root/repos/my-repo"
	varFiles := listParentDirsPossibleVarsFiles(currentDir)
	want := []string{
		"/home/root/repos/my-repo/.wharf-vars.yml",
		"/home/root/repos/.wharf-vars.yml",
		"/home/root/.wharf-vars.yml",
		"/home/.wharf-vars.yml",
		"/.wharf-vars.yml",
	}
	got := make([]string, len(varFiles))
	for i, f := range varFiles {
		got[i] = f.Path
	}
	assert.Equal(t, want, got)
}

func TestVarFilePrettyPath(t *testing.T) {
	currentDir := "/home/root/repos/my-repo"
	file := VarFile{Path: "/home/root/.wharf-vars.yml", IsRel: true}
	want := "../../.wharf-vars.yml"
	got := file.PrettyPath(currentDir)

	assert.Equal(t, want, got)
}
