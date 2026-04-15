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
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mitchellh/hashstructure/v2"
	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/runtime/trace"
	"github.com/sentrie-sh/sentrie/xerr"
)

func evalCall(ctx context.Context, ec *ExecutionContext, exec *executorImpl, p *index.Policy, t *ast.CallExpression) (response box.Value, traceNode *trace.Node, err error) {
	ctx, n, done := trace.New(ctx, t, "call", map[string]any{
		"target": t.Callee.String(),
		"args":   t.Arguments,
	})
	defer done()

	args := make([]box.Value, 0, len(t.Arguments))
	for _, a := range t.Arguments {
		v, child, err := eval(ctx, ec, exec, p, a)
		n.Attach(child)
		if err != nil {
			return box.Undefined(), n.SetErr(err), err
		}
		args = append(args, v)
	}

	if t.Memoized {
		for _, a := range args {
			if a.IsCallable() {
				err := fmt.Errorf("memoized call cannot take callable arguments")
				return box.Undefined(), n.SetErr(err), err
			}
		}
	}

	target, err := getTarget(ctx, ec, exec, p, t)
	if err != nil {
		return box.Undefined(), n.SetErr(err), err
	}

	// use a thin wrapper around the target to handle the caching
	wrappedTarget := func(ctx context.Context, args ...box.Value) (box.Value, error) {
		if !t.Memoized {
			return target(ctx, args...)
		}

		ttl := 5 * time.Minute // default to 5 minutes
		if t.MemoizeTTL != nil {
			ttl = *t.MemoizeTTL
		}

		hashKey := calculateHashKey(t, args)
		loader := func(ctx context.Context, key string) (any, error) {
			return target(ctx, args...)
		}
		out, _, err := exec.callMemoizePerch.Get(ctx, hashKey, ttl, loader)
		if err != nil {
			return box.Undefined(), err
		}
		v, ok := out.(box.Value)
		if ok {
			return v, nil
		}
		return box.FromAny(out), nil
	}

	// call the target
	out, err := wrappedTarget(ctx, args...)
	if err != nil {
		if errors.Is(err, xerr.InjectedError{}) {
			// if this error is injected from code, we revert to the error message
			return box.Undefined(), n.SetErr(err), err
		}
		err = fmt.Errorf("failed to call function '%s': %w", t.Callee.String(), err)
		return box.Undefined(), n.SetErr(err), err
	}
	return out, n.SetResult(out), nil
}

// Helper to split "alias.fn" if ever needed
func splitAliasFn(s string) (string, string) {
	parts := strings.SplitN(s, ".", 2)
	if len(parts) != 2 {
		return s, ""
	}
	return parts[0], parts[1]
}

func calculateHashKey(node *ast.CallExpression, args []box.Value) string {
	hashArgs := make([]any, 0, len(args))
	for _, a := range args {
		// Callables are rejected for memoized calls before we get here.
		h, err := box.TryToBoundaryAny(a)
		if err != nil {
			return ""
		}
		hashArgs = append(hashArgs, h)
	}
	arghash, err := hashstructure.Hash(hashArgs, hashstructure.FormatV2, nil)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%p:%d", node, arghash)
}

func getTarget(_ context.Context, ec *ExecutionContext, exec *executorImpl, p *index.Policy, c *ast.CallExpression) (func(context.Context, ...box.Value) (box.Value, error), error) {
	callee := c.Callee.String()

	if builtin, ok := Builtins[callee]; ok {
		return func(ctx context.Context, args ...box.Value) (box.Value, error) {
			site := &CallSite{EC: ec, Exec: exec, Policy: p}
			return builtin(ctx, site, args...)
		}, nil
	}

	module, fn := splitAliasFn(callee)

	if module == "" || fn == "" {
		e := xerr.ErrImportResolution(module, p.Namespace.FQN.String())
		return nil, e
	}

	modulebinding, ok := ec.Module(module)
	if !ok {
		e := xerr.ErrModuleInvocation(module, fn)
		return nil, e
	}

	return func(ctx context.Context, args ...box.Value) (box.Value, error) {
		anyArgs := make([]any, 0, len(args))
		for _, a := range args {
			if a.IsCallable() {
				return box.Undefined(), fmt.Errorf("cannot pass callable value to module function %s.%s", module, fn)
			}
			x, err := box.TryToBoundaryAny(a)
			if err != nil {
				return box.Undefined(), fmt.Errorf("module call %s.%s: %w", module, fn, err)
			}
			anyArgs = append(anyArgs, x)
		}
		out, err := modulebinding.Call(ctx, ec, fn, anyArgs...)
		return box.FromBoundaryAny(out), err
	}, nil
}
