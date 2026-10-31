// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"github.com/sentrie-sh/sentrie/ast"
)

// TestParseExpressionExponentUnaryMinus tests that a leading minus is a unary
// operator on the whole exponent literal.
func (s *ParserTestSuite) TestParseExpressionExponentUnaryMinus() {
	parser := NewParserFromString("-1e3", "test.sentra")
	expr := parser.parseExpression(s.T().Context(), LOWEST)
	s.Require().NoError(parser.err)

	unary, ok := expr.(*ast.UnaryExpression)
	s.Require().True(ok, "parsed as %T", expr)
	s.Equal("-", unary.Operator)
	operand, ok := unary.Right.(*ast.FloatLiteral)
	s.Require().True(ok, "operand parsed as %T", unary.Right)
	s.Equal(1e3, operand.Value)
}

// TestParseExpressionExponentPrecedence tests that exponent literals take part
// in operator precedence like any other number.
func (s *ParserTestSuite) TestParseExpressionExponentPrecedence() {
	parser := NewParserFromString("2e0 + 3e1 * 1e-1", "test.sentra")
	expr := parser.parseExpression(s.T().Context(), LOWEST)
	s.Require().NoError(parser.err)

	sum, ok := expr.(*ast.InfixExpression)
	s.Require().True(ok, "parsed as %T", expr)
	s.Equal("+", sum.Operator)
	product, ok := sum.Right.(*ast.InfixExpression)
	s.Require().True(ok, "right parsed as %T", sum.Right)
	s.Equal("*", product.Operator)
	left, ok := product.Left.(*ast.FloatLiteral)
	s.Require().True(ok, "product left parsed as %T", product.Left)
	s.Equal(3e1, left.Value)
	right, ok := product.Right.(*ast.FloatLiteral)
	s.Require().True(ok, "product right parsed as %T", product.Right)
	s.Equal(1e-1, right.Value)
}

// TestParseExpressionExponentInCollections tests exponent literals as list
// elements and map values.
func (s *ParserTestSuite) TestParseExpressionExponentInCollections() {
	ctx := s.T().Context()

	listParser := NewParserFromString("[1e2, 5, 2.5e-1]", "test.sentra")
	list, ok := listParser.parseExpression(ctx, LOWEST).(*ast.ListLiteral)
	s.Require().NoError(listParser.err)
	s.Require().True(ok)
	s.Require().Len(list.Values, 3)
	first, ok := list.Values[0].(*ast.FloatLiteral)
	s.Require().True(ok, "first parsed as %T", list.Values[0])
	s.Equal(1e2, first.Value)
	_, ok = list.Values[1].(*ast.IntegerLiteral)
	s.True(ok, "plain integer stays an integer, got %T", list.Values[1])
	third, ok := list.Values[2].(*ast.FloatLiteral)
	s.Require().True(ok, "third parsed as %T", list.Values[2])
	s.Equal(2.5e-1, third.Value)

	mapParser := NewParserFromString(`{ "k": 1e-9 }`, "test.sentra")
	m, ok := mapParser.parseExpression(ctx, LOWEST).(*ast.MapLiteral)
	s.Require().NoError(mapParser.err)
	s.Require().True(ok)
	s.Require().Len(m.Entries, 1)
	value, ok := m.Entries[0].Value.(*ast.FloatLiteral)
	s.Require().True(ok, "map value parsed as %T", m.Entries[0].Value)
	s.Equal(1e-9, value.Value)
}

// TestParseExpressionFieldNamedLikeExponent tests that a field whose name
// looks like an exponent is still a field access: the exponent rule applies
// only straight after digits.
func (s *ParserTestSuite) TestParseExpressionFieldNamedLikeExponent() {
	parser := NewParserFromString("x.e5", "test.sentra")
	expr := parser.parseExpression(s.T().Context(), LOWEST)
	s.Require().NoError(parser.err)

	access, ok := expr.(*ast.FieldAccessExpression)
	s.Require().True(ok, "parsed as %T", expr)
	s.Equal("e5", access.Field)
}

// TestParseProgramReportsOtherLexerErrorsVerbatim pins the behaviour change
// that came with #180: a lexer error token in expression position is
// reported with the lexer's message, such as an unterminated string, and not
// as a missing prefix parse function.
func (s *ParserTestSuite) TestParseProgramReportsOtherLexerErrorsVerbatim() {
	src := "namespace com/example\npolicy p {\n  rule r = \"abc\n  export decision of r\n}\n"
	parser := NewParserFromString(src, "test.sentra")
	_, err := parser.ParseProgram(s.T().Context())
	s.Require().Error(err)
	s.Contains(err.Error(), "unterminated string literal")
	s.NotContains(err.Error(), "no prefix parse function")
}

// TestParseTypeRefConstraintReportsLexerErrorVerbatim pins the same behaviour
// change in constraint-argument position: an unterminated string is reported
// as such, not as "constraint arguments must be literals".
func (s *ParserTestSuite) TestParseTypeRefConstraintReportsLexerErrorVerbatim() {
	parser := NewParserFromString(`shape T { amount:string @minlength("abc) }`, "test.sentra")
	stmt := parseShapeStatement(s.T().Context(), parser)
	s.Nil(stmt)
	s.Require().Error(parser.err)
	s.Contains(parser.err.Error(), "unterminated string literal")
	s.NotContains(parser.err.Error(), "must be literals")
}
