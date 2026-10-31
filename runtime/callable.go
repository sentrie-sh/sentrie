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

// Callable is a boxed runtime callable (lambda closure). v1 capture keeps a
// reference to the defining ExecutionContext so late-bound lexical lookups use
// the live parent chain, not a snapshot at creation time.
type Callable interface {
	Arity() int
	Invoke(ctx context.Context, site *CallSite, args []box.Value) (box.Value, error)
}

type lambdaCallable struct {
	lambda  *ast.LambdaExpression
	capture *ExecutionContext
}

type deriveCallable struct {
	derive *index.Derive
}

func newLambdaCallable(lambda *ast.LambdaExpression, capture *ExecutionContext) *lambdaCallable {
	return &lambdaCallable{lambda: lambda, capture: capture}
}

func (c *lambdaCallable) Arity() int {
	return ast.RequiredLambdaArity(c.lambda)
}

func (c *deriveCallable) Arity() int {
	if c == nil || c.derive == nil {
		return 0
	}
	return ast.RequiredLambdaArity(c.derive.Lambda)
}

func (c *lambdaCallable) Invoke(ctx context.Context, site *CallSite, args []box.Value) (box.Value, error) {
	// Collection builtins apply a lambda once per element, so this is the
	// per-iteration deadline check for module-free evaluation (#225).
	if err := ctx.Err(); err != nil {
		return box.Undefined(), err
	}
	// The lambda's types resolve where it was written: in the enclosing
	// derive's scope for a lambda in a derive body, else in the policy.
	scope := typeCheckScope(c.capture, site.Policy)
	args, err := padAndValidateCallableArgs(ctx, site.EC, site.Exec, scope, c.lambda, args, "lambda")
	if err != nil {
		return box.Undefined(), err
	}

	child := c.capture.AttachedChildContext()
	defer child.Dispose()
	for i, name := range c.lambda.Params {
		child.SetLocal(name, args[i], true)
	}
	v, _, err := evalBlock(ctx, child, site.Exec, site.Policy, c.lambda.Body)
	if err != nil {
		return box.Undefined(), err
	}
	if c.lambda.ReturnType != nil {
		if err := validateValueAgainstTypeRef(ctx, child, site.Exec, scope, v, c.lambda.ReturnType, c.lambda.Body.Span()); err != nil {
			return box.Undefined(), fmt.Errorf("lambda return: %w", err)
		}
		v = bindTruthValue(v, c.lambda.ReturnType)
	}
	return v, nil
}

func (c *deriveCallable) Invoke(ctx context.Context, site *CallSite, args []box.Value) (box.Value, error) {
	if c == nil || c.derive == nil {
		return box.Undefined(), fmt.Errorf("internal error: missing derive")
	}
	return invokeDerive(ctx, site.EC, site.Exec, site.Policy, c.derive, args)
}

// callableFromValue unwraps a boxed callable.
func callableFromValue(v box.Value) (Callable, error) {
	ref, ok := v.CallableRef()
	if !ok {
		return nil, fmt.Errorf("expected callable, got %s", v.Kind())
	}
	c, ok := ref.(Callable)
	if !ok {
		return nil, fmt.Errorf("internal error: callable payload is %T", ref)
	}
	return c, nil
}
