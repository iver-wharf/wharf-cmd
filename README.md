# Wharf CI

A command-line interface to run tasks specified in a `.wharf-ci.yml` file.

## Commands

### Run

`wharf run --namespace build --environment stage wharf-ci.yml`

## Components

- HTTP API using the [gin-gonic/gin](https://github.com/gin-gonic/gin)
  web framework.

- Command-line parsing using [spf13/cobra](https://github.com/spf13/cobra)

- Kubernetes access using [k8s.io/client-go](https://github.com/kubernetes/client-go)

- Git interface using [go-git/go-git](https://github.com/go-git/go-git)

## Development

1. Install Go 1.13 or later: <https://golang.org/>

2. Start hacking with your favorite tool. For example VS Code, GoLand,
   Vim, Emacs, or whatnot.

3. Install formatter and linters:

   ```sh
   go install github.com/mgechev/revive@latest
   go install golang.org/x/tools/cmd/goimports@latest
   npm install
   ```

## Linting

You can lint all of the above at the same time by running:

```sh
make lint

make lint-go # only lint Go code
make lint-md # only lint Markdown files
```

Some errors can be fixed automatically. Keep in mind that this updates the
files in place.

```sh
make lint-fix

#make lint-fix-go # Go linter does not support fixes
make lint-fix-md # only lint and fix Markdown files
```

---

Maintained by [Iver](https://www.iver.com/en).
Licensed under the [MIT license](./LICENSE).
