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
)

func (s *RuntimeTestSuite) TestEvalDistinctDirectScalars() {
	ctx := context.Background()
	p := newEvalTestPolicy()
	ec := NewExecutionContext(p, &executorImpl{})
	exec := &executorImpl{}

	list := ast.NewListLiteral([]ast.Expression{
		ast.NewIntegerLiteral(1, stubRange()),
		ast.NewIntegerLiteral(2, stubRange()),
		ast.NewIntegerLiteral(1, stubRange()),
		ast.NewIntegerLiteral(3, stubRange()),
		ast.NewIntegerLiteral(2, stubRange()),
	}, stubRange())
	call := ast.NewCallExpression(ast.NewIdentifier("distinct", stubRange()), []ast.Expression{list}, false, nil, stubRange())
	result, _, err := eval(ctx, ec, exec, p, call)
	s.NoError(err)
	s.Equal([]any{float64(1), float64(2), float64(3)}, result.Any())
}

func (s *RuntimeTestSuite) TestEvalDistinctSelectorKey() {
	ctx := context.Background()
	p := newEvalTestPolicy()
	ec := NewExecutionContext(p, &executorImpl{})
	exec := &executorImpl{}

	list := ast.NewListLiteral([]ast.Expression{
		ast.NewStringLiteral("a", stubRange()),
		ast.NewStringLiteral("a", stubRange()),
		ast.NewStringLiteral("b", stubRange()),
	}, stubRange())
	call := ast.NewCallExpression(ast.NewIdentifier("distinct", stubRange()), []ast.Expression{
		list,
		stubLambda([]string{"x"}, ast.NewIdentifier("x", stubRange())),
	}, false, nil, stubRange())
	result, _, err := eval(ctx, ec, exec, p, call)
	s.NoError(err)
	s.Equal([]any{"a", "b"}, result.Any())
}

func (s *RuntimeTestSuite) TestEvalDistinctNonListErrors() {
	ctx := context.Background()
	p := newEvalTestPolicy()
	ec := NewExecutionContext(p, &executorImpl{})
	exec := &executorImpl{}

	call := ast.NewCallExpression(ast.NewIdentifier("distinct", stubRange()), []ast.Expression{
		ast.NewStringLiteral("nope", stubRange()),
	}, false, nil, stubRange())
	_, _, err := eval(ctx, ec, exec, p, call)
	s.ErrorContains(err, "distinct: first argument must be a list")
}
