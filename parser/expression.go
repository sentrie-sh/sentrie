// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"context"
	"log/slog"
	"slices"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
)

// Core Pratt parsing method
func (p *Parser) parseExpression(ctx context.Context, precedence Precedence) ast.Expression {
	slog.DebugContext(ctx, "Starting expression parsing", "current", p.current, "precedence", precedence)
	defer slog.DebugContext(ctx, "Finished expression parsing", "current", p.current, "precedence", precedence)

	var comments []*tokens.Instance

	// if the next token is a comment
	for p.canExpectAnyOf(tokens.TrailingComment, tokens.LineComment) {
		c := p.advance()
		comments = append(comments, &c)
	}

	prefix, exists := p.prefixHandlers[p.current.Kind]
	if !exists {
		p.noPrefixParseFnError(p.current)
		return nil
	}

	leftExp := wrapWithTrailingComment(prefix(ctx, p), p)

	// An operand that failed has recorded its error. Handing nil on as the left
	// of the next operator panics (`a/not/f(x)` reached the call parser with a
	// nil callee), and wrapping it in a preceding comment hides the failure, so
	// a nil operand ends the expression on both paths.
	for leftExp != nil && precedences[p.current.Kind] > precedence {
		infixFn, exists := p.infixHandlers[p.current.Kind]
		if !exists {
			break
		}

		leftExp = wrapWithTrailingComment(infixFn(ctx, p, leftExp, precedences[p.current.Kind]), p)
	}

	if leftExp == nil {
		return nil
	}

	// if we had found a comment before hand
	if len(comments) > 0 {
		slices.Reverse(comments)
	}
	for len(comments) > 0 {
		leftExp = ast.NewPrecedingCommentExpression(comments[0].Value, leftExp, comments[0].Range)
		comments = comments[1:]
	}
	return leftExp
}

func wrapWithTrailingComment(expr ast.Expression, parser *Parser) ast.Expression {
	if expr == nil {
		return nil
	}
	if parser.head().IsOfKind(tokens.TrailingComment) {
		comment := parser.advance()
		return ast.NewTrailingCommentExpression(comment.Value, expr, comment.Range)
	}
	return expr
}
