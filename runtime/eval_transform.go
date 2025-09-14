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

	"github.com/binaek/sentra/ast"
	"github.com/binaek/sentra/index"
	"github.com/binaek/sentra/runtime/trace"
	"github.com/binaek/sentra/xerr"
)

func evalTransform(ctx context.Context, ec *ExecutionContext, exec *executorImpl, p *index.Policy, t *ast.TransformExpression) (any, *trace.Node, error) {
	node, done := trace.New("transform", t.Transformer, t, map[string]any{"ident": t.Argument})
	defer done()

	return nil, node, xerr.ErrNotImplemented
}
