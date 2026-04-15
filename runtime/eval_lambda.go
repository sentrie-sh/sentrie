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

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/runtime/trace"
)

func evalLambda(ctx context.Context, ec *ExecutionContext, _ *executorImpl, _ *index.Policy, lam *ast.LambdaExpression) (box.Value, *trace.Node, error) {
	_, n, done := trace.New(ctx, lam, "lambda", map[string]any{
		"params": lam.Params,
	})
	defer done()

	// v1: capture the current execution context by reference (not a value snapshot).
	fn := newLambdaCallable(lam, ec)
	out := box.Callable(fn)
	return out, n.SetResult(out), nil
}
