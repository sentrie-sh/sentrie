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

	"github.com/binaek/sentra/ast"
	"github.com/binaek/sentra/tokens"
)

func parseRuleExportStatement(ctx context.Context, p *Parser) ast.Statement {
	pos := p.head().Position

	stmt := &ast.RuleExportStatement{
		Pos: pos,
	}

	p.advance() // consume 'export'

	if !p.expect(tokens.KeywordDecision) {
		return nil
	}

	if !p.expect(tokens.KeywordOf) {
		return nil
	}

	ident, found := p.advanceExpected(tokens.Ident)
	if !found {
		return nil
	}

	stmt.Of = ident.Value // Set the name of the exported variable or decision

	for p.head().IsOfKind(tokens.KeywordAttach) {
		attachment := parseAttachmentClause(ctx, p)
		if attachment == nil {
			return nil
		}
		stmt.Attachments = append(stmt.Attachments, attachment)
	}

	return stmt
}

// 'attach @ident as @expr'
func parseAttachmentClause(ctx context.Context, p *Parser) *ast.AttachmentClause {
	pos := p.head().Position

	attachment := &ast.AttachmentClause{
		Pos: pos,
	}

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

	attachment.What = what.Value // Set the attachment what
	attachment.As = asExpr

	return attachment
}
