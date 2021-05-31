package kubernetes

type ContainerType bool

const (
	// ContainerTypeApp represents an app-container of a Kubernetes Pod.
	// In contrast to the ContainerTypeInit that represents the init-container.
	ContainerTypeApp = ContainerType(false)
	// ContainerTypeInit represents an init-container of a Kubernetes Pod.
	// In contrast to the ContainerTypeApp that represents the app-container.
	ContainerTypeInit = ContainerType(true)
)

func (t ContainerType) String() string {
	if t == ContainerTypeInit {
		return "init-container"
	} else {
		return "app-container"
	}
}
