// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"context"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
)

// 'export decision of @ident ( attach @ident as @expr )*'
func parseRuleExportStatement(ctx context.Context, p *Parser) ast.Statement {
	head := p.head()

	p.advance() // consume 'export'

	if p.head().Kind == tokens.KeywordDerive {
		p.errorf("policy-level derives cannot be exported; only namespace-level derives may use export derive")
		return nil
	}

	if !p.expect(tokens.KeywordDecision) {
		return nil
	}

	if !p.expect(tokens.KeywordOf) {
		return nil
	}

	ruleIdent, found := p.advanceExpected(tokens.Ident)
	if !found {
		return nil
	}

	of := ruleIdent.Value // Set the name of the exported variable or decision
	rnge := tokens.Range{
		File: head.Range.File,
		From: head.Range.From,
		To:   ruleIdent.Range.To,
	}

	attachments := []*ast.AttachmentClause{}
	for p.head().IsOfKind(tokens.KeywordAttach) {
		attachment := parseAttachmentClause(ctx, p)
		if attachment == nil {
			return nil
		}

		attachments = append(attachments, attachment)
		rnge.To = attachment.Span().To
	}

	return ast.NewRuleExportStatement(of, attachments, rnge)
}

// 'attach @ident as @expr'
func parseAttachmentClause(ctx context.Context, p *Parser) *ast.AttachmentClause {
	head := p.head()

	p.advance() // consume 'attach'

	// The attachment name may be a keyword (`attach version as ...`): right
	// after `attach` a keyword has no other meaning.
	what, found := advanceNameOrKeyword(p, tokens.Ident.String())
	if !found {
		return nil
	}

	if !p.expect(tokens.KeywordAs) {
		return nil
	}

	asExpr := p.parseExpression(ctx, LOWEST)
	if asExpr == nil {
		return nil
	}

	return ast.NewAttachmentClause(what.Value, asExpr, tokens.Range{
		File: head.Range.File,
		From: head.Range.From,
		To:   asExpr.Span().To,
	})
}
