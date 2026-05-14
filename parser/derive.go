// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"context"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
)

func parseDeriveStatement(ctx context.Context, p *Parser) ast.Statement {
	head := p.head()
	p.advance() // derive
	name, found := p.advanceExpected(tokens.Ident)
	if !found {
		return nil
	}
	if !p.expect(tokens.TokenAssign) {
		return nil
	}
	expr := p.parseExpression(ctx, LOWEST)
	if expr == nil {
		return nil
	}
	lam, ok := expr.(*ast.LambdaExpression)
	if !ok {
		p.errorf("derive value must be a lambda expression")
		return nil
	}
	rnge := tokens.Range{
		File: head.Range.File,
		From: head.Range.From,
		To:   lam.Span().To,
	}
	return ast.NewDeriveStatement(name.Value, lam, rnge)
}
