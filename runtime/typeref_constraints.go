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
)

// boundConstraint is a constraint of a type ref with its evaluated arguments,
// which have passed the definition's argument check.
type boundConstraint struct {
	constraint *ast.TypeRefConstraint
	def        constraints.ConstraintDefinition
	args       []box.Value
}

// bindConstraints evaluates the arguments of typeRef's own constraints and
// checks them against each definition (CheckArgs), without looking at a
// value. Every typeref validator calls it before the value's kind check, and
// validateValueAgainstTypeRef before a nullable type's null short-circuit, so
// a broken constraint argument is an error for every value rather than only
// for values of the constraint's own kind (#273). A failure is never a type
// mismatch, so `x is not T` propagates it instead of returning true.
func bindConstraints(ctx context.Context, ec *ExecutionContext, exec Executor, p *index.Policy, typeRef ast.TypeRef) ([]boundConstraint, error) {
	cs := typeRef.GetConstraints()
	if len(cs) == 0 {
		return nil, nil
	}
	table := constraints.ForTypeRef(typeRef)
	bound := make([]boundConstraint, 0, len(cs))
	for _, constraint := range cs {
		args := make([]box.Value, len(constraint.Args))
		for i, argExpr := range constraint.Args {
			csArg, _, err := eval(ctx, ec, exec.(*executorImpl), p, argExpr)
			if err != nil {
				return nil, errCheckFailed(err)
			}
			args[i] = csArg
		}
		def, ok := table[constraint.Name]
		if !ok {
			return nil, ErrUnknownConstraint(constraint)
		}
		if err := def.CheckArgs(args); err != nil {
			return nil, ErrConstraintFailed(constraint.Span(), constraint, err)
		}
		bound = append(bound, boundConstraint{constraint: constraint, def: def, args: args})
	}
	return bound, nil
}

// checkConstraints runs bound constraints against v in declaration order,
// stopping at the first failure.
func checkConstraints(ctx context.Context, v box.Value, bound []boundConstraint, pos tokens.Range) error {
	for _, b := range bound {
		if err := b.def.Checker(ctx, v, b.args); err != nil {
			return ErrConstraintFailed(pos, b.constraint, err)
		}
	}
	return nil
}
