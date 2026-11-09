// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"github.com/sentrie-sh/sentrie/ast"
)

func (s *RuntimeTestSuite) TestEvalDistinctDirectScalars() {
	ctx := s.T().Context()
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
	s.Equal([]any{int64(1), int64(2), int64(3)}, result.Any())
}

func (s *RuntimeTestSuite) TestEvalDistinctSelectorKey() {
	ctx := s.T().Context()
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
	ctx := s.T().Context()
	p := newEvalTestPolicy()
	ec := NewExecutionContext(p, &executorImpl{})
	exec := &executorImpl{}

	call := ast.NewCallExpression(ast.NewIdentifier("distinct", stubRange()), []ast.Expression{
		ast.NewStringLiteral("nope", stubRange()),
	}, false, nil, stubRange())
	_, _, err := eval(ctx, ec, exec, p, call)
	s.ErrorContains(err, "distinct: first argument must be a list")
}
