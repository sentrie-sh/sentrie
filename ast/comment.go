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
	Range   tokens.Range
	Content string
}

type TrailingCommentExpression struct {
	Range          tokens.Range
	CommentContent string
	Wrap           Expression
}

type PrecedingCommentExpression struct {
	Range          tokens.Range
	CommentContent string
	Wrap           Expression
}

func (c CommentStatement) Span() tokens.Range {
	return c.Range
}

func (c CommentStatement) String() string {
	return fmt.Sprintf("LineComment(%s)", c.Content)
}

func (c CommentStatement) statementNode() {}

func (t TrailingCommentExpression) String() string {
	return t.Wrap.String() + " -- " + t.CommentContent
}

func (t TrailingCommentExpression) Span() tokens.Range {
	return t.Range
}

func (t TrailingCommentExpression) expressionNode() {}

func (p PrecedingCommentExpression) String() string {
	return p.CommentContent + " -- " + p.Wrap.String()
}

func (p PrecedingCommentExpression) Span() tokens.Range {
	return p.Range
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
