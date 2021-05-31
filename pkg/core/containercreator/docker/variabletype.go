package docker

import (
	"strconv"
	"strings"

	"github.com/iver-wharf/wharf-cmd/pkg/core/containercreator"
	"github.com/iver-wharf/wharf-cmd/pkg/core/containercreator/kaniko"
)

type VariableType int

const (
	Tag = VariableType(iota + 1)
	Push
	Context
	Secret
	Args
	Backend
	Name
	Registry
	Group
	Cert
	File
)

var toVT = map[VariableType]string{
	Tag:      "tag",
	Push:     "push",
	Context:  "context",
	Secret:   "secret",
	Args:     "args",
	Backend:  "backend",
	Name:     "name",
	Registry: "registry",
	Group:    "group",
	Cert:     "append-cert",
	File:     "file",
}

func (t VariableType) String() string {
	return toVT[t]
}

func (t VariableType) getDefault(stageName string, secret string, builtinVars map[containercreator.BuiltinVar]string) string {
	switch t {
	case Push:
		return "true"
	case Secret:
		return secret
	case Backend:
		return kaniko.ContainerName
	case Name:
		return getJobName(stageName, builtinVars[containercreator.BuiltinVarRepoName])
	case Registry:
		return builtinVars[containercreator.BuiltinVarRegURL]
	case Group:
		return builtinVars[containercreator.BuiltinVarRepoGroup]
	case Cert:
		return strconv.FormatBool(strings.HasPrefix(strings.ToLower(builtinVars[containercreator.BuiltinVarRepoName]), "default"))
	case File:
		return "file"
	}
	return ""
}

func getJobName(stageName string, repoName string) string {
	jobName := stageName
	if stageName == repoName {
		jobName = ""
	}
	return jobName
}
