// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"context"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
)

func parseCommentStatement(ctx context.Context, parser *Parser) ast.Statement {
	head := parser.head()
	comment := parser.advance()
	rnge := tokens.Range{
		File: head.Range.File,
		From: head.Range.From,
		To:   comment.Range.To,
	}
	if comment.IsOfKind(tokens.TrailingComment) {
		return ast.NewTrailingCommentStatement(comment.Value, rnge)
	}
	return ast.NewCommentStatement(comment.Value, rnge)
}
