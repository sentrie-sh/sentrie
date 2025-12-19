// SPDX-License-Identifier: Apache-2.0

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

// '[' ( <expression> ( ',' <expression> )* )? ']'
func parseListLiteral(ctx context.Context, p *Parser) ast.Expression {
	leftBracket, found := p.advanceExpected(tokens.PunctLeftBracket)
	if !found {
		return nil
	}

	theList := []ast.Expression{}

	// Parse the elements of the list
	for p.hasTokens() && p.current.Kind != tokens.PunctRightBracket {
		element := p.parseExpression(ctx, LOWEST)
		if element == nil {
			return nil // Error in parsing an element
		}

		theList = append(theList, element)

		if p.current.Kind == tokens.PunctComma {
			p.advance() // Consume the comma
		}
		if p.head().IsOfKind(tokens.PunctRightBracket) {
			break // Exit if we reach the end of the list
		}
	}

	// Update the end position to the closing bracket
	rightBracket, found := p.advanceExpected(tokens.PunctRightBracket)
	if !found {
		return nil
	}

	listLiteral := ast.NewListLiteral(theList, tokens.Range{
		File: leftBracket.Range.File,
		From: leftBracket.Range.From,
		To:   rightBracket.Range.To,
	})

	return listLiteral
}

// '{' ( <string | '[' expression ']' > ':' <expression> ( ',' <string | '[' expression ']' > ':' <expression> )* )? '}'
func parseMapLiteral(ctx context.Context, p *Parser) ast.Expression {
	leftBrace := p.advance() // Consume the left curly brace

	entries := []ast.MapEntry{}

	// Parse the entries of the map
	for p.hasTokens() && p.current.Kind != tokens.PunctRightCurly {
		var keyExpression ast.Expression

		if p.canExpect(tokens.String) {

			key, found := p.advanceExpected(tokens.String)
			if !found {
				return nil
			}

			keyExpression = ast.NewStringLiteral(key.Value, tokens.Range{
				File: key.Range.File,
				From: key.Range.From,
				To:   key.Range.To,
			})

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
			p.errorf("expected string or [expression] as map key, got %s at %s", p.current.Kind, p.current.Range.From)
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
		entries = append(entries, entry)

		if p.current.Kind == tokens.PunctComma {
			p.advance() // Consume the comma
		}
	}

	// Expect the closing curly brace
	rightBrace, found := p.advanceExpected(tokens.PunctRightCurly)
	if !found {
		return nil
	}

	mapLiteral := ast.NewMapLiteral(entries, tokens.Range{
		File: leftBrace.Range.File,
		From: leftBrace.Range.From,
		To:   rightBrace.Range.To,
	})

	return mapLiteral
}
