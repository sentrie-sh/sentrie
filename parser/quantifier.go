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

func quantifierParserFactory(type_ tokens.Kind) prefixParser {
	// 'quantifier' @collection 'as' @iterator { @expression }
	return func(ctx context.Context, parser *Parser) ast.Expression {
		token := parser.advance() // the quantifier token

		collection := parser.parseExpression(ctx, LOWEST)
		if collection == nil {
			return nil
		}

		if !parser.expect(tokens.KeywordAs) {
			return nil
		}

		valueIterator := parser.advance() // the iterator token
		if valueIterator.Kind != tokens.Ident {
			parser.errorf("expected identifier for iterator, got %s at %s", valueIterator.Kind, valueIterator.Position)
			return nil
		}

		var indexIterator string
		// do we have a comma?
		if parser.head().IsOfKind(tokens.PunctComma) {
			// then we have an index iterator as well
			parser.advance()                                     // consume the comma
			idxIt, found := parser.advanceExpected(tokens.Ident) // the index iterator token
			if !found {
				return nil
			}
			indexIterator = idxIt.Value
		}

		expression := parseBlockExpression(ctx, parser)
		if expression == nil {
			return nil
		}

		var quantifierExpr ast.Expression

		switch type_ {
		case tokens.KeywordAny:
			quantifierExpr = &ast.AnyExpression{
				Pos:           token.Position,
				Collection:    collection,
				ValueIterator: valueIterator.Value,
				IndexIterator: indexIterator,
				Predicate:     expression,
			}
		case tokens.KeywordAll:
			quantifierExpr = &ast.AllExpression{
				Pos:           token.Position,
				Collection:    collection,
				ValueIterator: valueIterator.Value,
				IndexIterator: indexIterator,
				Predicate:     expression,
			}
		case tokens.KeywordFilter:
			quantifierExpr = &ast.FilterExpression{
				Pos:           token.Position,
				Collection:    collection,
				ValueIterator: valueIterator.Value,
				IndexIterator: indexIterator,
				Predicate:     expression,
			}
		case tokens.KeywordMap:
			quantifierExpr = &ast.MapExpression{
				Pos:           token.Position,
				Collection:    collection,
				ValueIterator: valueIterator.Value,
				IndexIterator: indexIterator,
				Transform:     expression,
			}
		case tokens.KeywordDistinct:
			quantifierExpr = &ast.DistinctExpression{
				Pos:           token.Position,
				Collection:    collection,
				ValueIterator: valueIterator.Value,
				IndexIterator: indexIterator,
				Predicate:     expression,
			}
		}

		return quantifierExpr
	}
}
