// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"context"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
)

// contribute <ident> ['when' expr] expr
func parseContributeStatement(ctx context.Context, parser *Parser) ast.Statement {
	contributeToken, found := parser.advanceExpected(tokens.KeywordContribute)
	if !found {
		return nil
	}
	rnge := contributeToken.Range

	nameToken, found := parser.advanceExpected(tokens.Ident)
	if !found {
		return nil
	}

	var whenExpr ast.Expression
	if parser.canExpect(tokens.KeywordWhen) {
		parser.advance() // consume 'when'
		whenExpr = parser.parseExpression(ctx, LOWEST)
		if whenExpr == nil {
			return nil
		}
	}

	body := parser.parseExpression(ctx, LOWEST)
	if body == nil {
		return nil
	}
	rnge.To = body.Span().To

	return ast.NewContributeStatement(nameToken.Value, whenExpr, body, rnge)
}
