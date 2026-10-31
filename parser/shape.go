// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"context"
	"log/slog"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
)

func parseShapeStatement(ctx context.Context, p *Parser) ast.Statement {
	shapeToken, found := p.advanceExpected(tokens.KeywordShape)
	if !found {
		return nil
	}
	rnge := shapeToken.Range

	nameToken, found := p.advanceExpected(tokens.Ident)
	if !found {
		return nil
	}

	name := nameToken.Value
	rnge.To = nameToken.Range.To
	if name == ast.AnyTypeName {
		p.errorAtf(nameToken.Range, "cannot declare a shape named %q: it is the built-in top type", name)
		return nil
	}

	var simpleTypeRef ast.TypeRef
	var complexShape *ast.Cmplx
	if p.canExpectAnyOf(tokens.PunctLeftCurly, tokens.KeywordWith) {
		complexShape = parseComplexShape(ctx, p)
	} else {
		simpleTypeRef = parseTypeRef(ctx, p)
	}

	if simpleTypeRef == nil && complexShape == nil /* both cannot be nil */ {
		return nil
	}

	if complexShape != nil {
		rnge.To = complexShape.Range.To
	} else {
		rnge.To = simpleTypeRef.Span().To
	}

	return ast.NewShapeStatement(name, simpleTypeRef, complexShape, rnge)
}

func parseComplexShape(ctx context.Context, p *Parser) *ast.Cmplx {
	stmt := &ast.Cmplx{
		Range:  p.head().Range,
		Fields: make(map[string]*ast.ShapeField),
	}

	if p.head().IsOfKind(tokens.KeywordWith) {
		p.advance()
		with := parseFQN(ctx, p)
		if with == nil {
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

	rCurly, found := p.advanceExpected(tokens.PunctRightCurly)
	if !found {
		return nil
	}

	// Update the end position to the closing curly brace
	stmt.Range.To = rCurly.Range.To

	return stmt
}

func parseShapeField(ctx context.Context, p *Parser) *ast.ShapeField {
	slog.DebugContext(ctx, "parseShapeField_start", "head", p.head().String())
	defer slog.DebugContext(ctx, "parseShapeField_end")

	field := &ast.ShapeField{
		Range: p.head().Range,
	}

	name, found := p.advanceExpected(tokens.Ident)
	if !found {
		return nil
	}
	field.Name = name.Value

	field.Optional = false
	if p.canExpect(tokens.TokenQuestion) {
		p.advance()
		field.Optional = true
	}

	if p.canExpect(tokens.TokenBang) {
		if field.Optional {
			p.errorf("legacy shape syntax '%s?!: T' is no longer supported; use '%s?: T?'", field.Name, field.Name)
			return nil
		}
		if p.peek().IsOfKind(tokens.TokenQuestion) {
			p.errorf("legacy shape syntax '%s!?: T' is no longer supported; use '%s?: T?'", field.Name, field.Name)
			return nil
		}
		p.errorf("legacy shape syntax '%s!: T' is no longer supported; use '%s: T'", field.Name, field.Name)
		return nil
	}

	if !p.expect(tokens.PunctColon) {
		return nil
	}

	field.Type = parseTypeRef(ctx, p)
	if p.err != nil {
		return nil
	}

	field.Range.To = field.Type.Span().To

	return field
}
