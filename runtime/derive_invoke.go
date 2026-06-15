// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"
	"fmt"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/index"
)

func padAndValidateDeriveArgs(ctx context.Context, ec *ExecutionContext, exec *executorImpl, p *index.Policy, lam *ast.LambdaExpression, args []box.Value) ([]box.Value, error) {
	return padAndValidateCallableArgs(ctx, ec, exec, p, lam, args, "derive")
}

// padAndValidateCallableArgs pads optional parameters with undefined and validates typed
// parameters for derive and anonymous lambda invocation.
func padAndValidateCallableArgs(ctx context.Context, ec *ExecutionContext, exec *executorImpl, p *index.Policy, lam *ast.LambdaExpression, args []box.Value, argKind string) ([]box.Value, error) {
	n := len(lam.Params)
	if len(args) > n {
		return nil, fmt.Errorf("too many arguments: want at most %d, got %d", n, len(args))
	}
	required := ast.RequiredLambdaArity(lam)
	if len(args) < required {
		return nil, fmt.Errorf("not enough arguments: want at least %d, got %d", required, len(args))
	}
	out := make([]box.Value, n)
	copy(out, args)
	for i := len(args); i < n; i++ {
		out[i] = box.Undefined()
	}
	for i := range n {
		if lam.ParamTypes == nil || i >= len(lam.ParamTypes) || lam.ParamTypes[i] == nil {
			continue
		}
		opt := lam.ParamOpts != nil && i < len(lam.ParamOpts) && lam.ParamOpts[i]
		if opt && out[i].IsUndefined() {
			continue
		}
		if err := validateValueAgainstTypeRef(ctx, ec, exec, p, out[i], lam.ParamTypes[i], lam.Span()); err != nil {
			name := fmt.Sprintf("argument %d", i)
			if i < len(lam.Params) {
				name = lam.Params[i]
			}
			return nil, fmt.Errorf("%s argument %q: %w", argKind, name, err)
		}
	}
	return out, nil
}

func invokeDerive(ctx context.Context, callerEC *ExecutionContext, exec *executorImpl, callerPolicy *index.Policy, d *index.Derive, args []box.Value) (box.Value, error) {
	lam := d.Lambda
	args, err := padAndValidateDeriveArgs(ctx, callerEC, exec, callerPolicy, lam, args)
	if err != nil {
		return box.Undefined(), err
	}

	child := callerEC.DetachedChildContext()
	defer child.Dispose()

	child.evalDerive = d

	if err := child.PushRefStack(d.FQN.String()); err != nil {
		return box.Undefined(), err
	}
	defer child.PopRefStack()

	for i, name := range lam.Params {
		child.SetLocal(name, args[i], true)
	}

	val, _, err := evalBlock(ctx, child, exec, callerPolicy, lam.Body)
	if err != nil {
		return box.Undefined(), err
	}
	if val.IsCallable() {
		return box.Undefined(), fmt.Errorf("derive cannot yield a callable value")
	}
	if lam.ReturnType != nil {
		if err := validateValueAgainstTypeRef(ctx, child, exec, callerPolicy, val, lam.ReturnType, lam.Body.Span()); err != nil {
			return box.Undefined(), fmt.Errorf("derive return: %w", err)
		}
	}
	return val, nil
}
