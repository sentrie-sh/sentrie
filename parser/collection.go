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

func parseListLiteral(ctx context.Context, p *Parser) ast.Expression {
	theList := ast.ListLiteral{
		Pos:    p.current.Position,
		Values: []ast.Expression{},
	}
	// Expect the opening bracket
	if !p.expect(tokens.PunctLeftBracket) {
		return &theList
	}

	if p.current.Kind == tokens.PunctRightBracket {
		p.advance() // Consume the closing bracket
		return &theList
	}

	// Parse the elements of the list
	for p.hasTokens() && p.current.Kind != tokens.PunctRightBracket {
		element := p.parseExpression(ctx, LOWEST)
		if element == nil {
			return nil // Error in parsing an element
		}

		theList.Values = append(theList.Values, element)

		if p.current.Kind == tokens.PunctComma {
			p.advance() // Consume the comma
		}
		if p.head().IsOfKind(tokens.PunctRightBracket) {
			break // Exit if we reach the end of the list
		}
	}

	// Expect the closing bracket
	if !p.expect(tokens.PunctRightBracket) {
		return nil
	}

	return &theList
}

func parseMapLiteral(ctx context.Context, p *Parser) ast.Expression {
	leftBrace := p.advance() // Consume the left curly brace

	theMap := ast.MapLiteral{
		Pos:     leftBrace.Position,
		Entries: []ast.MapEntry{},
	}

	// Parse the entries of the map
	for p.hasTokens() && p.current.Kind != tokens.PunctRightCurly {
		var keyExpression ast.Expression

		if p.canExpect(tokens.String) {
			key, found := p.advanceExpected(tokens.String)
			if !found {
				return nil
			}
			keyExpression = &ast.StringLiteral{
				Pos:   key.Position,
				Value: key.Value,
			}
		} else if p.canExpect(tokens.PunctLeftBracket) {
			if !p.expect(tokens.PunctLeftBracket) {
				return nil
			}
			keyExpression = p.parseExpression(ctx, LOWEST)
			if keyExpression == nil {
				return nil
			}
			if !p.expect(tokens.PunctRightBracket) {
				return nil
			}
		} else {
			p.errorf("expected string or [expression] as map key, got %s at %s", p.current.Kind, p.current.Position)
			return nil
		}

		if !p.expect(tokens.PunctColon) {
			return nil // Error in expecting colon
		}

		value := p.parseExpression(ctx, LOWEST)
		if value == nil {
			return nil // Error in parsing a value
		}

		entry := ast.MapEntry{
			Key:   keyExpression,
			Value: value,
		}
		theMap.Entries = append(theMap.Entries, entry)

		if p.current.Kind == tokens.PunctComma {
			p.advance() // Consume the comma
		}
	}

	// Expect the closing curly brace
	if !p.expect(tokens.PunctRightCurly) {
		return nil
	}

	return &theMap
}
