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

func parseIsExpression(ctx context.Context, p *Parser, left ast.Expression, precedence Precedence) ast.Expression {
	start := p.head()

	// consume the 'is' token
	if !p.expect(tokens.KeywordIs) {
		return nil
	}

	var not *tokens.Instance
	if p.head().IsOfKind(tokens.KeywordNot) {
		n := p.advance() // consume the 'not' token
		not = &n
	}

	var expr ast.Expression

	if p.canExpect(tokens.KeywordDefined) {
		// 'is [not] defined' case
		expr = &ast.IsDefinedExpression{
			Pos:  start.Position,
			Left: left,
		}
		p.advance()
	} else if p.canExpect(tokens.KeywordEmpty) {
		expr = &ast.IsEmptyExpression{
			Pos:  start.Position,
			Left: left,
		}
		p.advance()
	} else {
		expr = &ast.InfixExpression{
			Pos:      start.Position,
			Left:     left,
			Operator: start.Value,
			Right:    p.parseExpression(ctx, precedence),
		}
	}

	// if we have a 'not' then wrap with a not unary
	if not != nil {
		expr = &ast.UnaryExpression{
			Pos:      start.Position,
			Operator: not.Value,
			Right:    expr,
		}
	}

	return expr
}
