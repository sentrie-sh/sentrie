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
// ref the index holds (#273): facts, policy lets, shape aliases and fields,
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
	check := func(t ast.TypeRef) { errs = append(errs, typeRefConstraintArgErrors(t)...) }

	shapes := map[*ast.ShapeStatement]struct{}{}
	checkShape := func(s *Shape) {
		if s == nil || s.Statement == nil {
			return
		}
		if _, seen := shapes[s.Statement]; seen {
			return
		}
		shapes[s.Statement] = struct{}{}
		check(s.Statement.Simple)
		if s.Statement.Complex != nil {
			for _, f := range s.Statement.Complex.Fields {
				if f != nil {
					check(f.Type)
				}
			}
		}
	}

	for _, ns := range idx.Namespaces {
		for _, s := range ns.Shapes {
			checkShape(s)
		}
		for _, policy := range ns.Policies {
			for _, s := range policy.Shapes {
				checkShape(s)
			}
			for _, fact := range policy.Facts {
				check(fact.Type)
			}
			for _, let := range policy.Lets {
				check(let.Type)
			}
		}
	}
	for _, d := range idx.DerivesByFQN {
		if d.Lambda != nil {
			checkLambdaTypes(d.Lambda, check)
		}
	}

	err := idx.forEachScopedExpr(ctx, func(_ *kindCheckCtx, e ast.Expression) {
		switch n := e.(type) {
		case *ast.CastExpression:
			check(n.TargetType)
		case *ast.IsExpression:
			check(n.Type)
		case *ast.LambdaExpression:
			checkLambdaTypes(n, check)
		case *ast.BlockExpression:
			for _, st := range n.Statements {
				if vd, ok := st.(*ast.VarDeclaration); ok {
					check(vd.Type)
				}
			}
		}
	})
	if err != nil {
		return err
	}

	// Maps are walked in no fixed order; sort so the report is stable.
	slices.SortFunc(errs, func(a, b error) int { return cmp.Compare(a.Error(), b.Error()) })
	return errors.Join(errs...)
}

func checkLambdaTypes(lam *ast.LambdaExpression, check func(ast.TypeRef)) {
	for _, t := range lam.ParamTypes {
		check(t)
	}
	check(lam.ReturnType)
}

// typeRefConstraintArgErrors checks the constraints of t and of every type
// ref nested in it.
func typeRefConstraintArgErrors(t ast.TypeRef) []error {
	if t == nil {
		return nil
	}
	t = ast.UnwrapNullableTypeRef(t)

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

	switch n := t.(type) {
	case *ast.ListTypeRef:
		errs = append(errs, typeRefConstraintArgErrors(n.ElemType)...)
	case *ast.DictTypeRef:
		errs = append(errs, typeRefConstraintArgErrors(n.ValueType)...)
	case *ast.RecordTypeRef:
		for _, f := range n.Fields {
			errs = append(errs, typeRefConstraintArgErrors(f)...)
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
