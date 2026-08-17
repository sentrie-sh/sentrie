// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"context"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
)

func parseNotExpression(ctx context.Context, parser *Parser, left ast.Expression, precedence Precedence) ast.Expression {
	notToken := parser.advance()
	rnge := notToken.Range

	if !parser.canExpectAnyOf(tokens.KeywordNot, tokens.KeywordMatches, tokens.KeywordContains, tokens.KeywordIn) {
		parser.errorf("expected 'not', 'matches', 'contains', or 'in' after 'not', got %s", parser.head().Kind)
		return nil
	}

	opToken := parser.advance()
	rnge.To = opToken.Range.To

	right := parser.parseExpression(ctx, precedence)
	if right == nil {
		return nil
	}
	rnge.To = right.Span().To

	// build the infix expression of the right operand
	bin := ast.NewInfixExpression(left, right, opToken.Value, rnge)

	// wrap it in a unary expression with the not token
	return ast.NewUnaryExpression(notToken.Value, bin, rnge)
}
