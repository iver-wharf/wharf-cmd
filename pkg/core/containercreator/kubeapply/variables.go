package kubeapply

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/iver-wharf/wharf-cmd/pkg/core/containercreator/utils"
	"github.com/iver-wharf/wharf-core/pkg/logger"
)

const (
	force         = "--force"
	defaultScript = `#!/bin/bash
IFS='' read -r -d '' String <<"EOF"
${KubeResource}
EOF

echo "\${String}" > values.yml
echo "Applying YAML:"
echo
cat values.yml
echo

echo '\$' kubectl ${KubeAction} ${KubeNamespace} -f values.yml ${KubeForce}
kubectl ${KubeAction} ${KubeNamespace} -f values.yml ${KubeForce}`
)

var log = logger.New()

type Variables map[string]interface{}

func (v Variables) GetScript() string {
	resource := v.getStringValue(KubeResource)
	if resource == "" {
		return ""
	}

	return v.replaceVariables(defaultScript)
}

func (v Variables) getStringValue(t ParamType) string {
	switch t {
	case KubeResource:
		return v.replaceVariables(v.getResource())
	case KubeForce:
		f := v.getBoolParam(Force)
		if f {
			return force
		}
		return ""
	case KubeNamespace:
		namespace := v.GetVariableOrDefault(Namespace)
		if namespace == "" {
			return ""
		}

		return fmt.Sprintf("-n %s", namespace)
	case KubeAction:
		return v.GetVariableOrDefault(Action)
	case ConfigMapName:
		return v.GetVariableOrDefault(Cluster)

	}

	return ""
}

func (v Variables) getResource() string {
	// TODO: replace variables from environments. @estefans
	content := ""
	file := v.GetVariableOrDefault(File)
	fileContent, err := v.getFileContent(file)
	if err == nil {
		content += fileContent
	}

	files, ok := v[Files.String()].([]string)
	if !ok {
		return content
	}

	content += "\n---\n"

	for _, f := range files {
		fileContent, err := v.getFileContent(f)
		if err == nil {
			content += fileContent + "\n---\n"
		}
	}
	return content
}

func (v Variables) getFileContent(fileName string) (string, error) {
	if fileName == "" {
		return "", errors.New("invalid file name")
	}

	filecontent, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Error().WithString("file", fileName).Message("Unable to read file content.")
	}
	return v.replaceVariables(string(filecontent)), nil
}

func (v Variables) replaceVariables(content string) string {
	if content == "" {
		return ""
	}

	params := utils.GetListOfParamsNames(content)
	for key, value := range params {
		newValue := v.getStringValue(ParseParam(key))
		content = strings.Replace(content, value, newValue, -1)
	}
	return content
}

func (v Variables) GetVariableOrDefault(t VariableType) string {
	variable, ok := v[t.String()].(string)
	if !ok {
		return t.getDefault()
	}

	return variable
}

func (v Variables) getBoolParam(t VariableType) bool {
	strValue, ok := v[t.String()].(string)
	if !ok {
		return false
	}

	value, err := strconv.ParseBool(strValue)
	if err != nil {
		log.Error().WithError(err).
			WithStringer("type", t).
			WithString("value", strValue).
			Message("Unable to parse bool param.")
		value = false
	}
	return value
}
