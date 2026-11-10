// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"context"
	"errors"
	"strconv"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
)

// parseConstraintLiteral parses only literal expressions for constraint arguments
// This ensures constraints can only use compile-time constants, not runtime expressions
func parseConstraintLiteral(ctx context.Context, p *Parser) ast.Expression {
	switch p.current.Kind {
	case tokens.String:
		return parseStringLiteral(ctx, p)
	case tokens.Int:
		return parseIntegerValueLiteral(ctx, p)
	case tokens.Float:
		return parseFloatLiteral(ctx, p)
	case tokens.KeywordTrue, tokens.KeywordFalse, tokens.KeywordUnknown:
		return parseTrinaryLiteral(ctx, p)
	case tokens.KeywordNull:
		return parseNullLiteral(ctx, p)
	case tokens.TokenMinus:
		return parseConstraintNegativeLiteral(ctx, p)
	case tokens.PunctLeftBracket:
		return parseConstraintListLiteral(ctx, p)
	case tokens.PunctLeftCurly:
		return parseConstraintMapLiteral(ctx, p)
	case tokens.Error:
		// A lexical error such as a malformed number literal: report the
		// lexer's message rather than calling it a non-literal.
		p.errorf("%s", p.current.Value)
		return nil
	default:
		p.errorf("constraint arguments must be literals, got %s", p.current.Kind)
		return nil
	}
}

// parseConstraintNegativeLiteral parses a unary minus followed by a number
// literal (`@min(-5)`, `@gt(-0.5)`) into one negative literal spanning both
// tokens. The minus applies to number literals only: `-x`, `-"a"` and
// `-(1)` stay errors, so constraint arguments remain literals.
func parseConstraintNegativeLiteral(ctx context.Context, p *Parser) ast.Expression {
	minus := p.advance()
	number := p.current
	rnge := tokens.Range{File: minus.Range.File, From: minus.Range.From, To: number.Range.To}
	switch number.Kind {
	case tokens.Int:
		p.advance()
		text := "-" + number.Value
		value, err := strconv.ParseInt(text, 10, 64)
		if err == nil {
			return ast.NewIntegerLiteral(value, rnge)
		}
		// Below the int64 range: a float, as for a positive literal (#166).
		if f, ferr := strconv.ParseFloat(text, 64); errors.Is(err, strconv.ErrRange) && ferr == nil {
			return ast.NewFloatLiteral(f, rnge)
		}
		p.errorAtf(number.Range, "invalid integer literal %q: %w", number.Value, err)
		return nil
	case tokens.Float:
		p.advance()
		value, err := strconv.ParseFloat("-"+number.Value, 64)
		if err != nil {
			p.errorAtf(number.Range, "invalid float literal %q: %w", number.Value, err)
			return nil
		}
		return ast.NewFloatLiteral(value, rnge)
	default:
		p.errorf("a minus in a constraint argument must be followed by a number literal, got %s", number.Kind)
		return nil
	}
}

// parseConstraintListLiteral parses a list literal for constraint arguments (literal-only)
func parseConstraintListLiteral(ctx context.Context, p *Parser) ast.Expression {
	lBracket, found := p.advanceExpected(tokens.PunctLeftBracket)
	if !found {
		return nil
	}

	var elements []ast.Expression

	// Parse list elements (only literals)
	for p.hasTokens() && p.current.Kind != tokens.PunctRightBracket {
		element := parseConstraintLiteral(ctx, p)
		if element == nil {
			return nil
		}
		elements = append(elements, element)

		if p.head().IsOfKind(tokens.PunctRightBracket) {
			break
		}

		if !p.expect(tokens.PunctComma) {
			return nil
		}

	}

	rBracket, found := p.advanceExpected(tokens.PunctRightBracket)
	if !found {
		return nil
	}

	return ast.NewListLiteral(elements, tokens.Range{
		File: lBracket.Range.File,
		From: lBracket.Range.From,
		To:   rBracket.Range.To,
	})
}

// parseConstraintMapLiteral parses a map literal for constraint arguments (literal-only)
func parseConstraintMapLiteral(ctx context.Context, p *Parser) ast.Expression {
	lCurly, found := p.advanceExpected(tokens.PunctLeftCurly)
	if !found {
		return nil
	}

	var entries []ast.MapEntry

	// Parse map entries (only literals)
	for p.hasTokens() && p.current.Kind != tokens.PunctRightCurly {
		// Parse key (must be string literal)
		if !p.canExpect(tokens.String) {
			p.errorf("map keys must be string literals, got %s", p.current.Kind)
			return nil
		}
		keyToken, found := p.advanceExpected(tokens.String)
		if !found {
			return nil
		}

		// Expect colon
		if !p.canExpect(tokens.PunctColon) {
			return nil
		}

		// Parse value (must be literal)
		value := parseConstraintLiteral(ctx, p)
		if value == nil {
			return nil
		}

		entries = append(entries, ast.MapEntry{
			Key:   ast.NewStringLiteral(keyToken.Value, keyToken.Range),
			Value: value,
		})

		if p.head().IsOfKind(tokens.PunctRightCurly) {
			break
		}

		if !p.expect(tokens.PunctComma) {
			return nil
		}
	}

	rCurly, found := p.advanceExpected(tokens.PunctRightCurly)
	if !found {
		return nil
	}

	return ast.NewMapLiteral(entries, tokens.Range{
		File: lCurly.Range.File,
		From: lCurly.Range.From,
		To:   rCurly.Range.To,
	})
}
