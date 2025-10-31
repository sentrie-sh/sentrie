// Copyright 2025 Binaek Sarkar
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

	what, found := p.advanceExpected(tokens.Ident)
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
