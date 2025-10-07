package runtime

import (
	"context"
	"fmt"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/constraints"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/tokens"
)

func validateAgainstIntTypeRef(ctx context.Context, ec *ExecutionContext, exec Executor, p *index.Policy, val any, typeRef *ast.IntTypeRef, pos tokens.Position) error {
	if _, ok := val.(int64); !ok {
		return fmt.Errorf("value %v is not an int at %s - expected int", val, pos)
	}

	for _, constraint := range typeRef.GetConstraints() {
		args := make([]any, len(constraint.Args))
		for i, argExpr := range constraint.Args {
			csArg, _, err := eval(ctx, ec, exec.(*executorImpl), p, argExpr)
			if err != nil {
				return err
			}
			args[i] = csArg
		}

		checker, ok := constraints.IntContraintCheckers[constraint.Name]
		if !ok {
			return ErrUnknownConstraint(constraint)
		}

		if err := checker.Checker(ctx, p, val.(int64), args); err != nil {
			return ErrConstraintFailed(pos, constraint, err)
		}
	}

	return nil
}
