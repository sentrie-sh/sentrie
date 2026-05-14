// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"context"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
)

func parseTernaryExpression(ctx context.Context, p *Parser, condition ast.Expression, precedence Precedence) ast.Expression {
	rnge := condition.Span()

	// Parse the '?' token
	if !p.expect(tokens.TokenQuestion) {
		return nil
	}

	// Elvis: `a ?: b` — `?` immediately followed by `:`
	if p.canExpect(tokens.PunctColon) {
		p.advance()
		rhs := p.parseExpression(ctx, COMPARISON)
		if rhs == nil {
			return nil
		}
		rnge.To = rhs.Span().To
		return ast.NewTernaryElvis(condition, rhs, rnge)
	}

	// Parse the true branch
	// we default ot the condition expression itself
	trueBranch := condition
	if !p.canExpect(tokens.PunctColon) {
		trueBranch = p.parseExpression(ctx, precedence)
		if trueBranch == nil {
			return nil
		}
		rnge.To = trueBranch.Span().To
	}

	// Parse the ':' token
	if !p.expect(tokens.PunctColon) {
		return nil
	}

	// Parse the false branch
	falseBranch := p.parseExpression(ctx, precedence)
	if falseBranch == nil {
		return nil
	}
	rnge.To = falseBranch.Span().To

	return ast.NewTernaryExpression(condition, trueBranch, falseBranch, rnge)
}
