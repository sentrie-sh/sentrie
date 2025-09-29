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
[*] importClause ::= 'import' 'decision' IDENT 'from' IDENT ( '/' IDENT )* ( withClause )* ;
[*] withClause   ::= 'with' IDENT 'as' IDENT ;
[*] blockExpr    ::= '{' expr '}' ;
*/
func parseImportExpression(ctx context.Context, p *Parser) ast.Expression {
	head := p.head()
	if !p.expect(tokens.KeywordImport) {
		return nil
	}

	if !p.expect(tokens.KeywordDecision) {
		return nil // Error in parsing the import expression
	}

	importExp := &ast.ImportClause{
		Pos: head.Position,
	}

	what, found := p.advanceExpected(tokens.Ident)
	if !found {
		return nil // Error in parsing the import expression
	}
	importExp.RuleToImport = what.Value

	if !p.expect(tokens.KeywordFrom) {
		return nil // Error in parsing the import expression
	}

	from, found := p.advanceExpected(tokens.Ident)
	if !found {
		return nil // Error in parsing the import expression
	}
	importExp.FromPolicyFQN = []string{from.Value}
	// Check for additional segments in the 'from' path
	for p.head().IsOfKind(tokens.TokenDiv) {
		p.advance() // consume the slash
		segment, found := p.advanceExpected(tokens.Ident)
		if !found {
			return nil // Error in parsing the import expression
		}
		importExp.FromPolicyFQN = append(importExp.FromPolicyFQN, segment.Value)
	}

	// Check for 'with' clauses
	for p.head().IsOfKind(tokens.KeywordWith) {
		withClause := parseWithClause(ctx, p)
		if withClause != nil {
			importExp.Withs = append(importExp.Withs, withClause)
		}
	}

	return importExp
}

func parseWithClause(ctx context.Context, p *Parser) *ast.WithClause {
	head := p.head()
	if !p.expect(tokens.KeywordWith) {
		return nil // Error in parsing the with clause
	}

	name, found := p.advanceExpected(tokens.Ident)
	if !found {
		return nil // Error in parsing the with clause
	}

	if !p.expect(tokens.KeywordAs) {
		return nil // Error in parsing the with clause
	}

	val := p.parseExpression(ctx, LOWEST)
	if val == nil {
		return nil // Error in parsing the with clause
	}

	return &ast.WithClause{
		Pos:  head.Position,
		Name: name.Value,
		Expr: val,
	}
}
