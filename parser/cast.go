// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"context"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
)

// parseAsCastExpression parses postfix "expr as typeRef" (LED for KeywordAs).
func parseAsCastExpression(ctx context.Context, p *Parser, left ast.Expression, _ Precedence) ast.Expression {
	if !p.expect(tokens.KeywordAs) {
		return nil
	}

	typeRef := parseTypeRef(ctx, p)
	if typeRef == nil {
		return nil
	}

	return ast.NewCastExpression(left, typeRef, tokens.Range{
		File: left.Span().File,
		From: tokens.Pos{
			Line:   left.Span().From.Line,
			Column: left.Span().From.Column,
			Offset: left.Span().From.Offset,
		},
		To: tokens.Pos{
			Line:   typeRef.Span().To.Line,
			Column: typeRef.Span().To.Column,
			Offset: typeRef.Span().To.Offset,
		},
	})
}
