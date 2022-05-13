package visit

import (
	"fmt"

	"github.com/iver-wharf/wharf-cmd/internal/errutil"
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

func NewPosFromNode(node *yaml.Node) Pos {
	return Pos{
		Line:   node.Line,
		Column: node.Column,
	}
}

func WrapPosError(err error, pos Pos) error {
	return errutil.Pos{
		Err:    err,
		Line:   pos.Line,
		Column: pos.Column,
	}
}

func WrapPosErrorNode(err error, node *yaml.Node) error {
	return WrapPosError(err, NewPosFromNode(node))
}
