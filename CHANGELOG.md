# Wharf command changelog

This project tries to follow [SemVer 2.0.0](https://semver.org/).

<!--
	When composing new changes to this list, try to follow convention.

	The WIP release shall be updated just before adding the Git tag.
	From (WIP) to (YYYY-MM-DD), ex: (2021-02-09) for 9th of Febuary, 2021

	A good source on conventions can be found here:
	https://changelog.md/
-->

## v0.8.0 (WIP)

- Added new implementation for `wharf run`. (#33)

- Added dependency on `github.com/iver-wharf/wharf-core`. (#2, #7)

- Added build result (logs, status updates) caching via file system. New
  package in `pkg/resultstore`. (#43)

- Changed from `github.com/sirupsen/logrus` to
  `github.com/iver-wharf/wharf-core/pkg/logger` for logging. (#2, #7)

- Removed dependency on `github.com/sirupsen/logrus`. (#2)

- Removed commands `init`, `setup`, and `serve`. (#8)

- Changed versions of numerous dependencies: (#8)

  - `k8s.io/api` from v0.0.0 to v0.23.3
  - `k8s.io/apimachinery` from v0.0.0 to v0.23.3
  - `k8s.io/client-go` from v0.0.0 to v0.23.3
  - `sigs.k8s.io/yaml` from v1.1.0 to v1.2.0

- Removed dependencies: (#8)

  - `github.com/gin-gonic/gin`
  - `github.com/go-git/go-git`

- Changed Go runtime from v1.13 to v1.17. (#8)

- Changed logging on CLI errors (ex "unknown command") to be more terse. (#34)

## v0.7.0 (scrapped)

- Added parsing of `"environments"` fields in `.wharf-ci.yml` files. (!2)

- Added CHANGELOG.md to repository. (!8)

- Changed package structure, refactored out a lot of code from
  `/pkg/core/types.go` into two new packages, `/pkg/core/kubernetes` and
  `/pkg/core/wharfyml`. (!6)

- Changed code to comply better with "Go best practices" when it comes to,
  logging, variable naming, package naming, et.al. (!1)

- Fixed errors not getting properly returned from functions in the code base.
  (!5)

- Fixed cloning type error regression introduced when we updated
  `gopkg.in/src-d/go-git.v4`. (!3)

- Changed libs versions in mod file. (!10)

- Added new open sourced Wharf API client
  [github.com/iver-wharf/wharf-api-client-go](https://github.com/iver-wharf/wharf-api-client-go)
  v1.2.0. (!11, !14)

- Added `buildclient.Client` with posting logs and build statuses
  functionality. (!11)

- Added `ContainerReadyWaiter` interface with implementation. (!12)

- Added `StreamScanner` interface with implementation. (!13)

- Added `SanitizationFlags` for `StreamScanner`. (!13)

- Added `ContainerLogsReader` interface with implementation. (!13)

- Changed `StepType`. New parsing delivered. String method implemented. (!15)

- Added `ContainerStateWatcher` interface with implementation for done
  container and ready container. (!16)

- Changed `ContainerReadyWaiter` to use `ContainerStateWatcher` and renamed to
  `ContainerWaiter`. (!16)

- Added delete pod functionality. (!16)

- Added reading variables from `Environment` section in wharf-ci.yml file.
  (!17)

- Added replacement variables functionality for the step. (!17)

- Added `BuiltinVarType` type. Grabbed variables from URL and git repository.
  (!18)

- Added `Input` array parsing from `wharf-ci.yml` file. (!19)

- Changed `go-git` package version from v4.13.1 to v5.3.0. (!20)

## v0.6.0 (2020-02-04)

- Added initial proof of concept to build in Kubernetes, based on a
  `.wharf-ci.yml` file. (07abc2a4...77c28565)

- Added `go.mod` with dependency on Go 1.12. (6cbae31c)

- Added core Wharf library for parsing `.wharf-ci.yml` files.
  (3d0f3ae0, ce83ec59)

- Added commands: (07abc2a4, 387fbca9, 9a93e2c7)

  - `wharf init`

  - `wharf setup`

  - `wharf wharf-ci`
    *Ci application to generate .wharf-ci.yml files and execute them against a
    kubernetes cluster*

  - `wharf run`
    *Run the specified .wharf-ci.yml file against kubernetes*

- Added global arguments: (07abc2a4, 021c02ce)

  - `wharf --loglevel info`
    *Show debug information*

  - `wharf --kubeconfig ~/.kube/config`
    *Path to kubeconfig file*

- Added CLI arguments parsing via
  [github.com/spf13/cobra](https://github.com/spf13/cobra). (07abc2a4)
