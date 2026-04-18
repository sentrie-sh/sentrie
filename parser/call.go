// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"context"
	"time"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
)

func parseCallExpression(ctx context.Context, p *Parser, left ast.Expression, precedence Precedence) ast.Expression {
	if !p.expect(tokens.PunctLeftParentheses) {
		return nil
	}

	rnge := left.Span()

	arguments := parseExpressionList(ctx, p, tokens.PunctRightParentheses)
	if arguments == nil {
		return nil
	}

	// Find the closing parenthesis position
	rparen, found := p.advanceExpected(tokens.PunctRightParentheses)
	if !found {
		return nil
	}

	rnge.To = rparen.Range.To

	exp := ast.NewCallExpression(left, arguments, false, nil, rnge)

	if p.head().IsOfKind(tokens.TokenBang) {
		// advance() is enough here: we already matched TokenBang on the head.
		bang := p.advance()
		suffixTo := bang.Range.To
		var memoTTL *time.Duration
		if p.head().IsOfKind(tokens.Int) {
			literal := parseIntegerLiteral(ctx, p)
			if literal == nil {
				return nil
			}
			ttl := time.Duration(literal.(*ast.IntegerLiteral).Value) * time.Second
			memoTTL = &ttl
			suffixTo = literal.Span().To
		}
		exp.Memoized = true
		exp.MemoizeTTL = memoTTL
		rnge.To = suffixTo
	}

	return exp
}

func parseExpressionList(ctx context.Context, parser *Parser, end tokens.Kind) []ast.Expression {
	exps := []ast.Expression{}

	for parser.hasTokens() && !parser.canExpect(end) {
		exp := parser.parseExpression(ctx, LOWEST)
		if exp == nil {
			return nil
		}
		exps = append(exps, exp)
		if parser.canExpect(tokens.PunctComma) {
			parser.advance() // consume the comma
			continue
		}

	}

	return exps
}
