// SPDX-License-Identifier: Apache-2.0
//
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

	for precedences[p.current.Kind] > precedence {
		infixFn, exists := p.infixHandlers[p.current.Kind]
		if !exists {
			break
		}

		leftExp = wrapWithTrailingComment(infixFn(ctx, p, leftExp, precedences[p.current.Kind]), p)
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
