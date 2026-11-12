// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"context"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
)

func parseFieldAccessExpression(ctx context.Context, p *Parser, left ast.Expression, precedence Precedence) ast.Expression {
	operatorToken := p.advance()
	if !operatorToken.IsOfKind(tokens.TokenDot) {
		return nil // Error in parsing field access
	}

	// After `.` a keyword has no other meaning, so it is read as the field
	// name: `d.title`, `x.in`, `x.not`.
	fieldName, found := advanceNameOrKeyword(p, tokens.Ident.String())
	if !found {
		return nil
	}

	return ast.NewFieldAccessExpression(left, fieldName.Value, tokens.Range{
		File: operatorToken.Range.File,
		From: operatorToken.Range.From,
		To:   fieldName.Range.To,
	})
}

func parseIndexAccessExpression(ctx context.Context, p *Parser, left ast.Expression, precedence Precedence) ast.Expression {
	lbracket, found := p.advanceExpected(tokens.PunctLeftBracket)
	if !found {
		return nil // Error in parsing index access
	}

	index := p.parseExpression(ctx, LOWEST)
	if index == nil {
		return nil // Error in parsing index expression
	}

	rBracket, found := p.advanceExpected(tokens.PunctRightBracket)
	if !found {
		return nil // Error in parsing index access
	}

	return ast.NewIndexAccessExpression(left, index, tokens.Range{
		File: rBracket.Range.File,
		From: lbracket.Range.From,
		To:   rBracket.Range.To,
	})
}
