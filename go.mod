module github.com/iver-wharf/wharf-cmd

go 1.16

require (
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/iver-wharf/wharf-api-client-go v1.2.0
	github.com/iver-wharf/wharf-core v1.3.0
	github.com/pborman/ansi v1.0.0
	github.com/spf13/cobra v1.1.3
	github.com/stretchr/testify v1.7.0
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	k8s.io/api v0.23.3
	k8s.io/apimachinery v0.23.3
	k8s.io/client-go v0.23.3
	sigs.k8s.io/yaml v1.2.0
)

replace golang.org/x/net => golang.org/x/net v0.0.0-20210119194325-5f4716e94777
