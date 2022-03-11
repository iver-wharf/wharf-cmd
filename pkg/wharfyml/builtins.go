package wharfyml

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/iver-wharf/wharf-cmd/internal/slices"
	"github.com/iver-wharf/wharf-cmd/pkg/varsub"
)

const (
	builtInVarsConfFile = "wharf-vars.yml"
	builtInVarsDotfile  = ".wharf-vars.yml"
)

func ParseBuiltinVars() (varsub.Source, error) {
	return nil, errors.New("not implemented")
}

func listPossibleBuiltinVarsFiles(currentDir string) []string {
	paths := listOSPossibleBuiltInVars()

	confDir, err := os.UserConfigDir()
	if err == nil {
		paths = append(paths,
			filepath.Join(confDir, "iver-wharf", "wharf-cmd", builtInVarsConfFile),
		)
	}

	paths = append(paths, listParentDirsPossibleBuiltinVarsFiles(currentDir)...)

	return paths
}

func listParentDirsPossibleBuiltinVarsFiles(currentDir string) []string {
	var paths []string
	for {
		paths = append(paths, filepath.Join(currentDir, builtInVarsDotfile))
		prevDir := currentDir
		currentDir = filepath.Dir(currentDir)
		if prevDir == currentDir {
			break
		}
	}
	// We reverse it because we want the path closest to the current dir
	// to be merged in last into the varsub.Source.
	slices.ReverseStrings(paths)
	return paths
}
