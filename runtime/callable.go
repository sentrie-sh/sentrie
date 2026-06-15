// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2026 Binaek Sarkar
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	return requiredLambdaArity(c.lambda)
}

func requiredLambdaArity(lam *ast.LambdaExpression) int {
	if lam == nil {
		return 0
	}
	n := len(lam.Params)
	required := 0
	for i := 0; i < n; i++ {
		opt := lam.ParamOpts != nil && i < len(lam.ParamOpts) && lam.ParamOpts[i]
		if !opt {
			required++
		}
	}
	return required
}

func (c *deriveCallable) Arity() int {
	if c == nil || c.derive == nil {
		return 0
	}
	return requiredLambdaArity(c.derive.Lambda)
}

func (c *lambdaCallable) Invoke(ctx context.Context, site *CallSite, args []box.Value) (box.Value, error) {
	args, err := padAndValidateCallableArgs(ctx, site.EC, site.Exec, site.Policy, c.lambda, args, "lambda")
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
		if err := validateValueAgainstTypeRef(ctx, child, site.Exec, site.Policy, v, c.lambda.ReturnType, c.lambda.Body.Span()); err != nil {
			return box.Undefined(), fmt.Errorf("lambda return: %w", err)
		}
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
