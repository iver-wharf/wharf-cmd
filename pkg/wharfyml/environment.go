package wharfyml

import "github.com/goccy/go-yaml/ast"

type Environment struct {
	Variables map[string]interface{}
}

type Env struct {
	Name string
	Vars map[string]interface{}
}

func parseEnvironments(mapItem ast.Node) (map[string]Env, []error) {
	return nil, nil
}
