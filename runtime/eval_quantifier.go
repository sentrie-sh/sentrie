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
	"github.com/sentrie-sh/sentrie/trinary"
)

func evalAny(ctx context.Context, ec *ExecutionContext, exec *executorImpl, p *index.Policy, q *ast.AnyExpression) (any, *trace.Node, error) {
	ctx, node, done := trace.New(ctx, q, "any", map[string]any{
		"collection": q.Collection.String(),
		"value_iter": q.Iterator1,
		"index_iter": q.Iterator2,
		"predicate":  q.Quantifier.String(),
	})
	defer done()

	col, colNode, err := eval(ctx, ec, exec, p, q.Collection)
	node.Attach(colNode)
	if err != nil {
		return nil, node.SetErr(err), err
	}

	if IsUndefined(col) {
		return false, node, nil
	}

	list, ok := col.([]any)
	if !ok {
		return nil, node.SetErr(fmt.Errorf("any expects list source")), fmt.Errorf("any expects list source")
	}

	for idx, item := range list {
		childContext := ec.AttachedChildContext()
		if q.Iterator2 != "" {
			childContext.SetLocal(q.Iterator2, idx, true)
		}
		childContext.SetLocal(q.Iterator1, item, true)
		res, resNode, err := eval(ctx, childContext, exec, p, q.Quantifier)
		if err != nil {
			return nil, node.SetErr(err), err
		}
		node.Attach(resNode)
		childContext.Dispose()
		if trinary.From(res).IsTrue() {
			return true, node, nil
		}
	}

	// by this time, we have iterated through the entire collection and found no truthy values
	return false, node, nil
}

// evalAll evaluates an all expression
// it returns true if all items in the collection satisfy the predicate
func evalAll(ctx context.Context, ec *ExecutionContext, exec *executorImpl, p *index.Policy, q *ast.AllExpression) (any, *trace.Node, error) {
	ctx, node, done := trace.New(ctx, q, "all", map[string]any{
		"collection": q.Collection.String(),
		"value_iter": q.Iterator1,
		"index_iter": q.Iterator2,
		"predicate":  q.Quantifier.String(),
	})
	defer done()

	col, colNode, err := eval(ctx, ec, exec, p, q.Collection)
	node.Attach(colNode)
	if err != nil {
		return nil, node.SetErr(err), err
	}

	if IsUndefined(col) {
		return false, node, nil
	}

	list, ok := col.([]any)
	if !ok {
		return nil, node.SetErr(fmt.Errorf("all expects list source")), fmt.Errorf("all expects list source")
	}

	for idx, item := range list {
		childContext := ec.AttachedChildContext()
		if q.Iterator2 != "" {
			childContext.SetLocal(q.Iterator2, idx, true)
		}
		childContext.SetLocal(q.Iterator1, item, true)
		res, resNode, err := eval(ctx, childContext, exec, p, q.Quantifier)
		if err != nil {
			return nil, node.SetErr(err), err
		}
		node.Attach(resNode)
		childContext.Dispose()
		if !trinary.From(res).IsTrue() {
			return false, node, nil
		}
	}

	return true, node, nil
}

// evalFirst evaluates a first expression
// it returns the first item in the collection that satisfies the predicate
// if no item satisfies the predicate, it returns undefined
func evalFirst(ctx context.Context, ec *ExecutionContext, exec *executorImpl, p *index.Policy, q *ast.FirstExpression) (any, *trace.Node, error) {
	ctx, node, done := trace.New(ctx, q, "first", map[string]any{
		"collection": q.Collection.String(),
		"value_iter": q.Iterator1,
		"index_iter": q.Iterator2,
		"predicate":  q.Quantifier.String(),
	})
	defer done()

	col, colNode, err := eval(ctx, ec, exec, p, q.Collection)
	node.Attach(colNode)
	if err != nil {
		return nil, node.SetErr(err), err
	}

	if IsUndefined(col) {
		return Undefined, node.SetResult(Undefined), nil
	}

	list, ok := col.([]any)
	if !ok {
		return nil, node.SetErr(fmt.Errorf("first expects list source")), fmt.Errorf("first expects list source")
	}

	for idx, item := range list {
		childContext := ec.AttachedChildContext()
		if q.Iterator2 != "" {
			childContext.SetLocal(q.Iterator2, idx, true)
		}
		childContext.SetLocal(q.Iterator1, item, true)
		res, resNode, err := eval(ctx, childContext, exec, p, q.Quantifier)
		if err != nil {
			return nil, node.SetErr(err), err
		}
		node.Attach(resNode)
		childContext.Dispose()
		if trinary.From(res).IsTrue() {
			return item, node, nil
		}
	}

	// by this time, we have iterated through the entire collection and found no truthy values
	// return undefined
	return Undefined, node.SetResult(Undefined), nil
}

// evalFilter evaluates a filter expression
// it returns a list of items that satisfy the predicate
// if the predicate is not satisfied, the item is not included in the list
func evalFilter(ctx context.Context, ec *ExecutionContext, exec *executorImpl, p *index.Policy, q *ast.FilterExpression) (any, *trace.Node, error) {
	ctx, node, done := trace.New(ctx, q, "filter", map[string]any{
		"collection": q.Collection.String(),
		"value_iter": q.Iterator1,
		"index_iter": q.Iterator2,
		"predicate":  q.Quantifier.String(),
	})
	defer done()

	col, colNode, err := eval(ctx, ec, exec, p, q.Collection)
	node.Attach(colNode)
	if err != nil {
		return nil, node.SetErr(err), err
	}

	if IsUndefined(col) {
		return []any{}, node, nil
	}

	list, ok := col.([]any)
	if !ok {
		return nil, node.SetErr(fmt.Errorf("filter expects list source")), fmt.Errorf("filter expects list source")
	}
	filtered := make([]any, 0, len(list))

	for idx, item := range list {
		childContext := ec.AttachedChildContext()
		if q.Iterator2 != "" {
			childContext.SetLocal(q.Iterator2, idx, true)
		}
		childContext.SetLocal(q.Iterator1, item, true)
		res, resNode, err := eval(ctx, childContext, exec, p, q.Quantifier)
		if err != nil {
			return nil, node.SetErr(err), err
		}
		node.Attach(resNode)
		childContext.Dispose()

		if trinary.From(res).IsTrue() {
			filtered = append(filtered, item)
		}
	}

	return filtered, node, nil
}

func evalMap(ctx context.Context, ec *ExecutionContext, exec *executorImpl, p *index.Policy, m *ast.MapExpression) (any, *trace.Node, error) {
	ctx, node, done := trace.New(ctx, m, "map", map[string]any{
		"collection": m.Collection.String(),
		"value_iter": m.Iterator1,
		"index_iter": m.Iterator2,
		"transform":  m.Quantifier.String(),
	})
	defer done()

	col, colNode, err := eval(ctx, ec, exec, p, m.Collection)
	node.Attach(colNode)
	if err != nil {
		return nil, node.SetErr(err), err
	}

	list, ok := col.([]any)
	if !ok {
		return nil, node.SetErr(fmt.Errorf("map expects list source")), fmt.Errorf("map expects list source")
	}

	transformed := make([]any, 0, len(list))
	for idx, item := range list {
		childContext := ec.AttachedChildContext()
		childContext.SetLocal(m.Iterator2, idx, true)
		childContext.SetLocal(m.Iterator1, item, true)
		res, resNode, err := eval(ctx, childContext, exec, p, m.Quantifier)
		node.Attach(resNode)
		if err != nil {
			return nil, node.SetErr(err), err
		}
		childContext.Dispose()
		transformed = append(transformed, res)
	}

	return transformed, node, nil
}
