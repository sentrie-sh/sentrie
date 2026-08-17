// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"testing"

	"github.com/sentrie-sh/sentrie/ast"
)

func (s *ParserTestSuite) TestParseNotExpressionForms() {
	s.T().Run("notInList", func(t *testing.T) {
		parser := NewParserFromString("9 not in [1, 2, 3]", "test.sentrie")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.Require().NotNil(expr)

		unary, ok := expr.(*ast.UnaryExpression)
		s.Require().True(ok)
		s.Equal("not", unary.Operator)

		infix, ok := unary.Right.(*ast.InfixExpression)
		s.Require().True(ok)
		s.Equal("in", infix.Operator)

		list, ok := infix.Right.(*ast.ListLiteral)
		s.Require().True(ok)
		s.Len(list.Values, 3)
	})

	s.T().Run("notContains", func(t *testing.T) {
		parser := NewParserFromString(`["x", "y"] not contains "z"`, "test.sentrie")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.Require().NotNil(expr)
		s.Equal(`not(["x", "y"] contains "z")`, expr.String())
	})

	s.T().Run("notMatches", func(t *testing.T) {
		parser := NewParserFromString(`"foo" not matches "[0-9]+"`, "test.sentrie")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.Require().NotNil(expr)
		s.Equal(`not("foo" matches "[0-9]+")`, expr.String())
	})

	s.T().Run("invalidNotOperand", func(t *testing.T) {
		parser := NewParserFromString("9 not 3", "test.sentrie")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.Nil(expr)
		s.Error(parser.err)
		s.Contains(parser.err.Error(), "expected 'not', 'matches', 'contains', or 'in' after 'not'")
	})
}
