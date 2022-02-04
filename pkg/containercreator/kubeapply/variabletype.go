package kubeapply

type VariableType int

const (
	File = VariableType(iota + 1)
	Files
	Namespace
	Cluster
	Action
	Force
)

var toVT = map[VariableType]string{
	File:      "file",
	Files:     "files",
	Namespace: "namespace",
	Cluster:   "cluster",
	Action:    "action",
	Force:     "force",
}

var vtToStr = map[VariableType]string{}

func (t VariableType) String() string {
	return toVT[t]
}

var toDefaultVT = map[VariableType]string{
	File:      "",
	Files:     "",
	Namespace: "",
	Cluster:   "kubectl-config",
	Action:    "apply",
	Force:     "false",
}

func (t VariableType) getDefault() string {
	return toDefaultVT[t]
}
