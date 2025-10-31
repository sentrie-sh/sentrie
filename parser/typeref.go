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
		ref = ast.NewStringTypeRef(p.advance().Range)
	case tokens.KeywordNumber:
		ref = ast.NewNumberTypeRef(p.advance().Range)
	case tokens.KeywordBoolean:
		ref = ast.NewBoolTypeRef(p.advance().Range)
	case tokens.Ident:
		fqn := parseFQN(ctx, p)
		if fqn == nil {
			return nil
		}
		ref = ast.NewShapeTypeRef(fqn, fqn.Rnge)
	case tokens.KeywordList:
		ref = ast.NewListTypeRef(nil, p.advance().Range) // elemType will be set later
	case tokens.KeywordMap:
		ref = ast.NewMapTypeRef(nil, p.advance().Range) // valueType will be set later
	case tokens.KeywordRecord:
		ref = ast.NewRecordTypeRef(nil, p.advance().Range) // fields will be set later
	case tokens.KeywordDocument:
		ref = ast.NewDocumentTypeRef(p.advance().Range)
	}

	if ref == nil {
		return nil
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
		r.Rnge.To = rBracket.Range.To
	} else if r, ok := ref.(*ast.MapTypeRef); ok {
		if !p.expect(tokens.PunctLeftBracket) {
			return nil
		}
		r.ValueType = parseTypeRef(ctx, p)
		rBracket, found := p.advanceExpected(tokens.PunctRightBracket)
		if !found {
			return nil
		}
		r.Rnge.To = rBracket.Range.To
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
		r.Rnge.To = rBracket.Range.To
	}

	for p.head().IsOfKind(tokens.TokenAt) {
		constraint := parseTypeRefConstraint(ctx, p, ref)
		if constraint == nil {
			return nil
		}
		if err := ref.AddConstraint(constraint); err != nil {
			p.errorf("cannot add constraint %s: %s at %s", constraint.Name, err, constraint.Rnge)
			return nil
		}
	}

	return ref
}

func parseTypeRefConstraint(ctx context.Context, p *Parser, _ ast.TypeRef) *ast.TypeRefConstraint {
	slog.DebugContext(ctx, "parseTypeRefConstraint_start", "head", p.head().String())
	defer slog.DebugContext(ctx, "parseTypeRefConstraint_end")

	atToken, found := p.advanceExpected(tokens.TokenAt)
	if !found {
		return nil
	}

	rnge := atToken.Range

	name, found := p.advanceExpected(tokens.Ident)
	if !found {
		return nil
	}

	if !p.expect(tokens.PunctLeftParentheses) {
		return nil
	}

	args := []ast.Expression{}

	for !p.head().IsOfKind(tokens.PunctRightParentheses) {
		arg := parseConstraintLiteral(ctx, p)
		if arg == nil {
			return nil
		}

		args = append(args, arg)

		if p.head().IsOfKind(tokens.PunctComma) {
			p.advance()
		}
	}

	rightParentheses, found := p.advanceExpected(tokens.PunctRightParentheses)
	if !found {
		return nil
	}

	rnge.To = rightParentheses.Range.To

	return ast.NewTypeRefConstraint(name.Value, args, rnge)
}
