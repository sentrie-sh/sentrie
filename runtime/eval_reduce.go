// SPDX-License-Identifier: Apache-2.0

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

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/runtime/trace"
)

func evalReduce(ctx context.Context, ec *ExecutionContext, exec *executorImpl, p *index.Policy, r *ast.ReduceExpression) (any, *trace.Node, error) {
	ctx, node, done := trace.New(ctx, r, "reduce", map[string]any{
		"collection":  r.Collection,
		"from":        r.From,
		"value_iter":  r.ValueIterator,
		"index_iter":  r.IndexIterator,
		"accumulator": r.Accumulator,
		"reducer":     r.Reducer,
	})
	defer done()

	col, colNode, err := eval(ctx, ec, exec, p, r.Collection)
	node.Attach(colNode)
	if err != nil {
		return nil, node.SetErr(err), err
	}

	if IsUndefined(col) {
		return Undefined, node, nil
	}

	list, ok := col.([]any)
	if !ok {
		return nil, node.SetErr(fmt.Errorf("filter expects list source")), fmt.Errorf("filter expects list source")
	}

	accumulator, accumulatorNode, err := eval(ctx, ec, exec, p, r.From)
	if err != nil {
		return nil, node.SetErr(err), err
	}
	node.Attach(accumulatorNode)

	for idx, item := range list {
		childContext := ec.AttachedChildContext()
		childContext.SetLocal(r.ValueIterator, item, true)
		childContext.SetLocal(r.Accumulator, accumulator, true)
		if r.IndexIterator != "" {
			childContext.SetLocal(r.IndexIterator, idx, true)
		}
		r, itNode, err := eval(ctx, childContext, exec, p, r.Reducer)
		node.Attach(itNode)
		if err != nil {
			return nil, itNode.SetErr(err), err
		}
		accumulator = r
	}

	return accumulator, node, nil
}
