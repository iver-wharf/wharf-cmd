package wharfyml

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/iver-wharf/wharf-cmd/internal/errutil"
	"github.com/iver-wharf/wharf-cmd/internal/util"
	"github.com/iver-wharf/wharf-cmd/pkg/varsub"
	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml/visit"
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
func ParseVarFiles(currentDir string) (varsub.Source, errutil.Slice) {
	workingDir, err := os.Getwd()
	if err != nil {
		log.Warn().WithError(err).
			Message("Failed getting working directory. Printing paths relative to the .wharf-ci.yml file instead.")
		workingDir = currentDir
	}

	varFiles := ListPossibleVarsFiles(currentDir)
	var errSlice errutil.Slice
	var filesSources varsub.SourceSlice
	for _, varFile := range varFiles {
		items, errs := tryReadVarsFileNodes(varFile.Path)
		prettyPath := varFile.PrettyPath(workingDir)
		errSlice = append(errSlice,
			errutil.ScopeSlice(errs, prettyPath)...)
		if len(items) == 0 {
			continue
		}
		source := make(varsub.SourceMap)
		for _, item := range items {
			source[item.Key.Value] = varsub.Val{
				Value:  visit.VarSubNode{item.Value},
				Source: prettyPath,
			}
		}
		filesSources = append(filesSources, source)
	}
	return filesSources, errSlice
}

func tryReadVarsFileNodes(path string) ([]visit.MapItem, errutil.Slice) {
	file, err := os.Open(path)
	if err != nil {
		// Silently ignore. Could not exist, be a directory, or not readable.
		// We don't care either way. We can't read it, so we ignore it.
		return nil, nil
	}
	defer file.Close()
	return parseVarsFileNodes(file)
}

func parseVarsFileNodes(reader io.Reader) ([]visit.MapItem, errutil.Slice) {
	rootNodes, err := visit.DecodeRootNodes(reader)
	if err != nil {
		return nil, errutil.Slice{err}
	}
	return visitVarsFileRootNodes(rootNodes)
}

func visitVarsFileRootNodes(rootNodes []*yaml.Node) ([]visit.MapItem, errutil.Slice) {
	var allVars []visit.MapItem
	var errSlice errutil.Slice
	for i, root := range rootNodes {
		docNodes, errs := visit.MapSlice(root)
		if len(rootNodes) > 1 {
			errSlice.Add(errutil.ScopeSlice(errs, fmt.Sprintf("doc#%d", i+1))...)
		} else {
			errSlice.Add(errs...)
		}
		for _, node := range docNodes {
			if node.Key.Value != propVars {
				// just silently ignore
				continue
			}
			vars, errs := visit.MapSlice(node.Value)
			errSlice.Add(errutil.ScopeSlice(errs, propVars)...)
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
	return util.ShorthandHome(f.Path)
}

// ListPossibleVarsFiles returns all paths where we look for wharf-vars.yml and
// .wharf-vars.yml files.
//
// Returned paths include the filename.
//
// The ordering of the returned filenames are in the order of which file should
// have priority over the other; with the file of highest priority that should
// override all the others, first.
func ListPossibleVarsFiles(currentDir string) []VarFile {
	varFiles := listParentDirsPossibleVarsFiles(currentDir)

	confDir, err := os.UserConfigDir()
	if err == nil {
		varFiles = append(varFiles, VarFile{
			Path:  filepath.Join(confDir, "iver-wharf", "wharf-cmd", builtInVarsFile),
			IsRel: false,
		})
	}

	varFiles = append(varFiles, listOSPossibleVarsFiles()...)

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
	return varFiles
}
