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

func newLambdaCallable(lambda *ast.LambdaExpression, capture *ExecutionContext) *lambdaCallable {
	return &lambdaCallable{lambda: lambda, capture: capture}
}

func (c *lambdaCallable) Arity() int {
	return len(c.lambda.Params)
}

func (c *lambdaCallable) Invoke(ctx context.Context, site *CallSite, args []box.Value) (box.Value, error) {
	if len(args) != len(c.lambda.Params) {
		return box.Undefined(), fmt.Errorf("callable invoked with %d arguments, expected %d", len(args), len(c.lambda.Params))
	}
	child := c.capture.AttachedChildContext()
	defer child.Dispose()
	for i, name := range c.lambda.Params {
		child.SetLocal(name, args[i], true)
	}
	v, _, err := evalBlock(ctx, child, site.Exec, site.Policy, c.lambda.Body)
	return v, err
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

// invokeCallable invokes a boxed callable with args after arity check.
func invokeCallable(ctx context.Context, site *CallSite, v box.Value, args []box.Value) (box.Value, error) {
	c, err := callableFromValue(v)
	if err != nil {
		return box.Undefined(), err
	}
	if len(args) != c.Arity() {
		return box.Undefined(), fmt.Errorf("callable invoked with %d arguments, expected %d", len(args), c.Arity())
	}
	return c.Invoke(ctx, site, args)
}

// --- helpers for higher-order builtins (arity contract) ---

func iterArgs(_ *CallSite, c Callable, item box.Value, idx int) ([]box.Value, error) {
	switch c.Arity() {
	case 1:
		return []box.Value{item}, nil
	case 2:
		return []box.Value{item, box.Number(idx)}, nil
	default:
		return nil, fmt.Errorf("iterator callable must have arity 1 or 2, got %d", c.Arity())
	}
}

func reduceArgs(_ *CallSite, c Callable, acc, item box.Value, idx int) ([]box.Value, error) {
	switch c.Arity() {
	case 2:
		return []box.Value{acc, item}, nil
	case 3:
		return []box.Value{acc, item, box.Number(idx)}, nil
	default:
		return nil, fmt.Errorf("reducer callable must have arity 2 or 3, got %d", c.Arity())
	}
}
