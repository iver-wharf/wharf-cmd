package visit

import (
	"fmt"

	"github.com/iver-wharf/wharf-cmd/internal/errutil"
	"gopkg.in/yaml.v3"
)

func VerifyKind(node *yaml.Node, wantStr string, wantKind yaml.Kind) error {
	if node.Kind != wantKind {
		return errutil.NewPosNode(fmt.Errorf("%w: expected %s, but was %s",
			ErrInvalidFieldType, wantStr, PrettyNodeTypeName(node)), node)
	}
	return nil
}

func VerifyTag(node *yaml.Node, wantStr string, wantTag string) error {
	gotTag := node.ShortTag()
	if gotTag != wantTag {
		return errutil.NewPosNode(fmt.Errorf("%w: expected %s, but was %s",
			ErrInvalidFieldType, wantStr, PrettyNodeTypeName(node)), node)
	}
	return nil
}

func VerifyKindAndTag(node *yaml.Node, wantStr string, wantKind yaml.Kind, wantTag string) error {
	if err := VerifyKind(node, wantStr, wantKind); err != nil {
		return err
	}
	return VerifyTag(node, wantStr, wantTag)
}
