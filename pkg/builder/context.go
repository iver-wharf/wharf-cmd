package builder

import (
	"context"
	"fmt"
)

type contextKey int

const (
	contextKeyStageName contextKey = iota
	contextKeyStepName
)

func contextWithStageName(ctx context.Context, stage string) context.Context {
	return context.WithValue(ctx, contextKeyStageName, stage)
}

func contextStageName(ctx context.Context) (string, bool) {
	if v := ctx.Value(contextKeyStageName); v != nil {
		return v.(string), true
	}
	return "", false
}

func contextWithStepName(ctx context.Context, stage string) context.Context {
	return context.WithValue(ctx, contextKeyStepName, stage)
}

func contextStepName(ctx context.Context) (string, bool) {
	if v := ctx.Value(contextKeyStepName); v != nil {
		return v.(string), true
	}
	return "", false
}

func contextStageStepName(ctx context.Context) string {
	stage, hasStage := contextStageName(ctx)
	step, hasStep := contextStepName(ctx)
	switch {
	case hasStage && hasStep:
		return fmt.Sprintf("%s/%s", stage, step)
	case hasStage:
		return stage
	case hasStep:
		return step
	default:
		return ""
	}
}
