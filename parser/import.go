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
[*] importClause ::= 'import' 'decision' IDENT 'from' FQN ( withClause )* ;
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
		Range: tokens.Range{
			File: head.Range.File,
			From: head.Range.From,
		},
	}

	what, found := p.advanceExpected(tokens.Ident)
	if !found {
		return nil // Error in parsing the import expression
	}
	importExp.RuleToImport = what.Value

	if !p.expect(tokens.KeywordFrom) {
		return nil // Error in parsing the import expression
	}

	fqn, fqnRange := parseFQN(ctx, p)
	if fqn == nil {
		return nil // Error in parsing the import expression
	}
	importExp.FromPolicyFQN = fqn
	importExp.Range.To = fqnRange.To

	// Check for 'with' clauses
	for p.head().IsOfKind(tokens.KeywordWith) {
		withClause := parseWithClause(ctx, p)
		if withClause != nil {
			importExp.Withs = append(importExp.Withs, withClause)

			// update the range to the last with clause
			importExp.Range.To = withClause.Range.To
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
		Range: tokens.Range{
			File: head.Range.File,
			From: tokens.Pos{
				Line:   head.Range.From.Line,
				Column: head.Range.From.Column,
				Offset: head.Range.From.Offset,
			},
			To: tokens.Pos{
				Line:   head.Range.From.Line,
				Column: head.Range.From.Column,
				Offset: head.Range.From.Offset,
			},
		},
		Name: name.Value,
		Expr: val,
	}
}
