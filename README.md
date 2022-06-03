# Wharf CI

[![Codacy Badge](https://app.codacy.com/project/badge/Grade/2e59b0814f174cb2bebda4870797e15c)](https://www.codacy.com/gh/iver-wharf/wharf-cmd/dashboard?utm_source=github.com\&utm_medium=referral\&utm_content=iver-wharf/wharf-cmd\&utm_campaign=Badge_Grade)
[![Go Reference](https://pkg.go.dev/badge/github.com/iver-wharf/wharf-cmd.svg)](https://pkg.go.dev/github.com/iver-wharf/wharf-cmd)

A command-line interface to run tasks specified in a `.wharf-ci.yml` file.

## Installation

Requires Go 1.18 (or later)

```sh
go install github.com/iver-wharf/wharf-cmd/cmd/wharf@latest
```

## Commands

### Run

`wharf run --namespace build --environment stage wharf-ci.yml`

## Components

- HTTP API using the [gin-gonic/gin](https://github.com/gin-gonic/gin)
  web framework.

- gRPC API using [grpc/grpc-go](https://github.com/grpc/grpc-go).

- Command-line parsing using [spf13/cobra](https://github.com/spf13/cobra)

- Kubernetes access using [k8s.io/client-go](https://github.com/kubernetes/client-go)

## Development

1. Install

   - Go 1.18 or later (for compilation): <https://golang.org/>
   - NodeJS & NPM (for markdown linting): <https://nodejs.org/en/>
   - Protobuf runtime (for regenerating protobuf/gRPC files): <https://developers.google.com/protocol-buffers/>

2. Install formatters, protobuf dependencies, and linters:

   ```sh
   make deps
   ```

3. Start hacking with your favorite tool. For example VS Code, GoLand,
   Vim, Emacs, or whatnot.

## Linting

```sh
make lint

make lint-go # only lint Go code
make lint-md # only lint Markdown files
make lint-proto # only lint Protobuf (gRPC) files
```

Some errors can be fixed automatically. Keep in mind that this updates the
files in place.

```sh
make lint-fix

make lint-fix-go # only lint and fix Go files
make lint-fix-md # only lint and fix Markdown files
make lint-fix-proto # only lint and fix Protobuf (gRPC) files
```

---

Maintained by [Iver](https://www.iver.com/en).
Licensed under the [MIT license](./LICENSE).
