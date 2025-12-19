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

// 'is [not] defined' | 'is [not] empty' | '<expression> is <expression>'
func parseIsExpression(ctx context.Context, p *Parser, left ast.Expression, precedence Precedence) ast.Expression {
	start := p.head()

	rnge := start.Range

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
		expr = ast.NewIsDefinedExpression(left, rnge)
		p.advance()
	} else if p.canExpect(tokens.KeywordEmpty) {
		expr = ast.NewIsEmptyExpression(left, rnge)
		p.advance()
	} else {
		right := p.parseExpression(ctx, precedence)
		rnge.To = right.Span().To
		expr = ast.NewInfixExpression(left, right, start.Value, rnge)
	}

	// if we have a 'not' then wrap with a not unary
	if not != nil {
		expr = ast.NewUnaryExpression(not.Value, expr, rnge)
	}

	return expr
}
