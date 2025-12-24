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

/*
*

	{
		let statement = some_expression
		-- a comment statement
		yield another_expression -- must be the last statement
	}

*
*/
func parseBlockExpression(ctx context.Context, p *Parser) ast.Expression {
	lCurly, found := p.advanceExpected(tokens.PunctLeftCurly)
	if !found {
		return nil // Error in parsing the block expression
	}

	var statements []ast.Statement

	for p.canExpectAnyOf(tokens.KeywordLet, tokens.LineComment) {
		stmt := parsePolicyStatement(ctx, p)
		if stmt == nil {
			return nil // Error in parsing the block expression
		}
		statements = append(statements, stmt)
	}

	if !p.expect(tokens.KeywordYield) {
		return nil // Error in parsing the block expression
	}

	yieldExpr := p.parseExpression(ctx, LOWEST)
	if yieldExpr == nil {
		return nil // Error in parsing the block expression
	}

	rCurly, found := p.advanceExpected(tokens.PunctRightCurly)
	if !found {
		return nil // Error in parsing the block expression
	}

	return ast.NewBlockExpression(statements, yieldExpr, tokens.Range{
		File: lCurly.Range.File,
		From: tokens.Pos{
			Line:   lCurly.Range.From.Line,
			Column: lCurly.Range.From.Column,
			Offset: lCurly.Range.From.Offset,
		},
		To: tokens.Pos{
			Line:   rCurly.Range.From.Line,
			Column: rCurly.Range.From.Column,
			Offset: rCurly.Range.From.Offset,
		},
	})
}
