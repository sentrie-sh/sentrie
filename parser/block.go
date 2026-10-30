// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"context"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
)

/*
*

	{
		let statement = some_expression
		-- a comment statement
		yield another_expression -- must be the last statement
	}

*
*/
func parseBlockExpression(ctx context.Context, p *Parser) ast.Expression {
	lCurly, found := p.advanceExpected(tokens.PunctLeftCurly)
	if !found {
		return nil // Error in parsing the block expression
	}

	var statements []ast.Statement

	for p.canExpectAnyOf(tokens.KeywordLet, tokens.LineComment) {
		stmt := parsePolicyStatement(ctx, p)
		if stmt == nil {
			return nil // Error in parsing the block expression
		}
		statements = append(statements, stmt)
	}

	if !p.expect(tokens.KeywordYield) {
		return nil // Error in parsing the block expression
	}

	yieldExpr := p.parseExpression(ctx, LOWEST)
	if yieldExpr == nil {
		return nil // Error in parsing the block expression
	}

	rCurly, found := p.advanceExpected(tokens.PunctRightCurly)
	if !found {
		return nil // Error in parsing the block expression
	}

	return ast.NewBlockExpression(statements, yieldExpr, tokens.Range{
		File: lCurly.Range.File,
		From: tokens.Pos{
			Line:   lCurly.Range.From.Line,
			Column: lCurly.Range.From.Column,
			Offset: lCurly.Range.From.Offset,
		},
		To: rCurly.Range.To,
	})
}

/*
*

	( some_expression )

*
*/
func parseGroupedExpression(ctx context.Context, p *Parser) ast.Expression {
	lparen := p.advance() // consume the left parenthesis

	p.lexer.PushBack(p.next)
	p.lexer.PushBack(p.current)

	params, ok := tryReadLambdaSignature(p.lexer)
	if ok {
		seen := make(map[string]struct{}, len(params))
		for _, name := range params {
			if _, dup := seen[name]; dup {
				p.errorf("duplicate lambda parameter %q", name)
				return nil
			}
			seen[name] = struct{}{}
		}
		p.current = p.lexer.NextToken()
		p.next = p.lexer.NextToken()
		bodyExpr := parseBlockExpression(ctx, p)
		body, isBlock := bodyExpr.(*ast.BlockExpression)
		if !isBlock || body == nil {
			p.errorf("lambda body must be a block expression { ... yield ... }")
			return nil
		}
		rng := tokens.Range{
			File: lparen.Range.File,
			From: lparen.Range.From,
			To:   body.Span().To,
		}
		return ast.NewLambdaExpression(params, body, rng)
	}

	p.current = p.lexer.NextToken()
	p.next = p.lexer.NextToken()
	if p.canExpect(tokens.PunctRightParentheses) ||
		(p.current.Kind == tokens.Ident && (p.next.Kind == tokens.PunctColon || p.next.Kind == tokens.TokenQuestion)) {
		if lam := parseTypedLambdaExpression(ctx, p, lparen.Range.From); lam != nil {
			return lam
		}
		if p.err != nil {
			return nil
		}
	}

	expression := p.parseExpression(ctx, LOWEST)
	if expression == nil {
		return nil
	}
	if !p.expect(tokens.PunctRightParentheses) {
		return nil
	}
	return expression
}
