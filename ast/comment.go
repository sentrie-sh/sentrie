// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package ast

import (
	"fmt"

	"github.com/sentrie-sh/sentrie/tokens"
)

// CommentStatement is a `--` comment in statement position: at the top
// level, in a policy body, in a block body or in a shape body. Trailing is
// set when the comment followed other code on its line, so a printer can put
// it back at the end of that line.
type CommentStatement struct {
	*baseNode
	Content  string
	Trailing bool
}

func NewCommentStatement(content string, ssp tokens.Range) *CommentStatement {
	return &CommentStatement{
		baseNode: &baseNode{
			Rnge:  ssp,
			Kind_: "comment",
		},
		Content: content,
	}
}

// NewTrailingCommentStatement builds a CommentStatement for a comment that
// followed other code on its line.
func NewTrailingCommentStatement(content string, ssp tokens.Range) *CommentStatement {
	c := NewCommentStatement(content, ssp)
	c.Trailing = true
	return c
}

type TrailingCommentExpression struct {
	*baseNode
	CommentContent string
	Wrap           Expression
}

func NewTrailingCommentExpression(commentContent string, wrap Expression, ssp tokens.Range) *TrailingCommentExpression {
	return &TrailingCommentExpression{
		baseNode: &baseNode{
			Rnge:  ssp,
			Kind_: "trailing_comment",
		},
		CommentContent: commentContent,
		Wrap:           wrap,
	}
}

type PrecedingCommentExpression struct {
	*baseNode
	CommentContent string
	Wrap           Expression
}

func NewPrecedingCommentExpression(commentContent string, wrap Expression, ssp tokens.Range) *PrecedingCommentExpression {
	return &PrecedingCommentExpression{
		baseNode: &baseNode{
			Rnge:  ssp,
			Kind_: "preceding_comment",
		},
		CommentContent: commentContent,
		Wrap:           wrap,
	}
}

func (c CommentStatement) String() string {
	return fmt.Sprintf("LineComment(%s)", c.Content)
}

func (c CommentStatement) statementNode() {}

func (t TrailingCommentExpression) String() string {
	return t.Wrap.String() + " -- " + t.CommentContent
}

func (t TrailingCommentExpression) expressionNode() {}

func (p PrecedingCommentExpression) String() string {
	return p.CommentContent + " -- " + p.Wrap.String()
}

func (p PrecedingCommentExpression) expressionNode() {}

var (
	_ Statement  = &CommentStatement{}
	_ Node       = &CommentStatement{}
	_ Expression = &TrailingCommentExpression{}
	_ Node       = &TrailingCommentExpression{}
	_ Expression = &PrecedingCommentExpression{}
	_ Node       = &PrecedingCommentExpression{}
)
