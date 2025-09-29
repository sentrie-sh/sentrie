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
	"strings"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
)

// 'use' func,func 'from' moduleName 'as' alias
func parseUseStatement(ctx context.Context, p *Parser) ast.Statement {
	pos := p.head().Position

	stmt := &ast.UseStatement{
		Pos: pos,
	}

	p.advance() // consume 'use'

	fns := []string{}

	firstModuleName, found := p.advanceExpected(tokens.Ident)
	if !found {
		return nil
	}
	fns = append(fns, firstModuleName.Value)

	for p.head().IsOfKind(tokens.PunctComma) {
		p.advance() // consume ','
		fn, found := p.advanceExpected(tokens.Ident)
		if !found {
			return nil
		}
		fns = append(fns, fn.Value)
	}

	stmt.Modules = fns

	if !p.expect(tokens.KeywordFrom) {
		return nil
	}

	if !p.canExpectAnyOf(tokens.String, tokens.TokenAt) {
		p.errorf("expected string or '@' for module import")
		return nil
	}

	if p.head().IsOfKind(tokens.String) {
		fromModule, found := p.advanceExpected(tokens.String)
		if !found {
			return nil
		}
		stmt.RelativeFrom = fromModule.Value // Set the module name
	} else {
		p.advance() // consume '@'
		fromPackage, found := p.advanceExpected(tokens.Ident)
		if !found {
			return nil
		}
		from := []string{fromPackage.Value}
		for p.head().IsOfKind(tokens.TokenDiv) {
			p.advance() // consume '/'
			fromModule, found := p.advanceExpected(tokens.Ident)
			if !found {
				return nil
			}
			from = append(from, fromModule.Value)
		}
		stmt.LibFrom = from
	}

	// default alias to the module name
	// for @foo/bar, default alias is bar - the last part of the path
	// for quoted strings, it's the last part of the path
	if stmt.RelativeFrom != "" {
		parts := strings.Split(stmt.RelativeFrom, "/")
		stmt.As = parts[len(parts)-1]
	} else {
		stmt.As = stmt.LibFrom[len(stmt.LibFrom)-1]
	}

	if p.canExpect(tokens.KeywordAs) {
		p.advance() // consume 'as'
		asAlias, found := p.advanceExpected(tokens.Ident)
		if !found {
			return nil
		}
		stmt.As = asAlias.Value // Set the alias
	}

	return stmt
}
