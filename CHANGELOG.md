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

- Added provisioner commands: (#46, #59, #117, #129)

  - `wharf provisioner serve` that launches an HTTP REST api server with
    endpoints:

    - `GET /` to ping.

    - `GET /api/swagger/index.html` Swagger generated documentation.

    - `POST /api/worker` creates a new worker with certain labels.

    - `GET /api/worker` gets a list of all workers with certain labels.

    - `DELETE /api/worker/:workerId` deletes a worker, as long as it has
      certain labels.

  - `wharf provisioner create` that creates a new worker.

  - `wharf provisioner list` that lists all running workers with certain
    labels.

  - `wharf provisioner delete` with flag `--id` to specify the worker that
    should be deleted, as long as it has certain labels as well.

- Added Git credentials support to `wharf provisioner` when running in
  Kubernetes via a Kubernetes secret named `wharf-cmd-worker-git-ssh`.
  See [docs/provisioner-git-ssh-secret.md](docs/provisioner-git-ssh-secret.md)
  for more info. (#120)

- Added watchdog commands: (#62, #129)

  - `wharf watchdog serve` checks stray builds from the wharf-api and
    wharf-cmd-workers from the wharf-cmd-provisioner and kills them in an effort
    to clean up forgotten builds/workers.

- Added aggregator command `wharf aggregator serve` that looks for
  wharf-cmd-worker pods and pipes build results over to the wharf-api.
  (#77, #126)

- Added new implementation for `wharf run`. (#33, #45, #66, #84, #107)

- Added "vars" command `wharf vars` with:

  - `wharf vars list` that prints out all the variables that
    would be used in a `wharf run` invocation. (#93, #98, #102, #108, #110)

  - `wharf vars sub` that reads from STDIN or a file and performs variable
    substitution, and then writes to STDOUT. (#110)

- Added support for `.gitignore` ignored files and directories when transferring
  repo in `wharf run`. Can be disabled via new `--no-gitignore` flag. (#85)

- Added input variables support using the `inputs` field in `.wharf-ci.yml`
  files, and the `--input, -i` flag to `wharf run` and `wharf vars` commands
  through the CLI, ex: (#97, #111)

  ```sh
  wharf run --input myInputVar=myValue
  # => [ "myInputVar": "myValue" ]

  # Supports multiple:
  wharf run --input var1=value1 --input var2=value2
  # => [ "var1": "value1", "var2": "value2" ]

  # On collisions, the last value is used:
  wharf run --input myVar=foo --input myVar=bar
  # => [ "myVar": "bar" ]
  ```

- Added new implementation for `.wharf-ci.yml` file parsing that now supports
  returning multiple errors for the whole parsing as well as keep track of the
  line & column of each parse error. (#48, #58)

- Added support for a new file type: `.wharf-vars.yml`. It is used to define
  built-in variables, and wharf looks for it in multiple files in the
  following order, where former files take precedence over latter files on a
  per-variable basis: (#73)

  - `./.wharf-vars.yml` (in same directory as `.wharf-ci.yml`)
  - `./../.wharf-vars.yml` (in parent directory of `.wharf-ci.yml`)
  - `./../../.wharf-vars.yml` (etc; it continues recursively)
  - (Linux only) `~/.config/iver-wharf/wharf-cmd/wharf-vars.yml`
  - (Linux only) `/etc/iver-wharf/wharf-cmd/wharf-vars.yml`
  - (Darwin/OS X only) `~/Library/Application Support/iver-wharf/wharf-cmd/wharf-vars.yml`
  - (Windows only) `%APPDATA%\iver-wharf\wharf-cmd\wharf-vars.yml`

  Note the leading dot in the directory tree files (`.wharf-vars.yml`), while
  the files from config folders is without the dot (`wharf-vars.yml`).

  The file content should be structured as:

  ```yml
  # .wharf-vars.yml

  vars:
    CHART_REPO: http://harbor.example.com
  ```

- Added support for using OS environment variables prefixed with `WHARF_VAR_`
  in variable substitution, where `WHARF_VAR_REG_URL` would set the `REG_URL`
  Wharf variable. (#96)

- Added variable substitution support for referenced files in `kubectl` and
  `helm` step types. (#89)

- Added file transfer cache, stored in `/tmp/wharf-cmd-repo-xxxxx/full.tar`,
  that is reused by all steps in a single build. (#89)

- Added build result (logs, status updates) caching via file system. New
  package in `pkg/resultstore`. (#43, #69, #70)

- Added all kubeconfig-related flags from `kubectl` but with a `--k8s-*` prefix.
  This allows e.g Wharf to run as a service account via the `--k8s-as` flag,
  among other things. (#63)

- Fixed `wharf run` and `wharf provisioner` commands not using the namespace
  defined in the kubeconfig. (#63)

- Added CLI completions via Cobra. See the completion command's help text for
  your shell for more info: (#64)

  ```bash
  wharf completion bash --help
  wharf completion fish --help
  wharf completion powershell --help
  wharf completion zsh --help
  ```

- Added `--stage` and `--environment` completions to `wharf run` based on the
  parsed `.wharf-ci.yml` file. (#91)

- Added `--loglevel` completions. (#95)

- Added Git integration by executing `git` locally to obtain current branch,
  commit SHA, tags, etc. (#67, #78)

- Fixed `wharf run` not reading a pod's logs when it fails immediately on start.
  (#50)

- Fixed `wharf run` not failing due to pod config errors, such as "secret
  'cluster-config' not found" in `kubectl` steps. (#52)

- Changed from `github.com/sirupsen/logrus` to
  `github.com/iver-wharf/wharf-core/pkg/logger` for logging. (#2, #7)

- Added gRPC server for worker in `pkg/workerapi/workerserver`: (#51)

  - `StreamLogs` batches logs into chunks and serves to gRPC clients.
  - `StreamStatusEvents` serves status events to gRPC clients.
  - `StreamArtifactEvents` serves artifact events to gRPC clients.

- Added gRPC client in `pkg/workerapi/workerclient` to interface with a worker
  gRPC server. (#51)

- Added HTTP server for worker in `pkg/workerapi/workerserver`:
  (#51, #114, #117)

  - `GET /` to ping.
  - `GET /api/swagger/index.html` Swagger generated documentation.
  - `GET /api/artifact/:artifactId/download` Downloads an artifact.

- Added HTTP client in `pkg/workerapi/workerclient` to interface with
  worker HTTP server. (#51)

- Added `--version`, `-v` flag to show the version of wharf-cmd. (#76)

- Added dependencies:

  - `github.com/alta/protopatch` v0.5.0 (#51)
  - `github.com/cli/safeexec` v1.0.0 (#78)
  - `github.com/denormal/go-gitignore` v0.0.0-20180930084346-ae8ad1d07817 (#85)
  - `github.com/gin-contrib/cors` v1.3.1 (#51)
  - `github.com/gin-gonic/gin` v1.7.1 (#46)
  - `github.com/golang/protobuf` v1.5.2 (#51)
  - `github.com/iver-wharf/wharf-api-client-go/v2` v2.0.0 (#62)
  - `github.com/iver-wharf/wharf-core` (#2, #7)
  - `github.com/soheilhy/cmux` v0.1.4 (#51)
  - `github.com/spf13/pflag` v1.0.5 (#63)
  - `github.com/swaggo/gin-swagger` v1.4.1 (#59)
  - `github.com/swaggo/swag` v1.8.0 (#59)
  - `google.golang.org/grpc` v1.44.0 (#51)
  - `google.golang.org/protobuf` v1.27.1 (#51)
  - `gopkg.in/guregu/null.v4` v4.0.0 (#62)
  - `gopkg.in/typ.v4` v4.0.0 (#75, #89, #127)
  - `gopkg.in/yaml.v3` v3.0.0 (#48)

- Removed dependencies:

  - `github.com/go-git/go-git` (#8)
  - `github.com/sirupsen/logrus` (#2)
  - `sigs.k8s.io/yaml` (#48)

- Removed commands `init`, `setup`, and `serve`. (#8)

- Changed versions of numerous dependencies:

  - `github.com/gin-gonic/gin` from v1.7.1 to v1.7.7 (#59)
  - `github.com/spf13/cobra` v1.1.3 to v1.3.0 (#64)
  - `k8s.io/api` from v0.0.0 to v0.23.3 (#8)
  - `k8s.io/apimachinery` from v0.0.0 to v0.23.3 (#8)
  - `k8s.io/client-go` from v0.0.0 to v0.23.3 (#8)
  - `sigs.k8s.io/yaml` from v1.1.0 to v1.2.0 (#8)

- Changed Go runtime from v1.13 to v1.18. (#8, #74)

- Changed logging on CLI errors (ex "unknown command") to be more terse. (#34)

- Changed to trim away everything before the last CR (carriage return)
  character in a log line from a Kubernetes pod. (#49)

- Changed location of packages and code files: (#44, #87)

  - File `pkg/core/utils/variablesreplacer.go` to its own package in `pkg/varsub`
  - Package `pkg/core/wharfyml` to `pkg/wharfyml`
  - Command `main.go` to `cmd/wharf/main.go`

- Removed packages: (#44)

  - `pkg/core/buildclient`
  - `pkg/core/containercreator`
  - `pkg/core/kubernetes`
  - `pkg/core/utils`
  - `pkg/namespace`
  - `pkg/run`

- Removed `containercreator` references from `pkg/core/wharfyml`. (#44)

- Added collecting of build logs and status updates for build steps using
  `resultstore`. (#71)

- Added cancelling of builds via signals (once to shutdown with a grace period,
  twice for a forceful shutdown): (#90, #104)

  - `os.Interrupt`
  - `os.Kill`
  - `syscall.SIGTERM`
  - `syscall.SIGHUP`

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
