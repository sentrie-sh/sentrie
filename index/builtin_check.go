// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/builtins"
	"github.com/sentrie-sh/sentrie/xerr"
)

func (idx *Index) checkBuiltinCalls(ctx context.Context) error {
	var errs []error

	for _, ns := range idx.Namespaces {
		if ctx.Err() != nil {
			return fmt.Errorf("validation cancelled: %w", xerr.ErrIndex)
		}
		for _, policy := range ns.Policies {
			if ctx.Err() != nil {
				return fmt.Errorf("validation cancelled: %w", xerr.ErrIndex)
			}
			k := newRuleKindCheckCtx(idx, policy)
			for _, rule := range policy.Rules {
				errs = append(errs, k.walkExpr(rule.Default)...)
				errs = append(errs, k.walkExpr(rule.When)...)
				errs = append(errs, k.walkExpr(rule.Body)...)
			}
		}
	}

	for _, d := range idx.DerivesByFQN {
		if ctx.Err() != nil {
			return fmt.Errorf("validation cancelled: %w", xerr.ErrIndex)
		}
		k := newDeriveKindCheckCtx(idx, d)
		if d.Lambda != nil && d.Lambda.Body != nil {
			errs = append(errs, k.walkExpr(d.Lambda.Body)...)
		}
	}

	return errors.Join(errs...)
}

func (k *kindCheckCtx) walkExpr(e ast.Expression) []error {
	if e == nil {
		return nil
	}

	var errs []error

	switch n := e.(type) {
	case *ast.CallExpression:
		errs = append(errs, k.checkBuiltinCall(n)...)
		errs = append(errs, k.walkExpr(n.Callee)...)
		for _, a := range n.Arguments {
			errs = append(errs, k.walkExpr(a)...)
		}
		return errs
	case *ast.BlockExpression:
		child := *k
		child.scope = cloneBindingScope(k.scope)
		for _, st := range n.Statements {
			vd, ok := st.(*ast.VarDeclaration)
			if !ok {
				continue
			}
			errs = append(errs, child.walkExpr(vd.Value)...)
			child.bindLet(vd)
		}
		errs = append(errs, child.walkExpr(n.Yield)...)
		return errs
	case *ast.LambdaExpression:
		child := k.pushLambdaScope(n)
		errs = append(errs, child.walkExpr(n.Body)...)
		return errs
	default:
		_ = forEachDeriveExprChild(e, func(child ast.Expression) error {
			errs = append(errs, k.walkExpr(child)...)
			return nil
		})
		return errs
	}
}

func (k *kindCheckCtx) checkBuiltinCall(c *ast.CallExpression) []error {
	decl, ok := k.isBuiltinCall(c)
	if !ok {
		return nil
	}

	var errs []error
	sig := decl.Sig

	min := 0
	for _, p := range sig.Params {
		if !p.Optional {
			min++
		}
	}
	max := len(sig.Params)

	if sig.Variadic == nil {
		if len(c.Arguments) > max {
			msg := sig.TooManyError
			if msg == "" {
				msg = sig.TooFewError
			}
			at := c.Arguments[max].Span()
			errs = append(errs, xerr.ErrBuiltinCallArity(at, msg))
		}
	}

	if len(c.Arguments) < min {
		at := c.Span()
		if id, ok := c.Callee.(*ast.Identifier); ok {
			at = id.Span()
		}
		errs = append(errs, xerr.ErrBuiltinCallArity(at, sig.TooFewError))
	}

	for i, arg := range c.Arguments {
		ps, ok := builtinParamSigAt(sig, i)
		if !ok {
			break
		}

		if len(ps.Kinds) > 0 && ps.OnMismatch != builtins.MismatchUndefined {
			if kind, known := k.resolveKind(arg); known && !kindAllowed(ps.Kinds, kind) {
				msg := ps.KindError
				if msg == "" && slices.Contains(ps.Kinds, box.ValueCallable) && kind != box.ValueCallable {
					msg = fmt.Sprintf("expected callable, got %s", kind.String())
				}
				if msg != "" {
					errs = append(errs, xerr.ErrBuiltinArgKind(arg.Span(), msg))
				}
			}
		}

		if len(ps.CallableArities) > 0 {
			if required, known := k.resolveCallableArity(arg); known && !slices.Contains(ps.CallableArities, required) {
				errs = append(errs, xerr.ErrBuiltinCallableArity(arg.Span(), ps.CallableArityError))
			}
		}
	}

	return errs
}
