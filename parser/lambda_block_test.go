// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2026 Binaek Sarkar

package parser

import (
	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
)

func (s *ParserTestSuite) TestParseGroupedVsLambda_ParsesLambda() {
	p := NewParserFromString("(x, idx) => { yield x + idx }", "test.sentra")
	expr := p.parseExpression(s.T().Context(), LOWEST)
	s.Require().NotNil(expr)
	s.Require().NoError(p.err)

	lam, ok := expr.(*ast.LambdaExpression)
	s.Require().True(ok)
	s.Equal([]string{"x", "idx"}, lam.Params)
	s.NotNil(lam.Body)
}

func (s *ParserTestSuite) TestParseGroupedVsLambda_ParsesGroupedExpression() {
	p := NewParserFromString("(1 + 2)", "test.sentra")
	expr := p.parseExpression(s.T().Context(), LOWEST)
	s.Require().NotNil(expr)
	s.Require().NoError(p.err)
	_, ok := expr.(*ast.InfixExpression)
	s.True(ok)
}

func (s *ParserTestSuite) TestParseGroupedVsLambda_DuplicateParamsError() {
	p := NewParserFromString("(x, x) => { yield x }", "test.sentra")
	expr := p.parseExpression(s.T().Context(), LOWEST)
	s.Nil(expr)
	s.Error(p.err)
}

func (s *ParserTestSuite) TestTryReadLambdaSignature_FailurePushback() {
	// Missing fat arrow: this must fail and push all read tokens back.
	p := NewParserFromString("(x, y)", "test.sentra")
	// p.head() is "(" after constructor advances.
	lparen := p.head()
	s.Equal(tokens.PunctLeftParentheses, lparen.Kind)
	p.advance() // consume "(" to mimic parseGroupedExpression flow

	params, ok := tryReadLambdaSignature(p.lexer)
	s.False(ok)
	s.Nil(params)

	// Since lookahead failed, lexer stream should still return what it saw first.
	tok := p.lexer.NextToken()
	s.Equal(tokens.Ident, tok.Kind)
	s.Equal("y", tok.Value)
}
