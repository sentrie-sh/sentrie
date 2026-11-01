// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"cmp"
	"context"
	"errors"
	"slices"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/constraints"
	"github.com/sentrie-sh/sentrie/xerr"
)

// checkConstraintArgs checks the arguments of every constraint on every type
// ref forEachTypeRef visits (#273): facts, policy lets, shape aliases and fields,
// derive parameter and return types, and inside expressions block lets,
// lambda parameter and return types, casts and `x is T`. Element, value and
// field types of list, dict and record refs are checked too. The check is the
// definition's own CheckArgs, the same one the runtime makes before it looks
// at a value, so the two cannot disagree.
//
// The parser has already rejected an unknown constraint name or a wrong
// argument count through ast's generated tables; an unknown name is skipped
// here (#113), and an argument that is not a literal (the parser accepts only
// literals) is not checked.
func (idx *Index) checkConstraintArgs(ctx context.Context) error {
	var errs []error
	err := idx.forEachTypeRef(ctx, func(_ typeRefSite, t ast.TypeRef) {
		errs = append(errs, typeRefConstraintArgErrors(t)...)
	})
	if err != nil {
		return err
	}

	// Maps are walked in no fixed order; sort so the report is stable.
	slices.SortFunc(errs, func(a, b error) int { return cmp.Compare(a.Error(), b.Error()) })
	return errors.Join(errs...)
}

// typeRefConstraintArgErrors checks the constraints of t itself;
// forEachTypeRef supplies the type refs nested in it.
func typeRefConstraintArgErrors(t ast.TypeRef) []error {
	var errs []error
	table := constraints.ForTypeRef(t)
	for _, c := range t.GetConstraints() {
		def, ok := table[c.Name]
		if !ok {
			continue
		}
		args, ok := constraintLiteralArgs(c.Args)
		if !ok {
			continue
		}
		if err := def.CheckArgs(args); err != nil {
			errs = append(errs, xerr.ErrConstraintArg(c.Span(), c.Name, err))
		}
	}
	return errs
}

// constraintLiteralArgs folds constraint argument expressions to the values
// the runtime would evaluate them to. It reports false when any argument is
// not a literal.
func constraintLiteralArgs(exprs []ast.Expression) ([]box.Value, bool) {
	args := make([]box.Value, len(exprs))
	for i, e := range exprs {
		v, ok := constraintLiteral(e)
		if !ok {
			return nil, false
		}
		args[i] = v
	}
	return args, true
}

func constraintLiteral(e ast.Expression) (box.Value, bool) {
	switch n := e.(type) {
	case *ast.StringLiteral:
		return box.String(n.Value), true
	case *ast.IntegerLiteral:
		return box.Number(n.Value), true
	case *ast.FloatLiteral:
		return box.Number(n.Value), true
	case *ast.TrinaryLiteral:
		return box.Trinary(n.Value), true
	case *ast.NullLiteral:
		return box.Null(), true
	case *ast.ListLiteral:
		items, ok := constraintLiteralArgs(n.Values)
		if !ok {
			return box.Undefined(), false
		}
		return box.List(items), true
	case *ast.MapLiteral:
		m := make(map[string]box.Value, len(n.Entries))
		for _, entry := range n.Entries {
			key, ok := entry.Key.(*ast.StringLiteral)
			if !ok {
				return box.Undefined(), false
			}
			v, ok := constraintLiteral(entry.Value)
			if !ok {
				return box.Undefined(), false
			}
			m[key.Value] = v
		}
		return box.Dict(m), true
	}
	return box.Undefined(), false
}
