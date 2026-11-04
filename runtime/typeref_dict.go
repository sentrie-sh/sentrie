// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"
	"fmt"
	"maps"
	"slices"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/constraints"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/tokens"
)

func validateAgainstDictTypeRef(ctx context.Context, ec *ExecutionContext, exec Executor, p *index.Policy, v box.Value, typeRef *ast.DictTypeRef, pos tokens.Range) error {
	entries, ok := v.DictValue()
	if !ok {
		return errMismatch("value %v is not a dict", v)
	}

	// check every value against the element type, in key order so the
	// reported key is the same on every run
	for _, key := range slices.Sorted(maps.Keys(entries)) {
		if err := validateValueAgainstTypeRef(ctx, ec, exec, p, entries[key], typeRef.ValueType, pos); err != nil {
			return fmt.Errorf("value for key %q is not valid at %s: %w", key, pos, err)
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
		checker, ok := constraints.DictContraintCheckers[constraint.Name]
		if !ok {
			return ErrUnknownConstraint(constraint)
		}

		if err := checker.Checker(ctx, p, v, args); err != nil {
			return ErrConstraintFailed(pos, constraint, err)
		}
	}

	return nil
}
