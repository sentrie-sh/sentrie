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

// validateAgainstSetTypeRef checks v against `set[T]` (#162). A set conforms
// when its elements conform to T. A plain list conforms when, in addition, its
// elements can be held in a set and are distinct: a JSON array reaching a
// set[T] position is validated, never silently deduplicated. A duplicate is a
// mismatch, so `[1, 1] is set[number]` is false. A set[T] position takes the
// list constraints.
func validateAgainstSetTypeRef(ctx context.Context, ec *ExecutionContext, exec Executor, p *index.Policy, v box.Value, typeRef *ast.SetTypeRef, pos tokens.Range) error {
	items, ok := v.ListValue()
	if !ok {
		return errMismatch("value %v is not a set at %s - expected set", v, pos)
	}

	for _, item := range items {
		if err := validateValueAgainstTypeRef(ctx, ec, exec, p, item, typeRef.ElemType, pos); err != nil {
			return fmt.Errorf("item is not valid at %s: %w", pos, err)
		}
	}

	if !v.IsSet() {
		if _, err := box.NewSetUnique(setElementsAs(items, typeRef.ElemType)); err != nil {
			return errMismatch("value %v is not a valid set at %s: %s", v, pos, err)
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

		if err := checker.Checker(ctx, p, v, args); err != nil {
			return ErrConstraintFailed(pos, constraint, err)
		}
	}

	return nil
}

// bindSetValue returns the value a binding declared as typeRef holds: a list
// bound to a set[T] (or set[T]?) binding becomes a set, so it serialises in
// sorted order and reports true for `is set[T]` by tag, not by scan. Only the
// top-level type is considered; a set[T] nested in a list, dict, record or
// shape field is validated but keeps the list it arrived as. A list that cannot
// be a set (a duplicate or an unsupported element) is returned unchanged, for
// validation to report.
func bindSetValue(v box.Value, typeRef ast.TypeRef) box.Value {
	if typeRef == nil || v.IsSet() {
		return v
	}
	if ast.IsNullableTypeRef(typeRef) {
		typeRef = ast.UnwrapNullableTypeRef(typeRef)
	}
	st, ok := typeRef.(*ast.SetTypeRef)
	if !ok {
		return v
	}
	xs, ok := v.ListValue()
	if !ok {
		return v
	}
	s, err := box.NewSetUnique(setElementsAs(xs, st.ElemType))
	if err != nil {
		return v
	}
	return s
}

// setElementsAs returns the elements a set[elemType] holds: each element is
// bound as elemType is, so an imported decision in a set[boolean] is its state
// (#136), not the dict NewSet would otherwise keep, and an unknown state is
// then rejected, since a set cannot hold unknown. Validation checks the same
// elements, so `[impT, true]` is a duplicate in a set[boolean].
func setElementsAs(xs []box.Value, elemType ast.TypeRef) []box.Value {
	out := make([]box.Value, len(xs))
	for i, x := range xs {
		out[i] = bindTruthValue(x, elemType)
	}
	return out
}

// bindTypedValue applies every conversion a binding declared as typeRef makes
// to the value it holds: an imported decision bound as a truth value holds its
// state (bindTruthValue, #136), and a list bound as a set is a set
// (bindSetValue, #162).
func bindTypedValue(v box.Value, typeRef ast.TypeRef) box.Value {
	return bindSetValue(bindTruthValue(v, typeRef), typeRef)
}
