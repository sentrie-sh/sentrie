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
	"time"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
)

func parseCallExpression(ctx context.Context, p *Parser, left ast.Expression, precedence Precedence) ast.Expression {
	lparen := p.advance()

	exp := &ast.CallExpression{
		Callee:    left,
		Pos:       lparen.Position,
		Arguments: parseExpressionList(ctx, p, tokens.PunctRightParentheses),
	}

	if p.head().IsOfKind(tokens.TokenBang) {
		_ = p.advance()
		exp.Memoized = true
		exp.MemoizeTTL = nil

		if p.head().IsOfKind(tokens.Int) {
			literal := parseIntegerLiteral(ctx, p)
			if literal == nil {
				return nil
			}
			ttl := time.Duration(literal.(*ast.IntegerLiteral).Value) * time.Second
			exp.MemoizeTTL = &ttl
		}
	}

	return exp
}

func parseExpressionList(ctx context.Context, parser *Parser, end tokens.Kind) []ast.Expression {
	exps := []ast.Expression{}

	// TODO :: check if we can do this with an expect
	if parser.head().IsOfKind(end) {
		_ = parser.advance() // consume the end token
		// just return an empty list
		return exps
	}

	for parser.hasTokens() {
		exp := parser.parseExpression(ctx, LOWEST)
		if exp == nil {
			return nil
		}
		exps = append(exps, exp)
		if parser.head().IsOfKind(tokens.PunctComma) {
			_ = parser.advance() // consume the comma
		}
		if parser.head().IsOfKind(end) {
			_ = parser.advance() // consume the end token
			break                // exit the loop if we reach the end token
		}
	}

	return exps
}
