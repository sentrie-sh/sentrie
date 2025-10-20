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

func parseTypeRef(ctx context.Context, p *Parser) ast.TypeRef {
	validTypes := []tokens.Kind{tokens.Ident}
	validTypes = append(validTypes, PRIMITIVE_TYPES...)
	validTypes = append(validTypes, AGGREGATE_TYPES...)

	if !p.canExpectAnyOf(validTypes...) {
		p.errorf("expected one of %v, got %s", validTypes, p.head().Kind)
		return nil
	}

	var ref ast.TypeRef
	switch p.current.Kind {
	case tokens.KeywordString:
		ref = &ast.StringTypeRef{
			Range: p.advance().Range,
		}
	case tokens.KeywordNumber:
		ref = &ast.NumberTypeRef{
			Range: p.advance().Range,
		}
	case tokens.KeywordBoolean:
		ref = &ast.BoolTypeRef{
			Range: p.advance().Range,
		}
	case tokens.Ident:
		ref = &ast.ShapeTypeRef{
			Range: p.head().Range, // we cannot advance here, since this is a FQN which needs to be parsed
			Ref:   func() ast.FQN { f, _ := parseFQN(ctx, p); return f }(),
		}
	case tokens.KeywordList:
		ref = &ast.ListTypeRef{
			Range: p.advance().Range,
		}
	case tokens.KeywordMap:
		ref = &ast.MapTypeRef{
			Range: p.advance().Range,
		}
	case tokens.KeywordRecord:
		ref = &ast.RecordTypeRef{
			Range: p.advance().Range,
		}
	case tokens.KeywordDocument:
		ref = &ast.DocumentTypeRef{
			Range: p.advance().Range,
		}
	}

	if r, ok := ref.(*ast.ListTypeRef); ok {
		if !p.expect(tokens.PunctLeftBracket) {
			return nil
		}
		r.ElemType = parseTypeRef(ctx, p)
		rBracket, found := p.advanceExpected(tokens.PunctRightBracket)
		if !found {
			return nil
		}
		r.Range.To = rBracket.Range.To
	} else if r, ok := ref.(*ast.MapTypeRef); ok {
		if !p.expect(tokens.PunctLeftBracket) {
			return nil
		}
		r.ValueType = parseTypeRef(ctx, p)
		rBracket, found := p.advanceExpected(tokens.PunctRightBracket)
		if !found {
			return nil
		}
		r.Range.To = rBracket.Range.To
	} else if r, ok := ref.(*ast.RecordTypeRef); ok {
		if !p.expect(tokens.PunctLeftBracket) {
			return nil
		}
		for !p.head().IsOfKind(tokens.PunctRightBracket) {
			r.Fields = append(r.Fields, parseTypeRef(ctx, p))
			if p.head().IsOfKind(tokens.PunctComma) {
				p.advance()
			}
		}

		rBracket, found := p.advanceExpected(tokens.PunctRightBracket)
		if !found {
			return nil
		}
		r.Range.To = rBracket.Range.To
	}

	for p.head().IsOfKind(tokens.TokenAt) {
		constraint := parseTypeRefConstraint(ctx, p, ref)
		if constraint == nil {
			return nil
		}
		if err := ref.AddConstraint(constraint); err != nil {
			p.errorf("cannot add constraint %s: %s at %s", constraint.Name, err, constraint.Range)
			return nil
		}
	}

	return ref
}

func parseTypeRefConstraint(ctx context.Context, p *Parser, _ ast.TypeRef) *ast.TypeRefConstraint {
	slog.DebugContext(ctx, "parseShapeFieldConstraint_start", "head", p.head().String())
	defer slog.DebugContext(ctx, "parseShapeFieldConstraint_end")

	constraint := &ast.TypeRefConstraint{
		Range: p.head().Range,
	}

	if !p.expect(tokens.TokenAt) {
		return nil
	}

	name, found := p.advanceExpected(tokens.Ident)
	if !found {
		return nil
	}
	constraint.Name = name.Value

	if !p.expect(tokens.PunctLeftParentheses) {
		return nil
	}

	for !p.head().IsOfKind(tokens.PunctRightParentheses) {
		arg := parseConstraintLiteral(ctx, p)
		if arg == nil {
			return nil
		}

		constraint.Args = append(constraint.Args, arg)

		if p.head().IsOfKind(tokens.PunctComma) {
			p.advance()
		}
	}

	if !p.expect(tokens.PunctRightParentheses) {
		return nil
	}

	return constraint
}
