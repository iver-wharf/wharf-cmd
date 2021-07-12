module github.com/iver-wharf/wharf-cmd

go 1.13

require (
	github.com/gin-gonic/gin v1.7.1
	github.com/go-git/go-git/v5 v5.3.0
	github.com/iver-wharf/wharf-api-client-go v1.2.0
	github.com/iver-wharf/wharf-core v0.0.0-20210702112246-1601d2a7dd23
	github.com/pborman/ansi v1.0.0
	github.com/spf13/cobra v1.1.3
	github.com/stretchr/testify v1.7.0
	k8s.io/api v0.0.0-20200131232428-e3a917c59b04
	k8s.io/apimachinery v0.0.0-20200409202947-6e7c4b1e1854
	k8s.io/client-go v0.0.0-20200410023015-75e09fce8f36
	sigs.k8s.io/yaml v1.1.0
)

replace golang.org/x/net => golang.org/x/net v0.0.0-20210119194325-5f4716e94777
