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

// distinct <collection> as <leftIterator>, <rightIterator> <block_predicate>
func parseDistinctExpression(ctx context.Context, parser *Parser) ast.Expression {
	head := parser.head()

	collection := parser.parseExpression(ctx, LOWEST)
	if collection == nil {
		return nil
	}

	if !parser.expect(tokens.KeywordAs) {
		return nil
	}

	leftIterator, found := parser.advanceExpected(tokens.Ident) // the iterator token
	if !found {
		return nil
	}

	if !parser.expect(tokens.PunctComma) {
		return nil
	}

	rightIterator, found := parser.advanceExpected(tokens.Ident) // the iterator token
	if !found {
		return nil
	}

	predicateBlock := parseBlockExpression(ctx, parser)
	if predicateBlock == nil {
		return nil
	}

	return ast.NewDistinctExpression(collection, leftIterator.Value, rightIterator.Value, predicateBlock, tokens.Range{
		File: head.Range.File,
		From: head.Range.From,
		To:   predicateBlock.Span().To,
	})
}
