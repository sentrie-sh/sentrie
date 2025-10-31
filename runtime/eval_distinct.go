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
	"slices"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/runtime/trace"
)

func evalDistinct(ctx context.Context, ec *ExecutionContext, exec *executorImpl, p *index.Policy, d *ast.DistinctExpression) (any, *trace.Node, error) {
	ctx, node, done := trace.New(ctx, d, "distinct", map[string]any{
		"collection": d.Collection.String(),
		"left_iter":  d.Iterator1,
		"right_iter": d.Iterator2,
	})
	defer done()

	col, colNode, err := eval(ctx, ec, exec, p, d.Collection)
	node.Attach(colNode)
	if err != nil {
		return nil, node.SetErr(err), err
	}

	list, ok := col.([]any)
	if !ok {
		return nil, node.SetErr(fmt.Errorf("distinct expects list source")), fmt.Errorf("distinct expects list source")
	}

	if len(list) < 2 {
		// nothing to do here
		return list, node, nil
	}

	// clone the list
	list = slices.Clone(list)

	theDistinct := make([]any, 0, len(list))
	theDistinct = append(theDistinct, list[0]) // start with the first item
	list = list[1:]

	// start with a distinct list of 1
	// for every item in the distinct list, iterate through the list with:
	// - the distinct item as left iterator
	// - the item as right iterator
	// - the predicate
	// if the predicate is truthy, the distinct item is the same as the item - continue to next item
	// if all predicates are falsey, add the item to the distinct list

	// iterate through the list
	for len(list) > 0 {
		// get the next item
		item := list[0]
		list = list[1:]
		foundMatch := false

		// now, iterate through the current known distinct items
		for _, distinctItem := range theDistinct {
			childContext := ec.AttachedChildContext()
			childContext.SetLocal(d.Iterator1, distinctItem, true)
			childContext.SetLocal(d.Iterator2, item, true)
			res, resNode, err := eval(ctx, childContext, exec, p, d.Quantifier)
			node.Attach(resNode)
			childContext.Dispose()
			if err != nil {
				return nil, node.SetErr(err), err
			}
			if IsTruthy(res) {
				foundMatch = true
				break
			}
		}

		// if no match was found, add the item to the distinct list
		if !foundMatch {
			theDistinct = append(theDistinct, item)
		}
	}

	theDistinct = slices.Clip(theDistinct)

	return theDistinct, node, nil
}
