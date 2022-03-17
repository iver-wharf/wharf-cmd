package wharfyml

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/iver-wharf/wharf-cmd/internal/slices"
	"github.com/iver-wharf/wharf-cmd/pkg/varsub"
	"gopkg.in/yaml.v3"
)

const (
	builtInVarsFile    = "wharf-vars.yml"
	builtInVarsDotfile = ".wharf-vars.yml"
)

// ParseVarFiles produces a varsub.Source of *yaml.Node values. The files it
// checks depends on your OS.
//
// For GNU/Linux:
// 	/etc/iver-wharf/wharf-cmd/wharf-vars.yml
// 	$XDG_CONFIG_HOME/iver-wharf/wharf-cmd/wharf-vars.yml
// 	(if $XDG_CONFIG_HOME is unset) $HOME/.config/iver-wharf/wharf-cmd/wharf-vars.yml
//
// For Windows:
// 	%APPDATA%\iver-wharf\wharf-cmd\wharf-vars.yml
//
// For Darwin (Mac OS X):
// 	$HOME/Library/Application Support/iver-wharf/wharf-cmd/wharf-vars.yml
//
// In addition, this function also checks the current directory and parent
// directories above it, recursively, for a dotfile variant (.wharf-vars.yml):
// 	./.wharf-vars.yml
// 	../.wharf-vars.yml
// 	../../.wharf-vars.yml
// 	../../..(etc)/.wharf-vars.yml
func ParseVarFiles(currentDir string) (varsub.Source, Errors) {
	varFiles := ListPossibleVarsFiles(currentDir)
	var errSlice Errors
	nodeVarSource := make(varsub.SourceMap)
	for _, varFile := range varFiles {
		items, errs := tryReadVarsFileNodes(varFile.Path)
		errSlice = append(errSlice,
			wrapPathErrorSlice(errs, varFile.PrettyPath(currentDir))...)
		for _, item := range items {
			nodeVarSource[item.key.value] = item.value
		}
	}
	return nodeVarSource, errSlice
}

func tryReadVarsFileNodes(path string) ([]mapItem, Errors) {
	file, err := os.Open(path)
	if err != nil {
		// Silently ignore. Could not exist, be a directory, or not readable.
		// We don't care either way. We can't read it, so we ignore it.
		return nil, nil
	}
	defer file.Close()
	return parseVarsFileNodes(file)
}

func parseVarsFileNodes(reader io.Reader) ([]mapItem, Errors) {
	rootNodes, err := decodeRootNodes(reader)
	if err != nil {
		return nil, Errors{err}
	}
	return visitVarsFileRootNodes(rootNodes)
}

func visitVarsFileRootNodes(rootNodes []*yaml.Node) ([]mapItem, Errors) {
	var allVars []mapItem
	var errSlice Errors
	for i, root := range rootNodes {
		docNodes, errs := visitMapSlice(root)
		if len(rootNodes) > 1 {
			errSlice.add(wrapPathErrorSlice(errs, fmt.Sprintf("doc#%d", i+1))...)
		} else {
			errSlice.add(errs...)
		}
		for _, node := range docNodes {
			if node.key.value != propVars {
				// just silently ignore
				continue
			}
			vars, errs := visitMapSlice(node.value)
			errSlice.add(wrapPathErrorSlice(errs, propVars)...)
			allVars = append(allVars, vars...)
		}
	}
	return allVars, errSlice
}

// VarFile is a place and kind definition of a variable file.
type VarFile struct {
	Path  string
	IsRel bool
}

// PrettyPath returns a formatted version of the path, based on if its relative,
// and using "~" as shorthand for the user's home directory.
func (f VarFile) PrettyPath(currentDir string) string {
	if f.IsRel {
		rel, err := filepath.Rel(currentDir, f.Path)
		if err == nil {
			return rel
		}
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return f.Path
	}
	return useShorthandHomePrefix(f.Path, home)
}

func useShorthandHomePrefix(path, home string) string {
	if !strings.HasPrefix(path, home) {
		return path
	}
	return "~" + strings.TrimPrefix(path, home)
}

// ListPossibleVarsFiles returns all paths where we look for wharf-vars.yml and
// .wharf-vars.yml files.
//
// Returned paths include the filename.
func ListPossibleVarsFiles(currentDir string) []VarFile {
	varFiles := listOSPossibleVarsFiles()

	confDir, err := os.UserConfigDir()
	if err == nil {
		varFiles = append(varFiles, VarFile{
			Path:  filepath.Join(confDir, "iver-wharf", "wharf-cmd", builtInVarsFile),
			IsRel: false,
		})
	}

	varFiles = append(varFiles, listParentDirsPossibleVarsFiles(currentDir)...)
	return varFiles
}

func listParentDirsPossibleVarsFiles(currentDir string) []VarFile {
	var varFiles []VarFile
	for {
		varFiles = append(varFiles, VarFile{
			Path:  filepath.Join(currentDir, builtInVarsDotfile),
			IsRel: true,
		})
		prevDir := currentDir
		currentDir = filepath.Dir(currentDir)
		if prevDir == currentDir {
			break
		}
	}
	// We reverse it because we want the path closest to the current dir
	// to be merged in last into the varsub.Source.
	slices.Reverse(len(varFiles), func(i, j int) {
		varFiles[i], varFiles[j] = varFiles[j], varFiles[i]
	})
	return varFiles
}
