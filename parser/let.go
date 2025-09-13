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

func parseLetsStatement(ctx context.Context, p *Parser) ast.Statement {
	stmt := &ast.VarDeclaration{
		Pos: p.head().Position,
	}
	p.advance() // consume 'let'

	name, found := p.advanceExpected(tokens.Ident)
	if !found {
		return nil
	}
	stmt.Name = name.Value

	if p.current.Kind == tokens.PunctColon {
		p.advance() // consume ':'
		typeRef := parseTypeRef(ctx, p)
		if typeRef == nil {
			return nil
		}
		stmt.Type = typeRef
	}

	if !p.expect(tokens.TokenAssign) { // expect '='
		return nil
	}

	val := p.parseExpression(ctx, LOWEST)
	if val == nil {
		return nil
	}

	stmt.Value = val

	return stmt
}
