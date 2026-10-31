// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/runtime/trace"
)

// evalIsType evaluates `x is T`: true when x validates against T, false when it
// does not or x is undefined. Only an error raised while making the check (an
// unresolvable shape, a constraint whose own arguments are invalid, a
// constraint argument that fails to evaluate) is returned as an error.
//
// When the index resolved the right-hand side to a value, the expression is
// the legacy equality, evaluated as the infix `is` (==) is.
//
// Inside a derive body (ec.evalDerive != nil) the check is allowed: it reads
// only the operand, the shape declarations and the constraint arguments, and
// the operand and arguments resolve through eval, whose identifier and call
// paths apply the derive restrictions. Shapes resolve against the derive's own
// scope, as the index resolved them, not against the calling policy.
func evalIsType(ctx context.Context, ec *ExecutionContext, exec *executorImpl, p *index.Policy, is *ast.IsExpression) (box.Value, *trace.Node, error) {
	scope := typeCheckScope(ec, p)
	typ := is.Type
	if is.Ambiguous() {
		// The index settles this during validation; an index that was never
		// validated gets the same shape lookup, without the scope checks.
		var err error
		if typ, err = resolveUnvalidatedIsType(exec, scope, typ); err != nil {
			_, node, done := trace.New(ctx, is, "is", nil)
			defer done()
			return box.Undefined(), node.SetErr(err), err
		}
	}
	if typ == nil {
		return evalIsEquality(ctx, ec, exec, p, is)
	}

	ctx, node, done := trace.New(ctx, is, "is", map[string]any{
		"type": typ.String(),
	})
	defer done()

	val, child, err := eval(ctx, ec, exec, p, is.Left)
	node.Attach(child)
	if err != nil {
		return box.Undefined(), node.SetErr(err), err
	}

	if val.IsUndefined() {
		out := box.Bool(false)
		return out, node.SetResult(out), nil
	}

	if err := validateValueAgainstTypeRef(ctx, ec, exec, scope, val, typ, is.Left.Span()); err != nil {
		if !isTypeMismatch(err) {
			return box.Undefined(), node.SetErr(err), err
		}
		out := box.Bool(false)
		return out, node.SetResult(out), nil
	}

	out := box.Bool(true)
	return out, node.SetResult(out), nil
}

// evalIsEquality is `x is y` with a value on the right: the same operation as
// the infix `is` in evalInfix (undefined if either side is undefined, else
// box.EqualValues), kept here so the equality path allocates no InfixExpression.
func evalIsEquality(ctx context.Context, ec *ExecutionContext, exec *executorImpl, p *index.Policy, is *ast.IsExpression) (box.Value, *trace.Node, error) {
	ctx, node, done := trace.New(ctx, is, "infix", map[string]any{
		"operator": "is",
	})
	defer done()

	l, ln, err := eval(ctx, ec, exec, p, is.Left)
	node.Attach(ln)
	if err != nil {
		return box.Undefined(), node.SetErr(err), err
	}
	r, rn, err := eval(ctx, ec, exec, p, is.Right)
	node.Attach(rn)
	if err != nil {
		return box.Undefined(), node.SetErr(err), err
	}
	if l.IsUndefined() || r.IsUndefined() {
		return box.Undefined(), node.SetResult(box.Undefined()), nil
	}
	out := box.Bool(box.EqualValues(l, r))
	return out, node.SetResult(out), nil
}

// resolveUnvalidatedIsType returns typ when its name is a shape visible from p
// and nil when it is not, so the expression falls back to equality.
func resolveUnvalidatedIsType(exec *executorImpl, p *index.Policy, typ ast.TypeRef) (ast.TypeRef, error) {
	ref, ok := typ.(*ast.ShapeTypeRef)
	if !ok {
		return typ, nil
	}
	var ns *index.Namespace
	if p != nil {
		ns = p.Namespace
	}
	shape, err := exec.Index().LookupShape(p, ns, ref.Ref)
	if err != nil {
		return nil, err
	}
	if shape == nil {
		return nil, nil
	}
	return typ, nil
}
