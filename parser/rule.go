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

// 'when' expr { expr } | import
func parseRuleStatement(ctx context.Context, parser *Parser) ast.Statement {
	stmt := &ast.RuleStatement{
		Pos: parser.head().Position,
	}

	parser.advance() // consume 'rule'

	name, found := parser.advanceExpected(tokens.Ident)
	if !found {
		return nil
	}
	stmt.RuleName = name.Value

	// Assignment operator is required for rule declarations
	if !parser.expect(tokens.TokenAssign) {
		return nil // Error in parsing the rule statement
	}

	if parser.canExpect(tokens.KeywordDefault) {
		parser.advance() // consume 'default'
		defaultExpr := parser.parseExpression(ctx, LOWEST)
		if defaultExpr == nil {
			return nil // Error in parsing the default expression
		}
		stmt.Default = defaultExpr
	}

	if parser.canExpect(tokens.KeywordWhen) {
		parser.advance() // consume 'when'
		whenExpr := parser.parseExpression(ctx, LOWEST)
		if whenExpr == nil {
			return nil // Error in parsing the when expression
		}
		stmt.When = whenExpr
	}

	// Parse rule body - can be import clause or expression (including block expressions)
	if parser.canExpect(tokens.KeywordImport) {
		// If we have an import clause, parse it
		importClause := parseImportExpression(ctx, parser)
		if importClause == nil {
			return nil // Error in parsing the import clause
		}
		stmt.Body = importClause
	} else {
		// Parse as expression (handles direct expressions, block expressions, etc.)
		expression := parser.parseExpression(ctx, LOWEST)
		if expression == nil {
			return nil // Error in parsing the rule body
		}
		stmt.Body = expression
	}

	return stmt
}
