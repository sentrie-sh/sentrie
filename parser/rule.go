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

	if !parser.canExpectAnyOf(tokens.PunctLeftCurly, tokens.KeywordImport) {
		parser.errorf("expected '{' or 'import', got %s at %s", parser.peek().Kind, parser.peek().Position)
		return nil
	}

	if parser.head().IsOfKind(tokens.PunctLeftCurly) {
		// If we have a block, parse it
		expression := parseBlockExpression(ctx, parser)
		if expression == nil {
			return nil // Error in parsing the rule body
		}
		stmt.Body = expression
	} else if parser.head().IsOfKind(tokens.KeywordImport) {
		// If we have an import clause, parse it
		importClause := parseImportExpression(ctx, parser)
		if importClause == nil {
			return nil // Error in parsing the import clause
		}
		stmt.Body = importClause
	}

	return stmt
}
