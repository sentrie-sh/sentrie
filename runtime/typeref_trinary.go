// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/constraints"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/tokens"
	"github.com/sentrie-sh/sentrie/trinary"
)

func validateAgainstTrinaryTypeRef(ctx context.Context, ec *ExecutionContext, exec Executor, p *index.Policy, v box.Value, typeRef *ast.TrinaryTypeRef, valueRange tokens.Range) error {
	var tv trinary.Value
	if b, ok := v.BoolValue(); ok {
		tv = trinary.From(b)
	} else if t, ok := v.TrinaryValue(); ok {
		tv = t
	} else {
		return errMismatch("value '%v' is not a bool at %s - expected bool", v, valueRange)
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
		checker, ok := constraints.TrinaryConstraintCheckers[constraint.Name]
		if !ok {
			return ErrUnknownConstraint(constraint)
		}

		if err := checker.Checker(ctx, p, box.Trinary(tv), args); err != nil {
			return ErrConstraintFailed(valueRange, constraint, err)
		}
	}
	return nil
}
