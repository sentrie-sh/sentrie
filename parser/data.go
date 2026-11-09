// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"context"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
)

// dataKeyword starts a data declaration inside a policy. It is contextual, not
// a lexer keyword (see parsePolicyStatement).
const dataKeyword = "data"

// 'data' @ident ( ':' <type> )? ( 'as' @ident )?
//
// Declares that the policy reads the pack's static `[data]` entry @ident.
func parseDataStatement(ctx context.Context, p *Parser) ast.Statement {
	head := p.head()
	if !head.IsOfKind(tokens.Ident) || head.Value != dataKeyword {
		p.errorf("expected 'data', got '%s'", head.Value)
		return nil
	}
	p.advance() // consume 'data'
	rnge := head.Range

	nameIdent, found := p.advanceExpected(tokens.Ident)
	if !found {
		return nil
	}
	name := nameIdent.Value
	alias := nameIdent.Value
	rnge.To = nameIdent.Range.To

	var typ_ ast.TypeRef
	if p.canExpect(tokens.PunctColon) {
		p.advance() // consume ':'
		typ_ = parseTypeRef(ctx, p)
		if typ_ == nil {
			return nil
		}
		rnge.To = typ_.Span().To
	}

	if p.canExpect(tokens.KeywordAs) {
		p.advance() // consume 'as'
		aliasIdent, found := p.advanceExpected(tokens.Ident)
		if !found {
			return nil
		}
		alias = aliasIdent.Value
		rnge.To = aliasIdent.Range.To
	}

	return ast.NewDataStatement(name, typ_, alias, rnge)
}
