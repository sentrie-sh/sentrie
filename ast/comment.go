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

package ast

import (
	"fmt"

	"github.com/sentrie-sh/sentrie/tokens"
)

type CommentStatement struct {
	*baseNode
	Content string
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
