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
	"log/slog"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
)

func parseShapeStatement(ctx context.Context, p *Parser) ast.Statement {
	stmt := &ast.ShapeStatement{
		Pos: p.head().Position,
	}

	if !p.expect(tokens.KeywordShape) {
		return nil
	}

	name, found := p.advanceExpected(tokens.Ident)
	if !found {
		return nil
	}

	stmt.Name = name.Value

	if p.canExpectAnyOf(tokens.PunctLeftCurly, tokens.KeywordWith) {
		stmt.Complex = parseComplexShape(ctx, p)
	} else {
		stmt.Simple = parseTypeRef(ctx, p)
	}

	if stmt.Simple == nil && stmt.Complex == nil /* both cannot be nil */ {
		return nil
	}

	return stmt
}

func parseComplexShape(ctx context.Context, p *Parser) *ast.Cmplx {
	stmt := &ast.Cmplx{
		Pos:    p.head().Position,
		Fields: make(map[string]*ast.ShapeField),
	}

	if p.head().IsOfKind(tokens.KeywordWith) {
		p.advance()
		with := parseFQN(ctx, p)
		if len(with) == 0 {
			return nil
		}
		stmt.With = with
	}

	if !p.expect(tokens.PunctLeftCurly) {
		return nil
	}

	for !p.head().IsOfKind(tokens.PunctRightCurly) {
		field := parseShapeField(ctx, p)
		if field == nil {
			return nil
		}
		stmt.Fields[field.Name] = field

		// consume trailing comments
		for p.canExpectAnyOf(tokens.TrailingComment, tokens.LineComment) {
			p.advance()
		}
	}

	if !p.expect(tokens.PunctRightCurly) {
		return nil
	}

	return stmt
}

func parseShapeField(ctx context.Context, p *Parser) *ast.ShapeField {
	slog.DebugContext(ctx, "parseShapeField_start", "head", p.head().String())
	defer slog.DebugContext(ctx, "parseShapeField_end")

	field := &ast.ShapeField{
		Pos: p.head().Position,
	}

	name, found := p.advanceExpected(tokens.Ident)
	if !found {
		return nil
	}
	field.Name = name.Value

	// Parse field modifiers (! and ?)
	// Default: Required field that can be null
	field.Required = true
	field.NotNullable = false

	/*
		Field modifier combinations:
		- Default: Required field that can be null
		- `!`: Required field that cannot be null
		- `?`: Optional field that can be null
		- `!?`: Optional field that cannot be null (if present)
		- `?!`: Same as `!?` (order doesn't matter)

		Examples:
		name!: string           -- Required, cannot be null
		age: int                -- Required, can be null
		email?: string          -- Optional, can be omitted
		phone!?: string         -- Optional, but if present cannot be null
		phone?!: string         -- Same as above
	*/

	// Parse modifiers (both can be present)
	for p.canExpectAnyOf(tokens.TokenBang, tokens.TokenQuestion) {
		if p.head().IsOfKind(tokens.TokenBang) {
			field.NotNullable = true
		} else if p.head().IsOfKind(tokens.TokenQuestion) {
			field.Required = false
		}
		p.advance()
	}

	if !p.expect(tokens.PunctColon) {
		return nil
	}

	field.Type = parseTypeRef(ctx, p)
	if p.err != nil {
		return nil
	}

	return field
}
