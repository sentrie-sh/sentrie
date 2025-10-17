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
		return parseIntegerLiteral(ctx, p)
	case tokens.Float:
		return parseFloatLiteral(ctx, p)
	case tokens.KeywordTrue, tokens.KeywordFalse, tokens.KeywordUnknown:
		return parseTrinaryLiteral(ctx, p)
	case tokens.KeywordNull:
		return parseNullLiteral(ctx, p)
	case tokens.PunctLeftBracket:
		return parseConstraintListLiteral(ctx, p)
	case tokens.PunctLeftCurly:
		return parseConstraintMapLiteral(ctx, p)
	default:
		p.errorf("constraint arguments must be literals, got %s at %s", p.current.Kind, p.current.Position)
		return nil
	}
}

// parseConstraintListLiteral parses a list literal for constraint arguments (literal-only)
func parseConstraintListLiteral(ctx context.Context, p *Parser) ast.Expression {
	token := p.advance() // consume '['

	var elements []ast.Expression

	// Handle empty list
	if p.current.Kind == tokens.PunctRightBracket {
		p.advance() // consume ']'
		return &ast.ListLiteral{
			Pos:    token.Position,
			Values: elements,
		}
	}

	// Parse list elements (only literals)
	for {
		element := parseConstraintLiteral(ctx, p)
		if element == nil {
			return nil
		}
		elements = append(elements, element)

		if p.current.Kind == tokens.PunctComma {
			p.advance() // consume ','
		} else if p.current.Kind == tokens.PunctRightBracket {
			break
		} else {
			p.errorf("expected ',' or ']' in list literal, got %s at %s", p.current.Kind, p.current.Position)
			return nil
		}
	}

	if !p.expect(tokens.PunctRightBracket) {
		return nil
	}

	return &ast.ListLiteral{
		Pos:    token.Position,
		Values: elements,
	}
}

// parseConstraintMapLiteral parses a map literal for constraint arguments (literal-only)
func parseConstraintMapLiteral(ctx context.Context, p *Parser) ast.Expression {
	token := p.advance() // consume '{'

	var entries []ast.MapEntry

	// Handle empty map
	if p.current.Kind == tokens.PunctRightCurly {
		p.advance() // consume '}'
		return &ast.MapLiteral{
			Pos:     token.Position,
			Entries: entries,
		}
	}

	// Parse map entries (only literals)
	for {
		// Parse key (must be string literal)
		if p.current.Kind != tokens.String {
			p.errorf("map keys must be string literals, got %s at %s", p.current.Kind, p.current.Position)
			return nil
		}
		keyToken := p.advance()

		// Expect colon
		if !p.expect(tokens.PunctColon) {
			return nil
		}

		// Parse value (must be literal)
		value := parseConstraintLiteral(ctx, p)
		if value == nil {
			return nil
		}

		entries = append(entries, ast.MapEntry{
			Key: &ast.StringLiteral{
				Pos:   keyToken.Position,
				Value: keyToken.Value,
			},
			Value: value,
		})

		if p.current.Kind == tokens.PunctComma {
			p.advance() // consume ','
		} else if p.current.Kind == tokens.PunctRightCurly {
			break
		} else {
			p.errorf("expected ',' or '}' in map literal, got %s at %s", p.current.Kind, p.current.Position)
			return nil
		}
	}

	if !p.expect(tokens.PunctRightCurly) {
		return nil
	}

	return &ast.MapLiteral{
		Pos:     token.Position,
		Entries: entries,
	}
}
