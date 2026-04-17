// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"context"

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

	hadBang := p.head().IsOfKind(tokens.TokenBang)
	if suffix := parseMemoizationSuffix(ctx, p); suffix != nil {
		exp.Memoized = true
		exp.MemoizeTTL = suffix.TTL
		rnge.To = suffix.To
	} else if hadBang {
		return nil
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
