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
	if !p.expect(tokens.PunctLeftParentheses) {
		return nil
	}

	rnge := left.Span()

	arguments := parseExpressionList(ctx, p, tokens.PunctRightParentheses)
	if arguments == nil {
		return nil
	}

	// Find the closing parenthesis position
	rparen, found := p.advanceExpected(tokens.PunctRightParentheses)
	if !found {
		return nil
	}

	rnge.To = rparen.Range.To

	exp := ast.NewCallExpression(left, arguments, false, nil, rnge)

	if p.head().IsOfKind(tokens.TokenBang) {
		bang, found := p.advanceExpected(tokens.TokenBang)
		if !found {
			return nil
		}
		rnge.To = bang.Range.To

		exp.Memoized = true
		exp.MemoizeTTL = nil

		if p.head().IsOfKind(tokens.Int) {
			literal := parseIntegerLiteral(ctx, p)
			if literal == nil {
				return nil
			}
			ttl := time.Duration(literal.(*ast.IntegerLiteral).Value) * time.Second
			exp.MemoizeTTL = &ttl
			rnge.To = literal.Span().To
		}
	}

	return exp
}

func parseExpressionList(ctx context.Context, parser *Parser, end tokens.Kind) []ast.Expression {
	exps := []ast.Expression{}

	for parser.hasTokens() && !parser.canExpect(end) {
		exp := parser.parseExpression(ctx, LOWEST)
		if exp == nil {
			return nil
		}
		exps = append(exps, exp)
		if parser.canExpect(tokens.PunctComma) {
			parser.advance() // consume the comma
			continue
		}

	}

	return exps
}
