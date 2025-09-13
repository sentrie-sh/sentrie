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
		key, found := p.advanceExpected(tokens.String)
		if !found {
			return nil // Error in expecting a key
		}

		if !p.expect(tokens.PunctColon) {
			return nil // Error in expecting colon
		}

		value := p.parseExpression(ctx, LOWEST)
		if value == nil {
			return nil // Error in parsing a value
		}

		theMap.Entries = append(theMap.Entries, ast.MapEntry{
			Key:   key.Value,
			Value: value,
		})

		if p.current.Kind == tokens.PunctComma {
			p.advance() // Consume the comma
		}
		if p.head().IsOfKind(tokens.PunctRightCurly) {
			break // Exit if we reach the end of the map
		}
	}

	// Expect the closing curly brace
	if !p.expect(tokens.PunctRightCurly) {
		return nil
	}

	return &theMap
}
