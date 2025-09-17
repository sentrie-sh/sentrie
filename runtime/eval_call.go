// Copyright 2025 Binaek Sarkar
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
	"strings"
	"time"

	"github.com/binaek/sentra/ast"
	"github.com/binaek/sentra/index"
	"github.com/binaek/sentra/runtime/trace"
	"github.com/binaek/sentra/xerr"
	"github.com/mitchellh/hashstructure/v2"
)

func evalCall(ctx context.Context, ec *ExecutionContext, exec *executorImpl, p *index.Policy, c *ast.CallExpression) (response any, traceNode *trace.Node, err error) {
	n, done := trace.New("call", "", c, map[string]any{
		"target": c.Callee.String(),
		"args":   c.Arguments,
	})
	defer done()

	args := make([]any, 0, len(c.Arguments))
	for _, a := range c.Arguments {
		v, child, err := eval(ctx, ec, exec, p, a)
		n.Attach(child)
		if err != nil {
			return nil, n.SetErr(err), err
		}
		args = append(args, v)
	}

	target, err := getTarget(ctx, ec, p, c)
	if err != nil {
		return nil, n.SetErr(err), err
	}

	// use a thin wrapper around the target to handle the caching
	wrappedTarget := func(ctx context.Context, args ...any) (any, error) {
		if !c.Memoized {
			// no memoization, so we can just call the target
			// quickly call the target without caching
			return target(ctx, args...)
		}

		ttl := 5 * time.Minute // default to 5 minutes
		if c.MemoizeTTL != nil {
			ttl = *c.MemoizeTTL
		}

		hashKey := calculateHashKey(c, args)
		loader := func(ctx context.Context, key string) (any, error) {
			return target(ctx, args...)
		}
		return exec.callMemoizePerch.Get(ctx, hashKey, ttl, loader)
	}

	// call the target
	out, err := wrappedTarget(ctx, args...)
	return out, n.SetResult(out).SetErr(err), err
}

// Helper to split "alias.fn" if ever needed
func splitAliasFn(s string) (string, string) {
	parts := strings.SplitN(s, ".", 2)
	if len(parts) != 2 {
		return s, ""
	}
	return parts[0], parts[1]
}

func calculateHashKey(node *ast.CallExpression, args []any) string {
	arghash, err := hashstructure.Hash(args, hashstructure.FormatV2, nil)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%p:%d", node, arghash)
}

func getTarget(_ context.Context, ec *ExecutionContext, p *index.Policy, c *ast.CallExpression) (func(context.Context, ...any) (any, error), error) {
	callee := c.Callee.String()

	// check if we have a builtin function
	if builtin, ok := Builtins[callee]; ok {
		return func(ctx context.Context, args ...any) (any, error) {
			return builtin(ctx, args)
		}, nil
	}

	// otherwise, assume that's a module function
	module, fn := splitAliasFn(callee)

	// if the module or fn are empty, it's a problem
	if module == "" || fn == "" {
		e := xerr.ErrImportResolution(module, p.Namespace.FQN.String())
		return nil, e
	}

	modulebinding, ok := ec.Module(module)
	if !ok {
		e := xerr.ErrModuleInvocation(module, fn)
		return nil, e
	}

	return func(ctx context.Context, args ...any) (any, error) {
		return modulebinding.Call(ctx, fn, args...)
	}, nil
}
