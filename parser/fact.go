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

// 'fact' exposed_name:type as internal_name
func parseFactStatement(ctx context.Context, p *Parser) ast.Statement {
	stmt := &ast.FactStatement{
		Pos: p.head().Position,
	}

	if !p.expect(tokens.KeywordFact) {
		return nil
	}

	name, found := p.advanceExpected(tokens.Ident)
	if !found {
		return nil
	}

	stmt.Name = name.Value  // Set the fact name
	stmt.Alias = name.Value // Set the fact alias

	if !p.expect(tokens.PunctColon) {
		return nil
	}

	typ_ := parseTypeRef(ctx, p)
	if typ_ == nil {
		return nil
	}
	stmt.Type = typ_

	if p.canExpect(tokens.KeywordAs) {
		p.advance() // consume 'as'
		alias, found := p.advanceExpected(tokens.Ident)
		if !found {
			return nil
		}
		stmt.Alias = alias.Value // Set the fact alias
	}

	if p.canExpect(tokens.KeywordDefault) {
		p.advance() // consume 'default'
		defaultExpr := p.parseExpression(ctx, LOWEST)
		if defaultExpr == nil {
			return nil
		}
		stmt.Default = defaultExpr
	}

	return stmt
}
