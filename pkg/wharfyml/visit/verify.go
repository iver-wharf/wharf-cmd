package visit

import (
	"fmt"

	"github.com/iver-wharf/wharf-cmd/internal/errutil"
	"gopkg.in/yaml.v3"
)

// VerifyKind checks if the given node has the wanted kind. The wantStr is only
// used as a pretty string of the wanted type in the potential error message.
func VerifyKind(node *yaml.Node, wantStr string, wantKind yaml.Kind) error {
	if node.Kind != wantKind {
		return errutil.NewPosFromNode(fmt.Errorf("%w: expected %s, but was %s",
			ErrInvalidFieldType, wantStr, PrettyNodeTypeName(node)), node)
	}
	return nil
}

// VerifyTag checks if the given node has the wanted tag. The wantStr is only
// used as a pretty string of the wanted type in the potential error message.
func VerifyTag(node *yaml.Node, wantStr string, wantTag string) error {
	gotTag := node.ShortTag()
	if gotTag != wantTag {
		return errutil.NewPosFromNode(fmt.Errorf("%w: expected %s, but was %s",
			ErrInvalidFieldType, wantStr, PrettyNodeTypeName(node)), node)
	}
	return nil
}

// VerifyKindAndTag checks if the given node has both the wanted kind and then
// also then wanted tag. The wantStr is only used as a pretty string of the
// wanted type in the potential error message.
func VerifyKindAndTag(node *yaml.Node, wantStr string, wantKind yaml.Kind, wantTag string) error {
	if err := VerifyKind(node, wantStr, wantKind); err != nil {
		return err
	}
	return VerifyTag(node, wantStr, wantTag)
}
