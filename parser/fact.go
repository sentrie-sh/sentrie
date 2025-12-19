// SPDX-License-Identifier: Apache-2.0

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

// 'fact' @ident ('?'?) ':' <type> ( 'as' @ident )? ( 'default' <expression> )?
// ? = optional (sets optional=true)
// Facts are always non-nullable
func parseFactStatement(ctx context.Context, p *Parser) ast.Statement {
	start := p.head()

	rnge := start.Range

	if !p.expect(tokens.KeywordFact) {
		return nil
	}

	nameIdent, found := p.advanceExpected(tokens.Ident)
	if !found {
		return nil
	}

	name := nameIdent.Value  // Set the fact name
	alias := nameIdent.Value // Set the fact alias
	rnge.To = nameIdent.Range.To

	optional := false

	if !p.canExpectAnyOf(tokens.PunctColon, tokens.TokenQuestion) {
		p.errorf("expected ':' or '?' after fact name at %s", rnge.String())
		return nil
	}

	// Parse optional modifier
	if p.canExpect(tokens.TokenQuestion) {
		p.advance() // consume '?'
		optional = true
		rnge.To = p.head().Range.To
	}

	if !p.expect(tokens.PunctColon) {
		return nil
	}

	typ_ := parseTypeRef(ctx, p)
	if typ_ == nil {
		return nil
	}
	rnge.To = typ_.Span().To

	if p.canExpect(tokens.KeywordAs) {
		p.advance() // consume 'as'
		aliasIdent, found := p.advanceExpected(tokens.Ident)
		if !found {
			return nil
		}
		alias = aliasIdent.Value // Set the fact alias
		rnge.To = aliasIdent.Range.To
	}

	var defaultExpr ast.Expression
	if p.canExpect(tokens.KeywordDefault) {
		p.advance() // consume 'default'
		defaultExpr = p.parseExpression(ctx, LOWEST)
		if defaultExpr == nil {
			return nil
		}
		rnge.To = defaultExpr.Span().To
	}

	return ast.NewFactStatement(name, typ_, alias, defaultExpr, optional, rnge)
}
