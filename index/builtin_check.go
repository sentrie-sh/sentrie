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

// checkBuiltinCalls statically checks the arity and argument kinds of every
// native builtin call the index holds. It walks the same expression sites as
// the nondeterminism and capabilities checks, through forEachBuiltinCall, so
// a new expression site is covered by all three at once (#218). The runtime
// makes the same checks when the call is evaluated, from the same signature.
func (idx *Index) checkBuiltinCalls(ctx context.Context) error {
	var errs []error
	err := idx.forEachBuiltinCall(ctx, func(k *kindCheckCtx, decl *builtins.Decl, call *ast.CallExpression) {
		errs = append(errs, k.checkBuiltinCall(decl, call)...)
	})
	if err != nil {
		return err
	}
	return errors.Join(errs...)
}

func (k *kindCheckCtx) checkBuiltinCall(decl *builtins.Decl, c *ast.CallExpression) []error {
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
		if len(c.Arguments) > max && sig.TooManyError != "" {
			at := c.Arguments[max].Span()
			errs = append(errs, xerr.ErrBuiltinCallArity(at, sig.TooManyError))
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
