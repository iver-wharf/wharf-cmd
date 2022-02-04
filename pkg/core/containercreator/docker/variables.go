package docker

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/iver-wharf/wharf-cmd/pkg/core/containercreator"
	"github.com/iver-wharf/wharf-cmd/pkg/core/containercreator/kaniko"
	"github.com/iver-wharf/wharf-cmd/pkg/core/containercreator/utils"
	"github.com/iver-wharf/wharf-core/pkg/logger"
)

const (
	insecureArg    = "--insecure"
	noPushArg      = "--no-push"
	buildArgs      = "--build-arg"
	destinationArg = "--destination"
)

var log = logger.New()

type Variables map[string]interface{}

func (v Variables) getBoolParam(t VariableType, stageName string, secret string, builtinVars map[containercreator.BuiltinVar]string) bool {
	strValue := v.GetVariableOrDefault(t, stageName, secret, builtinVars)
	value, err := strconv.ParseBool(strValue)
	if err != nil {
		log.Error().WithError(err).
			WithStringer("type", t).
			WithString("value", strValue).
			WithString("stage", stageName).
			Message("Unable to parse bool param.")
		value = false
	}
	return value
}

func (v Variables) GetImageDestination(stageName string, secret string, builtinVars map[containercreator.BuiltinVar]string) string {
	registry := v.GetVariableOrDefault(Registry, stageName, secret, builtinVars)
	group := v.GetVariableOrDefault(Group, stageName, secret, builtinVars)
	name := v.GetVariableOrDefault(Name, stageName, secret, builtinVars)
	destination := fmt.Sprintf("%s/%s/%s", registry, group, builtinVars[containercreator.BuiltinVarRepoName])
	if name != "" && name != builtinVars[containercreator.BuiltinVarRepoName] {
		return fmt.Sprintf("%s/%s", destination, strings.ToLower(name))
	}
	return destination
}

func (v Variables) GetVariableOrDefault(t VariableType, stageName string, secret string, builtinVars map[containercreator.BuiltinVar]string) string {
	variable, ok := v[t.String()].(string)
	if !ok {
		return t.getDefault(stageName, secret, builtinVars)
	}

	return variable
}

func (v Variables) GetScript(stageName string, secret string, builtinVars map[containercreator.BuiltinVar]string) string {
	script := kaniko.GetScript()
	params := utils.GetListOfParamsNames(script)
	for key, value := range params {
		newValue := v.getStringValue(ParseParam(key), stageName, secret, builtinVars)
		script = strings.Replace(script, value, newValue, -1)
	}
	return script
}

func (v Variables) getStringValue(t ParamType, stageName string, secret string, builtinVars map[containercreator.BuiltinVar]string) string {
	switch t {
	case FileArgs:
		return v.GetVariableOrDefault(Args, stageName, secret, builtinVars)
	case FileAppendCert:
		if v.getBoolParam(Cert, stageName, secret, builtinVars) {
			return "true"
		}
		return "false"
	case FileContext:
		return v.GetVariableOrDefault(Context, stageName, secret, builtinVars)
	case FilePath:
		return v.GetVariableOrDefault(File, stageName, secret, builtinVars)
	case ImageDestination:
		return v.GetImageDestination(stageName, secret, builtinVars)
	case ImageTags:
		return v.GetVariableOrDefault(Tag, stageName, secret, builtinVars)
	case ImagePush:
		push := v.getBoolParam(Push, stageName, secret, builtinVars)
		if push {
			return "true"
		}
		return "false"
	case Insecure:
		return "false"
	case RegSecret:
		return v.GetVariableOrDefault(Secret, stageName, secret, builtinVars)
	case RootCert:
		return "wget --no-check-certificate -q -O - https://gitlab.dgc.local/snippets/9/raw || echo 'No cert available'"
	case KanikoArgs:
		fileArgs := v.getStringValue(FileArgs, stageName, secret, builtinVars)
		insecure := v.getBoolParam(Push, stageName, secret, builtinVars)
		return getKanikoArgs(fileArgs, insecure, false)
	case StringDest:
		imageTags := v.getStringValue(ImageTags, stageName, secret, builtinVars)
		destinations := v.getStringValue(ImageDestination, stageName, secret, builtinVars)
		return getDestinations(imageTags, destinations)
	}

	return ""
}

func getKanikoArgs(dockerFileArgs string, insecure bool, imagePush bool) string {
	var kanikoArgs []string

	for _, arg := range strings.Split(dockerFileArgs, "\n") {
		if arg != "" {
			esc := strings.Replace(arg, `/`, `\\`, -1)
			kanikoArgs = append(kanikoArgs, fmt.Sprintf("%s '%s'", buildArgs, esc))
		}
	}

	if insecure {
		kanikoArgs = append(kanikoArgs, insecureArg)
	}

	if !imagePush {
		kanikoArgs = append(kanikoArgs, noPushArg)
	}

	return strings.Join(kanikoArgs, " ")
}

func getDestinations(imageTags string, imageDestination string) string {
	var destinations []string
	for _, tag := range strings.Split(imageTags, ",") {
		if tag != "" {
			destinations = append(destinations, fmt.Sprintf("%s=%s:%s", destinationArg, imageDestination, tag))
		}
	}
	return strings.Replace(strings.Join(destinations, " "), "\n", "", -1)
}
