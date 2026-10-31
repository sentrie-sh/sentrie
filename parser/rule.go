// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"context"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
)

// rule <ident> [':' typeRef] = 'default' expr 'when' expr ( { expr } | importClause )
func parseRuleStatement(ctx context.Context, parser *Parser) ast.Statement {
	ruleToken, found := parser.advanceExpected(tokens.KeywordRule)
	if !found {
		return nil
	}
	rnge := ruleToken.Range

	nameToken, found := parser.advanceExpected(tokens.Ident)
	if !found {
		return nil
	}
	name := nameToken.Value
	rnge.To = nameToken.Range.To

	// Optional declared type, as on `let`
	var typeRef ast.TypeRef
	if parser.canExpect(tokens.PunctColon) {
		parser.advance() // consume ':'
		typeRef = parseTypeRef(ctx, parser)
		if typeRef == nil {
			return nil
		}
	}

	// Assignment operator is required for rule declarations
	if !parser.expect(tokens.TokenAssign) {
		return nil // Error in parsing the rule statement
	}

	var defaultExpr ast.Expression
	var whenExpr ast.Expression
	var bodyExpr ast.Expression

	if parser.canExpect(tokens.KeywordDefault) {
		parser.advance() // consume 'default'
		defaultExpr = parser.parseExpression(ctx, LOWEST)
		if defaultExpr == nil {
			return nil // Error in parsing the default expression
		}
	}

	if parser.canExpect(tokens.KeywordWhen) {
		parser.advance() // consume 'when'
		whenExpr = parser.parseExpression(ctx, LOWEST)
		if whenExpr == nil {
			return nil // Error in parsing the when expression
		}
	}

	// Parse rule body - can be import clause or expression (including block expressions)
	if parser.canExpect(tokens.KeywordImport) {
		// If we have an import clause, parse it
		importClause := parseImportExpression(ctx, parser)
		if importClause == nil {
			return nil // Error in parsing the import clause
		}
		bodyExpr = importClause
		rnge.To = importClause.Span().To
	} else {
		// Parse as expression (handles direct expressions, block expressions, etc.)
		expression := parser.parseExpression(ctx, LOWEST)
		if expression == nil {
			return nil // Error in parsing the rule body
		}
		bodyExpr = expression
		rnge.To = expression.Span().To
	}

	stmt := ast.NewRuleStatement(name, defaultExpr, whenExpr, bodyExpr, rnge)
	stmt.Type = typeRef
	return stmt
}
