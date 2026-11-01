// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"context"
	"math"
	"strconv"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
)

func parseUnaryExpression(ctx context.Context, p *Parser) ast.Expression {
	operatorToken := p.advance()
	minInt64 := operatorToken.Value == "-" && isTwoTo63Literal(p.head())

	right := p.parseExpression(ctx, UNARY)
	if right == nil {
		return nil
	}

	rnge := operatorToken.Range
	rnge.To = right.Span().To

	// 2^63 does not fit int64, so on its own it parses as a float literal
	// (#166). Directly negated it is exactly math.MinInt64; fold that one case
	// so MinInt64 can be written as a literal. If anything bound to the
	// literal, right is not a bare FloatLiteral and nothing is folded.
	if _, ok := right.(*ast.FloatLiteral); ok && minInt64 {
		return ast.NewIntegerLiteral(math.MinInt64, rnge)
	}

	return ast.NewUnaryExpression(operatorToken.Value, right, rnge)
}

// isTwoTo63Literal reports whether t is an integer token whose value is 2^63.
func isTwoTo63Literal(t tokens.Instance) bool {
	if !t.IsOfKind(tokens.Int) {
		return false
	}
	u, err := strconv.ParseUint(t.Value, 10, 64)
	return err == nil && u == 1<<63
}
