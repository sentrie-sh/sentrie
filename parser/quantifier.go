// SPDX-License-Identifier: Apache-2.0
//
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
			parser.errorf("expected identifier for iterator, got %s at %s", valueIterator.Kind, valueIterator.Range.From)
			return nil
		}

		var indexIterator tokens.Instance
		// do we have a comma?
		if parser.head().IsOfKind(tokens.PunctComma) {
			// then we have an index iterator as well
			parser.advance()                                     // consume the comma
			idxit, found := parser.advanceExpected(tokens.Ident) // the index iterator token
			if !found {
				return nil
			}
			indexIterator = idxit
		}

		expression := parseBlockExpression(ctx, parser)
		if expression == nil {
			return nil
		}

		var quantifierExpr ast.Expression

		switch type_ {
		case tokens.KeywordAny:
			quantifierExpr = ast.NewAnyExpression(collection, valueIterator.Value, indexIterator.Value, expression, token.Range)
		case tokens.KeywordAll:
			quantifierExpr = ast.NewAllExpression(collection, valueIterator.Value, indexIterator.Value, expression, token.Range)
		case tokens.KeywordFilter:
			quantifierExpr = ast.NewFilterExpression(collection, valueIterator.Value, indexIterator.Value, expression, token.Range)
		case tokens.KeywordFirst:
			quantifierExpr = ast.NewFirstExpression(collection, valueIterator.Value, indexIterator.Value, expression, token.Range)
		case tokens.KeywordMap:
			quantifierExpr = ast.NewMapExpression(collection, valueIterator.Value, indexIterator.Value, expression, token.Range)
		}

		return quantifierExpr
	}
}
