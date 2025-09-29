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
