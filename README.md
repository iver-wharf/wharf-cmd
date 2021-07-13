# Wharf CI

A command-line interface to run tasks specified in a `.wharf-ci.yml` file.

## Commands

### Init

`wharf init .`

### Run

`wharf run --namespace build --environment stage wharf-ci.yml`

### Setup

`wharf setup --namespace stage --namespace prod --service-account-namespace default`

### Serve

`wharf serve --listen :8080 --kubeconfig=~/.kube/config`

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

## Linting Golang

- Requires Node.js (npm) to be installed: <https://nodejs.org/en/download/>
- Requires Revive to be installed: <https://revive.run/>

```sh
go get -u github.com/mgechev/revive
```

```sh
npm run lint-go
```

## Linting markdown

- Requires Node.js (npm) to be installed: <https://nodejs.org/en/download/>

```sh
npm install

npm run lint-md

# Some errors can be fixed automatically. Keep in mind that this updates the
# files in place.
npm run lint-md-fix
```

## Linting

You can lint all of the above at the same time by running:

```sh
npm run lint

# Some errors can be fixed automatically. Keep in mind that this updates the
# files in place.
npm run lint-fix
```

---

Maintained by [Iver](https://www.iver.com/en).
Licensed under the [MIT license](./LICENSE).
