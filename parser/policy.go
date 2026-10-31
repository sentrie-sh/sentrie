// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"context"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
)

func parseThePolicyStatement(ctx context.Context, p *Parser) ast.Statement {
	start := p.head()

	rnge := start.Range

	if !p.expect(tokens.KeywordPolicy) {
		return nil
	}

	nameIdent, ok := p.advanceExpected(tokens.Ident)
	if !ok {
		return nil
	}

	name := nameIdent.Value

	if !p.expect(tokens.PunctLeftCurly) {
		return nil
	}

	var statements []ast.Statement

	for p.hasTokens() && !p.head().IsOfKind(tokens.PunctRightCurly) {
		stmt := parsePolicyStatement(ctx, p)
		if p.err != nil {
			return nil
		}
		if stmt == nil {
			return nil // Error in parsing the policy statement
		}
		statements = append(statements, stmt)

		// consume the optional semicolon
		if p.canExpect(tokens.PunctSemicolon) {
			p.advance()
		}

		// consume trailing comments
		for p.canExpectAnyOf(tokens.TrailingComment, tokens.LineComment) {
			p.advance()
		}
	}

	rCurly, found := p.advanceExpected(tokens.PunctRightCurly)
	if !found {
		return nil
	}
	rnge.To = rCurly.Range.To

	return ast.NewPolicyStatement(name, statements, rnge)
}

func parsePolicyStatement(ctx context.Context, p *Parser) ast.Statement {
	// `data` is a contextual keyword: it starts a declaration only at the head
	// of a policy statement, where a bare identifier is otherwise never valid,
	// so `data` stays usable as a fact, let, field or parameter name.
	if p.head().IsOfKind(tokens.Ident) && p.head().Value == dataKeyword {
		return parseDataStatement(ctx, p)
	}
	if handler, ok := p.policyStatementHandlers[p.head().Kind]; ok {
		return handler(ctx, p)
	}
	p.errorf("unexpected token '%s'", p.head().Kind)
	return nil
}

func (p *Parser) registerPolicyStatementHandler(tokenType tokens.Kind, fn statementParser) {
	p.policyStatementHandlers[tokenType] = fn
}
