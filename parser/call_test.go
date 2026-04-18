// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"testing"
	"time"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
)

func (s *ParserTestSuite) TestParseExpressionList() {
	ctx := s.T().Context()

	s.T().Run("empty_before_close", func(t *testing.T) {
		p := NewParserFromString(")", "test.sentra")
		args := parseExpressionList(ctx, p, tokens.PunctRightParentheses)
		s.NotNil(args)
		s.Empty(args)
		s.Nil(p.err)
	})

	s.T().Run("single_integer", func(t *testing.T) {
		p := NewParserFromString("42)", "test.sentra")
		args := parseExpressionList(ctx, p, tokens.PunctRightParentheses)
		s.Require().Len(args, 1)
		lit, ok := args[0].(*ast.IntegerLiteral)
		s.Require().True(ok)
		s.Equal(42.0, lit.Value)
		s.Nil(p.err)
	})

	s.T().Run("two_comma_separated", func(t *testing.T) {
		p := NewParserFromString("1, 2)", "test.sentra")
		args := parseExpressionList(ctx, p, tokens.PunctRightParentheses)
		s.Require().Len(args, 2)
		a0, ok := args[0].(*ast.IntegerLiteral)
		s.Require().True(ok)
		s.Equal(1.0, a0.Value)
		a1, ok := args[1].(*ast.IntegerLiteral)
		s.Require().True(ok)
		s.Equal(2.0, a1.Value)
		s.Nil(p.err)
	})

	s.T().Run("double_comma", func(t *testing.T) {
		p := NewParserFromString("1,,2)", "test.sentra")
		s.Nil(parseExpressionList(ctx, p, tokens.PunctRightParentheses))
		s.Error(p.err)
	})
}

func (s *ParserTestSuite) TestParseCallExpressionArguments() {
	ctx := s.T().Context()
	rng := tokens.BadRange("test.sentra")
	left := ast.NewIdentifier("fn", rng)

	s.T().Run("empty", func(t *testing.T) {
		p := NewParserFromString("()", "test.sentra")
		expr := parseCallExpression(ctx, p, left, CALL)
		s.Require().NotNil(expr)
		call := expr.(*ast.CallExpression)
		s.Empty(call.Arguments)
		s.Nil(p.err)
	})

	s.T().Run("two_literals", func(t *testing.T) {
		p := NewParserFromString("(1, 2)", "test.sentra")
		expr := parseCallExpression(ctx, p, left, CALL)
		s.Require().NotNil(expr)
		call := expr.(*ast.CallExpression)
		s.Require().Len(call.Arguments, 2)
		s.Nil(p.err)
	})

	s.T().Run("identifier_arg", func(t *testing.T) {
		p := NewParserFromString("(x)", "test.sentra")
		expr := parseCallExpression(ctx, p, left, CALL)
		s.Require().NotNil(expr)
		call := expr.(*ast.CallExpression)
		s.Require().Len(call.Arguments, 1)
		id, ok := call.Arguments[0].(*ast.Identifier)
		s.Require().True(ok)
		s.Equal("x", id.Value)
		s.Nil(p.err)
	})
}

func (s *ParserTestSuite) TestParseCallExpressionErrorBranches() {
	rng := tokens.BadRange("test.sentra")
	left := ast.NewIdentifier("fn", rng)

	p := NewParserFromString("fn", "test.sentra")
	s.Nil(parseCallExpression(s.T().Context(), p, left, CALL))
	s.Error(p.err)
	s.Contains(p.err.Error(), "expected")

	p = NewParserFromString("fn(,)", "test.sentra")
	callHead := p.parseExpression(s.T().Context(), LOWEST)
	s.Nil(callHead)
	s.Error(p.err)

	p = NewParserFromString("fn(1", "test.sentra")
	s.Nil(p.parseExpression(s.T().Context(), LOWEST))
	s.Error(p.err)
	s.Contains(p.err.Error(), "expected")
}

func (s *ParserTestSuite) TestParseCallExpressionMemoizationSuffix() {
	rng := tokens.BadRange("test.sentra")
	left := ast.NewIdentifier("fn", rng)

	p := NewParserFromString("()!15", "test.sentra")
	expr := parseCallExpression(s.T().Context(), p, left, CALL)
	s.Require().NotNil(expr)
	call, ok := expr.(*ast.CallExpression)
	s.Require().True(ok)
	s.True(call.Memoized)
	s.Require().NotNil(call.MemoizeTTL)
	s.Equal(15*time.Second, *call.MemoizeTTL)

	p = NewParserFromString("()!", "test.sentra")
	expr = parseCallExpression(s.T().Context(), p, left, CALL)
	s.Require().NotNil(expr)
	call = expr.(*ast.CallExpression)
	s.True(call.Memoized)
	s.Nil(call.MemoizeTTL)

	p = NewParserFromString("()", "test.sentra")
	expr = parseCallExpression(s.T().Context(), p, left, CALL)
	s.Require().NotNil(expr)
	call = expr.(*ast.CallExpression)
	s.False(call.Memoized)
	s.Nil(call.MemoizeTTL)
}

func (s *ParserTestSuite) TestParseCallExpressionMemoizationSuffixParseFailure() {
	rng := tokens.BadRange("test.sentra")
	left := ast.NewIdentifier("fn", rng)
	p := NewParserFromString("()!9223372036854775808", "test.sentra")
	s.Nil(parseCallExpression(s.T().Context(), p, left, CALL))
	s.Error(p.err)
	s.Contains(p.err.Error(), "invalid integer literal")
}
