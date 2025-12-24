// SPDX-License-Identifier: Apache-2.0
//
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

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/runtime/trace"
	"github.com/sentrie-sh/sentrie/trinary"
)

func evalTernary(ctx context.Context, ec *ExecutionContext, exec *executorImpl, p *index.Policy, t *ast.TernaryExpression) (any, *trace.Node, error) {
	ctx, n, done := trace.New(ctx, t, "ternary", map[string]any{})
	defer done()

	c, cn, err := eval(ctx, ec, exec, p, t.Condition)
	n.Attach(cn)
	if err != nil {
		return nil, n.SetErr(err), err
	}
	if trinary.From(c).IsTrue() {
		v, tn, err := eval(ctx, ec, exec, p, t.ThenBranch)
		n.Attach(tn)
		return v, n, err
	}
	v, en, err := eval(ctx, ec, exec, p, t.ElseBranch)
	n.Attach(en)
	return v, n, err
}
