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

func parseThePolicyStatement(ctx context.Context, p *Parser) ast.Statement {
	policy := &ast.PolicyStatement{
		Pos: p.head().Position,
	}
	if !p.expect(tokens.KeywordPolicy) {
		return nil
	}

	name, ok := p.advanceExpected(tokens.Ident)
	if !ok {
		return nil
	}

	policy.Name = name.Value

	if !p.expect(tokens.PunctLeftCurly) {
		return nil
	}

	for p.hasTokens() && !p.head().IsOfKind(tokens.PunctRightCurly) {
		stmt := parsePolicyStatement(ctx, p)
		if p.err != nil {
			return nil
		}
		if stmt == nil {
			continue
		}
		policy.Statements = append(policy.Statements, stmt)
	}

	if !p.expect(tokens.PunctRightCurly) {
		return nil
	}

	return policy
}

func parsePolicyStatement(ctx context.Context, p *Parser) ast.Statement {
	if handler, ok := p.policyStatementHandlers[p.current.Kind]; ok {
		return handler(ctx, p)
	}
	p.errorf("unexpected token '%s' at %s", p.current.Kind, p.current.Position)
	return nil
}

func (p *Parser) registerPolicyStatementHandler(tokenType tokens.Kind, fn statementParser) {
	p.policyStatementHandlers[tokenType] = fn
}
