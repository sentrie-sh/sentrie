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

	"github.com/binaek/sentra/ast"
	"github.com/binaek/sentra/index"
	"github.com/binaek/sentra/runtime/trace"
)

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
