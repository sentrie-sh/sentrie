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

func (p *Parser) registerParseFns() {
	// Initialize prefix parse functions
	p.prefixHandlers = make(map[tokens.Kind]prefixParser)
	p.registerPrefix(tokens.KeywordTrue, parseTrinaryLiteral)
	p.registerPrefix(tokens.KeywordFalse, parseTrinaryLiteral)
	p.registerPrefix(tokens.KeywordUnknown, parseTrinaryLiteral)

	p.registerPrefix(tokens.KeywordNull, parseNullLiteral)
	p.registerPrefix(tokens.KeywordAny, quantifierParserFactory(tokens.KeywordAny))
	p.registerPrefix(tokens.KeywordAll, quantifierParserFactory(tokens.KeywordAll))
	p.registerPrefix(tokens.KeywordFilter, quantifierParserFactory(tokens.KeywordFilter))
	p.registerPrefix(tokens.KeywordMap, quantifierParserFactory(tokens.KeywordMap))
	p.registerPrefix(tokens.KeywordDistinct, quantifierParserFactory(tokens.KeywordDistinct))
	p.registerPrefix(tokens.KeywordCount, parseCountExpression)

	p.registerPrefix(tokens.KeywordReduce, parseReduceExpression)
	p.registerPrefix(tokens.KeywordCast, parseCastExpression)

	p.registerPrefix(tokens.Ident, parseIdentifier)
	p.registerPrefix(tokens.String, parseStringLiteral)
	p.registerPrefix(tokens.Int, parseIntegerLiteral)
	p.registerPrefix(tokens.Float, parseFloatLiteral)

	p.registerPrefix(tokens.TokenBang, parseUnaryExpression)
	p.registerPrefix(tokens.TokenMinus, parseUnaryExpression)
	p.registerPrefix(tokens.TokenPlus, parseUnaryExpression)
	p.registerPrefix(tokens.KeywordTransform, parseTransformExpression)

	p.registerPrefix(tokens.PunctLeftParentheses, parseGroupedExpression)
	p.registerPrefix(tokens.PunctLeftBracket, parseListLiteral)

	// special case - left curly brace - switches based on peek token
	p.registerPrefix(tokens.PunctLeftCurly, parseFromLeftCurly)

	// Initialize infix parse functions
	p.infixHandlers = make(map[tokens.Kind]infixParser)
	p.registerInfix(tokens.KeywordAnd, parseInfixExpression)
	p.registerInfix(tokens.KeywordOr, parseInfixExpression)
	p.registerInfix(tokens.KeywordXor, parseInfixExpression)
	p.registerInfix(tokens.KeywordIn, parseInfixExpression)
	p.registerInfix(tokens.KeywordMatches, parseInfixExpression)
	p.registerInfix(tokens.KeywordContains, parseInfixExpression)
	p.registerInfix(tokens.KeywordIs, parseIsExpression)

	p.registerInfix(tokens.PunctLeftBracket, parseIndexAccessExpression)
	p.registerInfix(tokens.TokenDot, parseFieldAccessExpression)
	p.registerInfix(tokens.PunctLeftParentheses, parseCallExpression)

	p.registerInfix(tokens.TokenPlus, parseInfixExpression)
	p.registerInfix(tokens.TokenMinus, parseInfixExpression)
	p.registerInfix(tokens.TokenMul, parseInfixExpression)
	p.registerInfix(tokens.TokenDiv, parseInfixExpression)
	p.registerInfix(tokens.TokenMod, parseInfixExpression)
	p.registerInfix(tokens.TokenEq, parseInfixExpression)
	p.registerInfix(tokens.TokenNeq, parseInfixExpression)
	p.registerInfix(tokens.TokenLt, parseInfixExpression)
	p.registerInfix(tokens.TokenGt, parseInfixExpression)
	p.registerInfix(tokens.TokenLte, parseInfixExpression)
	p.registerInfix(tokens.TokenGte, parseInfixExpression)
	p.registerInfix(tokens.TokenQuestion, parseTernaryExpression)

	// not is a special case - since it may be a unary or a binary depending on it's placement
	// let x = not true
	// let x = "string" not in ["string", "other"]
	p.registerPrefix(tokens.KeywordNot, parseUnaryExpression)
	p.registerInfix(tokens.KeywordNot, parseNotExpression)

	// statementHandlers
	p.statementHandlers = make(map[tokens.Kind]statementParser)
	p.registerStatementHandler(tokens.KeywordNamespace, parseNamespaceStatement)
	p.registerStatementHandler(tokens.LineComment, parseCommentStatement)
	p.registerStatementHandler(tokens.TrailingComment, parseCommentStatement)
	p.registerStatementHandler(tokens.KeywordPolicy, parseThePolicyStatement)
	p.registerStatementHandler(tokens.KeywordShape, parseShapeStatement)
	p.registerStatementHandler(tokens.KeywordExport, parseShapeExportStatement)

	// policyStatementHandlers
	p.policyStatementHandlers = make(map[tokens.Kind]statementParser)
	p.registerPolicyStatementHandler(tokens.LineComment, parseCommentStatement)
	p.registerPolicyStatementHandler(tokens.TrailingComment, parseCommentStatement)
	p.registerPolicyStatementHandler(tokens.KeywordRule, parseRuleStatement)
	p.registerPolicyStatementHandler(tokens.KeywordFact, parseFactStatement)
	p.registerPolicyStatementHandler(tokens.KeywordExport, parseRuleExportStatement)
	p.registerPolicyStatementHandler(tokens.KeywordLet, parseLetsStatement)
	p.registerPolicyStatementHandler(tokens.KeywordUse, parseUseStatement)
	p.registerPolicyStatementHandler(tokens.KeywordShape, parseShapeStatement)
}

type prefixParser func(ctx context.Context, parser *Parser) ast.Expression
type infixParser func(ctx context.Context, parser *Parser, left ast.Expression, precedence Precedence) ast.Expression

type statementParser func(ctx context.Context, parser *Parser) ast.Statement

var PRIMITIVE_TYPES = []tokens.Kind{
	tokens.KeywordString,
	tokens.KeywordInt,
	tokens.KeywordFloat,
	tokens.KeywordBoolean,
	tokens.KeywordDocument,
}

var AGGREGATE_TYPES = []tokens.Kind{
	tokens.KeywordList,
	tokens.KeywordMap,
	tokens.KeywordRecord,
}
