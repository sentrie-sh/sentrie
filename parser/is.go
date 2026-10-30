// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"context"
	"slices"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
)

// 'is [not] defined' | 'is [not] empty' | 'is [not] <type>' | 'is [not] <expression>'
//
// A right-hand side that starts with a type keyword (string, number, boolean,
// trinary, document, list, dict, record) is a type and builds an
// IsExpression. A bare or slash-qualified name builds an IsExpression that
// carries both readings, for the index to resolve against the shapes in scope,
// unless a `?` or `@constraint` suffix marks it as a type. Any other
// right-hand side is the legacy equality synonym, an InfixExpression "is".
func parseIsExpression(ctx context.Context, p *Parser, left ast.Expression, precedence Precedence) ast.Expression {
	start := p.head()

	rnge := start.Range

	// consume the 'is' token
	if !p.expect(tokens.KeywordIs) {
		return nil
	}

	var not *tokens.Instance
	if p.head().IsOfKind(tokens.KeywordNot) {
		n := p.advance() // consume the 'not' token
		not = &n
	}

	var expr ast.Expression

	if p.canExpect(tokens.KeywordDefined) {
		// 'is [not] defined' case
		expr = ast.NewIsDefinedExpression(left, rnge)
		p.advance()
	} else if p.canExpect(tokens.KeywordEmpty) {
		expr = ast.NewIsEmptyExpression(left, rnge)
		p.advance()
	} else if p.canExpectAnyOf(isTypeKeywords...) {
		typ := parseBaseTypeRef(ctx, p)
		if typ == nil {
			return nil
		}
		typ = parseTypeRefSuffixes(ctx, p, typ, true)
		if typ == nil {
			return nil
		}
		rnge.To = typ.Span().To
		expr = ast.NewIsExpression(left, nil, typ, isSpan(left, rnge))
	} else {
		right := p.parseExpression(ctx, precedence)
		if right == nil {
			return nil
		}
		rnge.To = right.Span().To
		if parts, ok := ast.FlattenSlashIdentChain(right); ok {
			fqn := ast.NewFQN(parts, right.Span())
			var typ ast.TypeRef = ast.NewShapeTypeRef(&fqn, right.Span())
			// A name may be a value, so `?` before `{` stays the ternary here
			// (`x is y ? {...} : {...}`), unlike after a type keyword.
			if p.canExpect(tokens.TokenAt) || (p.canExpect(tokens.TokenQuestion) && p.questionIsNullable(false)) {
				// a suffix only a type can carry: this is a shape, not a value
				typ = parseTypeRefSuffixes(ctx, p, typ, true)
				if typ == nil {
					return nil
				}
				rnge.To = typ.Span().To
				expr = ast.NewIsExpression(left, nil, typ, isSpan(left, rnge))
			} else {
				expr = ast.NewIsExpression(left, right, typ, isSpan(left, rnge))
			}
		} else {
			expr = ast.NewInfixExpression(left, right, start.Value, rnge)
		}
	}

	// if we have a 'not' then wrap with a not unary
	if not != nil {
		expr = ast.NewUnaryExpression(not.Value, expr, rnge)
	}

	return expr
}

// isTypeKeywords lists the tokens that can only start a type, never an
// expression, on the right of `is`.
var isTypeKeywords = append(slices.Clone(PRIMITIVE_TYPES), AGGREGATE_TYPES...)

// isSpan runs from the left operand to the end of the right-hand side.
func isSpan(left ast.Expression, rest tokens.Range) tokens.Range {
	if left == nil {
		return rest
	}
	rest.From = left.Span().From
	return rest
}
