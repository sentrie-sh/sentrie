// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"context"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
)

// <expression> <operator> <expression>
func parseInfixExpression(ctx context.Context, p *Parser, left ast.Expression, precedence Precedence) ast.Expression {
	operatorToken := p.advance()
	rhsPrec := precedence
	// Slash division must bind tighter than a following call so `com/ex/f(x)` is a call on the
	// slash chain `(com/ex/f)(x)`, not `com/ex/(f(x))`.
	if operatorToken.Kind == tokens.TokenDiv {
		rhsPrec = INDEX
	}
	var right ast.Expression
	if operatorToken.Kind == tokens.TokenDiv && p.isNameOnlyKeyword(p.head()) {
		// A keyword that cannot start an expression is a name segment after
		// `/`, so `lib/in/double(2)` and `x is text/string/Item` name a derive
		// or shape in a keyword-segment namespace (#297). Nothing can follow
		// the segment at INDEX precedence, so the identifier is the whole
		// right operand.
		tok := p.advance()
		right = wrapWithTrailingComment(ast.NewIdentifier(tok.Value, tok.Range), p)
	} else {
		right = p.parseExpression(ctx, rhsPrec)
	}

	// Check if right operand parsing failed
	if right == nil || p.err != nil {
		return nil
	}

	return ast.NewInfixExpression(left, right, operatorToken.Value, tokens.Range{
		File: operatorToken.Range.File,
		From: operatorToken.Range.From,
		To:   right.Span().To,
	})
}
