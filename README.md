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
