// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"context"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
)

// parseTypedLambdaExpression parses ( params ) [: ret] => { ... } when tryReadLambdaSignature
// failed. Pre: p.current is the first token inside '(' (Ident or ')').
func parseTypedLambdaExpression(ctx context.Context, p *Parser, lparenFrom tokens.Pos) *ast.LambdaExpression {
	file := p.current.Range.File

	var names []string
	var types []ast.TypeRef
	var opts []bool
	anyPriorOptional := false

	if !p.canExpect(tokens.PunctRightParentheses) {
		for {
			nameTok, ok := p.advanceExpected(tokens.Ident)
			if !ok {
				return nil
			}
			optional := false
			if p.canExpect(tokens.TokenQuestion) {
				p.advance()
				optional = true
			}
			var typ ast.TypeRef
			if p.canExpect(tokens.PunctColon) {
				p.advance()
				typ = parseTypeRef(ctx, p)
				if typ == nil {
					return nil
				}
			}
			if !optional && anyPriorOptional {
				p.errorf("required lambda parameter %q cannot follow an optional parameter", nameTok.Value)
				return nil
			}
			if optional {
				anyPriorOptional = true
			}
			names = append(names, nameTok.Value)
			types = append(types, typ)
			opts = append(opts, optional)

			if p.canExpect(tokens.PunctComma) {
				p.advance()
				continue
			}
			if p.canExpect(tokens.PunctRightParentheses) {
				break
			}
			p.errorf("expected ',' or ')' in lambda parameter list, got %s", p.head().Kind)
			return nil
		}
	}

	if !p.expect(tokens.PunctRightParentheses) {
		return nil
	}

	var returnType ast.TypeRef
	if p.canExpect(tokens.PunctColon) {
		// return type (not to be confused with ':' inside params — we are past ')')
		p.advance()
		returnType = parseTypeRef(ctx, p)
		if returnType == nil {
			return nil
		}
	}

	if !p.expect(tokens.TokenFatArrow) {
		return nil
	}

	bodyExpr := parseBlockExpression(ctx, p)
	body, isBlock := bodyExpr.(*ast.BlockExpression)
	if !isBlock || body == nil {
		p.errorf("lambda body must be a block expression { ... yield ... }")
		return nil
	}

	rnge := tokens.Range{
		File: file,
		From: lparenFrom,
		To:   body.Span().To,
	}
	return ast.NewLambdaExpressionFull(names, types, opts, returnType, body, rnge)
}
