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
	expr := &ast.BlockExpression{
		Pos: p.head().Position,
	}
	if !p.expect(tokens.PunctLeftCurly) {
		return nil // Error in parsing the block expression
	}

	var statements []ast.Statement

	for p.canExpectAnyOf(tokens.KeywordLet, tokens.LineComment) {
		stmt := parseStatement(ctx, p)
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

	if !p.expect(tokens.PunctRightCurly) {
		return nil // Error in parsing the block expression
	}

	expr.Statements = statements
	expr.Yield = yieldExpr

	return expr
}
