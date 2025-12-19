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

func parseThePolicyStatement(ctx context.Context, p *Parser) ast.Statement {
	start := p.head()

	rnge := start.Range

	if !p.expect(tokens.KeywordPolicy) {
		return nil
	}

	nameIdent, ok := p.advanceExpected(tokens.Ident)
	if !ok {
		return nil
	}

	name := nameIdent.Value

	if !p.expect(tokens.PunctLeftCurly) {
		return nil
	}

	var statements []ast.Statement

	for p.hasTokens() && !p.head().IsOfKind(tokens.PunctRightCurly) {
		stmt := parsePolicyStatement(ctx, p)
		if p.err != nil {
			return nil
		}
		if stmt == nil {
			return nil // Error in parsing the policy statement
		}
		statements = append(statements, stmt)

		// consume the optional semicolon
		if p.canExpect(tokens.PunctSemicolon) {
			p.advance()
		}

		// consume trailing comments
		for p.canExpectAnyOf(tokens.TrailingComment, tokens.LineComment) {
			p.advance()
		}
	}

	rCurly, found := p.advanceExpected(tokens.PunctRightCurly)
	if !found {
		return nil
	}
	rnge.To = rCurly.Range.To

	return ast.NewPolicyStatement(name, statements, rnge)
}

func parsePolicyStatement(ctx context.Context, p *Parser) ast.Statement {
	if handler, ok := p.policyStatementHandlers[p.head().Kind]; ok {
		return handler(ctx, p)
	}
	p.errorf("unexpected token '%s'", p.head().Kind)
	return nil
}

func (p *Parser) registerPolicyStatementHandler(tokenType tokens.Kind, fn statementParser) {
	p.policyStatementHandlers[tokenType] = fn
}
