// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/trinary"
)

func (s *RuntimeTestSuite) TestEvalUnaryNotAndBangUseTrinaryNegation() {
	tests := []struct {
		name     string
		operator string
		in       trinary.Value
		want     trinary.Value
	}{
		{name: "bang true to false", operator: "!", in: trinary.True, want: trinary.False},
		{name: "bang false to true", operator: "!", in: trinary.False, want: trinary.True},
		{name: "bang unknown remains unknown", operator: "!", in: trinary.Unknown, want: trinary.Unknown},
		{name: "not true to false", operator: "not", in: trinary.True, want: trinary.False},
		{name: "not false to true", operator: "not", in: trinary.False, want: trinary.True},
		{name: "not unknown remains unknown", operator: "not", in: trinary.Unknown, want: trinary.Unknown},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			expr := ast.NewUnaryExpression(tt.operator, ast.NewTrinaryLiteral(tt.in, stubRange()), stubRange())
			got, _, err := evalUnary(s.T().Context(), NewExecutionContext(newEvalTestPolicy(), &executorImpl{}), &executorImpl{}, newEvalTestPolicy(), expr)
			s.Require().NoError(err)
			s.Require().Equal(box.Trinary(tt.want), got)
		})
	}
}

func (s *RuntimeTestSuite) TestEvalUnaryNumberOperatorsAndErrors() {
	p := newEvalTestPolicy()
	ec := NewExecutionContext(p, &executorImpl{})

	plus, _, err := evalUnary(s.T().Context(), ec, &executorImpl{}, p, ast.NewUnaryExpression("+", ast.NewIntegerLiteral(7, stubRange()), stubRange()))
	s.Require().NoError(err)
	s.Require().Equal(7.0, plus.Any())

	minus, _, err := evalUnary(s.T().Context(), ec, &executorImpl{}, p, ast.NewUnaryExpression("-", ast.NewIntegerLiteral(7, stubRange()), stubRange()))
	s.Require().NoError(err)
	s.Require().Equal(-7.0, minus.Any())

	_, _, err = evalUnary(s.T().Context(), ec, &executorImpl{}, p, ast.NewUnaryExpression("+", ast.NewStringLiteral("x", stubRange()), stubRange()))
	s.Require().ErrorContains(err, "unary + requires number")

	_, _, err = evalUnary(s.T().Context(), ec, &executorImpl{}, p, ast.NewUnaryExpression("-", ast.NewStringLiteral("x", stubRange()), stubRange()))
	s.Require().ErrorContains(err, "unary - requires number")
}

func (s *RuntimeTestSuite) TestEvalUnaryUndefinedPassthrough() {
	p := newEvalTestPolicy()
	ec := NewExecutionContext(p, &executorImpl{})
	missing := ast.NewFieldAccessExpression(ast.NewMapLiteral([]ast.MapEntry{}, stubRange()), "missing", stubRange())

	got, _, err := evalUnary(s.T().Context(), ec, &executorImpl{}, p, ast.NewUnaryExpression("!", missing, stubRange()))
	s.Require().NoError(err)
	s.Require().True(got.IsUndefined())
}
