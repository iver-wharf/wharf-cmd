package visit

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// Pos represents a position. Used to declare where an object was defined in
// the .wharf-ci.yml file. The first line and column starts at 1.
// The zero value is used to represent an undefined position.
type Pos struct {
	Line   int
	Column int
}

// String implements fmt.Stringer
func (p Pos) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Column)
}

// NewPosFromNode creates a new Pos using a YAML node's Line and Column.
func NewPosFromNode(node *yaml.Node) Pos {
	return Pos{
		Line:   node.Line,
		Column: node.Column,
	}
}
