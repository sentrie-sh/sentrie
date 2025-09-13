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

// 'reduce' @collection 'from' @startExpression 'as' @accumulator,@valueIterator(,@indexIterator)? { @expression }
func parseReduceExpression(ctx context.Context, parser *Parser) ast.Expression {
	expr := &ast.ReduceExpression{
		Pos: parser.head().Position,
	}

	parser.advance() // the 'reduce' token

	collection := parser.parseExpression(ctx, LOWEST)
	if collection == nil {
		return nil
	}
	expr.Collection = collection

	if !parser.expect(tokens.KeywordFrom) {
		return nil
	}

	startExpression := parser.parseExpression(ctx, LOWEST)
	if startExpression == nil {
		return nil
	}
	expr.From = startExpression

	if !parser.expect(tokens.KeywordAs) {
		return nil
	}

	accumulator, found := parser.advanceExpected(tokens.Ident)
	if !found {
		return nil
	}
	expr.Accumulator = accumulator.Value

	if !parser.expect(tokens.PunctComma) {
		return nil
	}

	valueIterator := parser.advance() // the iterator token
	if valueIterator.Kind != tokens.Ident {
		parser.errorf("expected identifier for iterator, got %s at %s", valueIterator.Kind, valueIterator.Position)
		return nil
	}
	expr.ValueIterator = valueIterator.Value

	// do we have a comma?
	if parser.head().IsOfKind(tokens.PunctComma) {
		// then we have an index iterator as well
		parser.advance()                                     // consume the comma
		idxIt, found := parser.advanceExpected(tokens.Ident) // the index iterator token
		if !found {
			return nil
		}
		expr.IndexIterator = idxIt.Value
	}

	blockExpr := parseBlockExpression(ctx, parser)
	if blockExpr == nil {
		return nil
	}

	expr.Reducer = blockExpr

	return expr
}
