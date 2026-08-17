// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"context"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
)

// 'is [not] defined' | 'is [not] empty' | '<expression> is <expression>'
func parseIsExpression(ctx context.Context, p *Parser, left ast.Expression, precedence Precedence) ast.Expression {
	start := p.head()

	rnge := start.Range

	// consume the 'is' token
	if !p.expect(tokens.KeywordIs) {
		return nil
	}

	var not *tokens.Instance
	if p.head().IsOfKind(tokens.KeywordNot) {
		n := p.advance() // consume the 'not' token
		not = &n
	}

	var expr ast.Expression

	if p.canExpect(tokens.KeywordDefined) {
		// 'is [not] defined' case
		expr = ast.NewIsDefinedExpression(left, rnge)
		p.advance()
	} else if p.canExpect(tokens.KeywordEmpty) {
		expr = ast.NewIsEmptyExpression(left, rnge)
		p.advance()
	} else {
		right := p.parseExpression(ctx, precedence)
		if right == nil {
			return nil
		}
		rnge.To = right.Span().To
		expr = ast.NewInfixExpression(left, right, start.Value, rnge)
	}

	// if we have a 'not' then wrap with a not unary
	if not != nil {
		expr = ast.NewUnaryExpression(not.Value, expr, rnge)
	}

	return expr
}
