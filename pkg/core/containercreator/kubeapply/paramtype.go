package kubeapply

type ParamType int

const (
	KubeResource = ParamType(iota + 1)
	KubeForce
	KubeNamespace
	KubeAction
	ConfigMapName
)

var toPT = map[string]ParamType{
	"KubeResource":  KubeResource,
	"KubeForce":     KubeForce,
	"KubeNamespace": KubeNamespace,
	"KubeAction":    KubeAction,
	"ConfigMapName": ConfigMapName,
}

var ptToStr = map[ParamType]string{}

func init() {
	for str, st := range toPT {
		ptToStr[st] = str
	}
}

func (t ParamType) String() string {
	return ptToStr[t]
}

func ParseParam(name string) ParamType {
	return toPT[name]
}
