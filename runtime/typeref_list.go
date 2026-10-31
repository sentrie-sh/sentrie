// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"
	"fmt"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/constraints"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/tokens"
)

func validateAgainstListTypeRef(ctx context.Context, ec *ExecutionContext, exec Executor, p *index.Policy, v box.Value, typeRef *ast.ListTypeRef, pos tokens.Range) error {
	items, ok := v.ListValue()
	if !ok {
		return errMismatch("value %v is not an array at %s - expected array", v, pos)
	}

	for _, item := range items {
		if err := validateValueAgainstTypeRef(ctx, ec, exec, p, item, typeRef.ElemType, pos); err != nil {
			return fmt.Errorf("item is not valid at %s: %w", pos, err) // TODO: improve this error message
		}
	}

	for _, constraint := range typeRef.GetConstraints() {
		args := make([]box.Value, len(constraint.Args))
		for i, argExpr := range constraint.Args {
			csArg, _, err := eval(ctx, ec, exec.(*executorImpl), p, argExpr)
			if err != nil {
				return errCheckFailed(err)
			}
			args[i] = csArg
		}
		checker, ok := constraints.ListContraintCheckers[constraint.Name]
		if !ok {
			return ErrUnknownConstraint(constraint)
		}

		if err := checker.Checker(ctx, v, args); err != nil {
			return ErrConstraintFailed(pos, constraint, err)
		}
	}

	return nil
}
