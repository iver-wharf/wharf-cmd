package wharfyml

type StepType int

const (
	Container = StepType(iota + 1)
	Kaniko
	Docker
	HelmDeploy
	HelmPackage
	KubeApply
)

var strToStepType = map[string]StepType{
	"container":    Container,
	"kaniko":       Kaniko,
	"helm":         HelmDeploy,
	"helm-package": HelmPackage,
	"kubectl":      KubeApply,
	"docker":       Docker,
}

var stepTypeToString = map[StepType]string{}

func init() {
	for str, st := range strToStepType {
		stepTypeToString[st] = str
	}
}

func (t StepType) String() string {
	return stepTypeToString[t]
}

func ParseStepType(name string) StepType {
	return strToStepType[name]
}
