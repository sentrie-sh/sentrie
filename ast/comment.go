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
	Pos     tokens.Position
	Content string
}

type TrailingCommentExpression struct {
	Pos            tokens.Position
	CommentContent string
	Wrap           Expression
}

type PrecedingCommentExpression struct {
	Pos            tokens.Position
	CommentContent string
	Wrap           Expression
}

func (c CommentStatement) Position() tokens.Position {
	return c.Pos
}

func (c CommentStatement) String() string {
	return fmt.Sprintf("LineComment(%s)", c.Content)
}

func (c CommentStatement) statementNode() {}

func (t TrailingCommentExpression) String() string {
	return t.Wrap.String() + " -- " + t.CommentContent
}

func (t TrailingCommentExpression) Position() tokens.Position {
	return t.Pos
}

func (t TrailingCommentExpression) expressionNode() {}

func (p PrecedingCommentExpression) String() string {
	return p.CommentContent + " -- " + p.Wrap.String()
}

func (p PrecedingCommentExpression) Position() tokens.Position {
	return p.Pos
}

func (p PrecedingCommentExpression) expressionNode() {}

var (
	_ Statement  = &CommentStatement{}
	_ Expression = &TrailingCommentExpression{}
	_ Expression = &PrecedingCommentExpression{}
)
