// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"context"
	"log/slog"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
)

func parseTypeRef(ctx context.Context, p *Parser) ast.TypeRef {
	ref := parseBaseTypeRef(ctx, p)
	if ref == nil {
		return nil
	}
	return parseTypeRefSuffixes(ctx, p, ref, false)
}

// parseBaseTypeRef parses a type reference without its trailing `?` and
// `@constraint(...)` suffixes.
func parseBaseTypeRef(ctx context.Context, p *Parser) ast.TypeRef {
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
	case tokens.KeywordBoolean, tokens.KeywordTrinary:
		ref = ast.NewTrinaryTypeRef(p.advance().Range)
	case tokens.Ident:
		fqn := parseFQN(ctx, p)
		if fqn == nil {
			return nil
		}
		ref = ast.NewShapeTypeRef(fqn, fqn.Rnge)
	case tokens.KeywordList:
		ref = ast.NewListTypeRef(nil, p.advance().Range) // elemType will be set later
	case tokens.KeywordDict:
		ref = ast.NewDictTypeRef(nil, p.advance().Range) // valueType will be set later
	case tokens.KeywordRecord:
		ref = ast.NewRecordTypeRef(nil, p.advance().Range) // fields will be set later
	case tokens.KeywordDocument:
		ref = ast.NewDocumentTypeRef(p.advance().Range)
	}

	if r, ok := ref.(*ast.ListTypeRef); ok {
		if !p.expect(tokens.PunctLeftBracket) {
			return nil
		}
		elemType := parseTypeRef(ctx, p)
		if elemType == nil {
			return nil
		}
		r.ElemType = elemType
		rBracket, found := p.advanceExpected(tokens.PunctRightBracket)
		if !found {
			return nil
		}
		r.Rnge.To = rBracket.Range.To
	} else if r, ok := ref.(*ast.DictTypeRef); ok {
		if !p.expect(tokens.PunctLeftBracket) {
			return nil
		}
		valueType := parseTypeRef(ctx, p)
		if valueType == nil {
			return nil
		}
		r.ValueType = valueType
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
			fieldType := parseTypeRef(ctx, p)
			if fieldType == nil {
				return nil
			}
			r.Fields = append(r.Fields, fieldType)
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

	return ref
}

// parseTypeRefSuffixes parses the optional nullable `?` and the constraints
// that follow a type. With ternaryAware set, a `?` is the nullable suffix only
// when questionIsNullable (with `{` read as a rule body) says so; otherwise it
// is left for the ternary operator, which is what `x is string ? a : b` means
// (the `is` operand sits above TERNARY precedence).
func parseTypeRefSuffixes(ctx context.Context, p *Parser, ref ast.TypeRef, ternaryAware bool) ast.TypeRef {
	if p.canExpect(tokens.TokenQuestion) && (!ternaryAware || p.questionIsNullable(true)) {
		question := p.advance()
		rnge := ref.Span()
		rnge.To = question.Range.To
		ref = ast.NewNullableTypeRef(ref, rnge)
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
