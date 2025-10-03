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

func evalAny(ctx context.Context, ec *ExecutionContext, exec *executorImpl, p *index.Policy, q *ast.AnyExpression) (any, *trace.Node, error) {
	node, done := trace.New("any", "", q, map[string]any{
		"collection": q.Collection.String(),
		"value_iter": q.ValueIterator,
		"index_iter": q.IndexIterator,
		"predicate":  q.Predicate.String(),
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
		if q.IndexIterator != "" {
			childContext.SetLocal(q.IndexIterator, idx, true)
		}
		childContext.SetLocal(q.ValueIterator, item, true)
		res, resNode, err := eval(ctx, childContext, exec, p, q.Predicate)
		if err != nil {
			return nil, node.SetErr(err), err
		}
		node.Attach(resNode)
		childContext.Dispose()
		if IsTruthy(res) {
			return true, node, nil
		}
	}

	// by this time, we have iterated through the entire collection and found no truthy values
	return false, node, nil
}

// evalAll evaluates an all expression
// it returns true if all items in the collection satisfy the predicate
func evalAll(ctx context.Context, ec *ExecutionContext, exec *executorImpl, p *index.Policy, q *ast.AllExpression) (any, *trace.Node, error) {
	node, done := trace.New("all", "", q, map[string]any{
		"collection": q.Collection.String(),
		"value_iter": q.ValueIterator,
		"index_iter": q.IndexIterator,
		"predicate":  q.Predicate.String(),
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
		if q.IndexIterator != "" {
			childContext.SetLocal(q.IndexIterator, idx, true)
		}
		childContext.SetLocal(q.ValueIterator, item, true)
		res, resNode, err := eval(ctx, childContext, exec, p, q.Predicate)
		if err != nil {
			return nil, node.SetErr(err), err
		}
		node.Attach(resNode)
		childContext.Dispose()
		if !IsTruthy(res) {
			return false, node, nil
		}
	}

	return true, node, nil
}

// evalFilter evaluates a filter expression
// it returns a list of items that satisfy the predicate
// if the predicate is not satisfied, the item is not included in the list
func evalFilter(ctx context.Context, ec *ExecutionContext, exec *executorImpl, p *index.Policy, q *ast.FilterExpression) (any, *trace.Node, error) {
	node, done := trace.New("filter", "", q, map[string]any{
		"collection": q.Collection.String(),
		"value_iter": q.ValueIterator,
		"index_iter": q.IndexIterator,
		"predicate":  q.Predicate.String(),
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
		if q.IndexIterator != "" {
			childContext.SetLocal(q.IndexIterator, idx, true)
		}
		childContext.SetLocal(q.ValueIterator, item, true)
		res, resNode, err := eval(ctx, childContext, exec, p, q.Predicate)
		if err != nil {
			return nil, node.SetErr(err), err
		}
		node.Attach(resNode)
		childContext.Dispose()

		if IsTruthy(res) {
			filtered = append(filtered, item)
		}
	}

	return filtered, node, nil
}

func evalCount(ctx context.Context, ec *ExecutionContext, exec *executorImpl, p *index.Policy, c *ast.CountExpression) (any, *trace.Node, error) {
	node, done := trace.New("count", "", c, map[string]any{
		"collection": c.Collection.String(),
	})
	defer done()

	col, colNode, err := eval(ctx, ec, exec, p, c.Collection)
	node.Attach(colNode)
	if err != nil {
		return nil, node.SetErr(err), err
	}

	var count int
	switch v := col.(type) {
	case []any:
		// List - count elements
		count = len(v)
	case map[string]any:
		// Map - count key-value pairs
		count = len(v)
	case string:
		// String - count characters
		count = len(v)
	default:
		err := fmt.Errorf("count expects list, map, or string, got %T", col)
		return nil, node.SetErr(err), err
	}

	return count, node.SetResult(count), nil
}

func evalMap(ctx context.Context, ec *ExecutionContext, exec *executorImpl, p *index.Policy, m *ast.MapExpression) (any, *trace.Node, error) {
	node, done := trace.New("map", "", m, map[string]any{
		"collection": m.Collection.String(),
		"value_iter": m.ValueIterator,
		"index_iter": m.IndexIterator,
		"transform":  m.Transform.String(),
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
		childContext.SetLocal(m.IndexIterator, idx, true)
		childContext.SetLocal(m.ValueIterator, item, true)
		res, resNode, err := eval(ctx, childContext, exec, p, m.Transform)
		node.Attach(resNode)
		if err != nil {
			return nil, node.SetErr(err), err
		}
		childContext.Dispose()
		transformed = append(transformed, res)
	}

	return transformed, node, nil
}

// TBD: DISTINCT
