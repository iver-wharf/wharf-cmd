package worker

// Config holds settings for the worker.
type Config struct {
	K8s   K8sConfig
	Steps StepsConfig
}

// K8sConfig holds settings when the worker is using kubernetes.
type K8sConfig struct {
	// Context is the context used when talking to kubernetes.
	Context string
	// Namespace is the kubernetes namespace to talk to.
	Namespace string
}

// StepsConfig holds settings for the different types of steps.
type StepsConfig struct {
	Docker  DockerStepConfig
	Kubectl KubectlStepConfig
	Helm    HelmStepConfig
}

// DockerStepConfig holds settings for the docker step type.
type DockerStepConfig struct {
	// KanikoImage is the image for the kaniko executor to use in docker steps.
	//
	// Added in v0.8.0
	KanikoImage string
}

// KubectlStepConfig holds settings for the kubectl step type.
type KubectlStepConfig struct {
	// KubectlImage is the image to use in kubectl steps.
	//
	// Added in v0.8.0
	KubectlImage string
}

// HelmStepConfig holds settings for the helm step type.
type HelmStepConfig struct {
	// HelmImage is the image to use in helm steps.
	//
	// Added in v0.8.0
	HelmImage string
}
