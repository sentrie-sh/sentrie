// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package trace

import (
	"errors"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/tokens"
)

func (s *TraceTestSuite) TestNewAndDoneSetsDuration() {
	ident := ast.NewIdentifier("x", tokens.Range{File: "test.sentrie"})
	_, node, done := New(s.T().Context(), ident, "ident", map[string]any{"key": "value"})
	s.Equal(ident.Kind(), node.Kind)
	s.Equal("ident", node.Op)
	s.Equal("value", node.Meta["key"])
	s.Zero(node.Duration)
	done()
	s.NotZero(node.Duration)
}

func (s *TraceTestSuite) TestIgnoredAndUnsupportedNodeKinds() {
	ident := ast.NewIdentifier("x", tokens.Range{File: "test.sentrie"})
	ignored := IgnoredStmt(ident)
	s.Equal("stmt-ignored", ignored.Kind)
	s.Contains(ignored.Meta["type"], "ast.Identifier")
	unsupported := UnsupportedExpression(ident)
	s.Equal("unsupported", unsupported.Kind)
	s.Contains(unsupported.Meta["type"], "ast.Identifier")
}

func (s *TraceTestSuite) TestAttachSetResultSetErr() {
	parent := &Node{Kind: "root"}
	left := &Node{Kind: "left"}
	right := &Node{Kind: "right"}
	s.Same(parent, parent.Attach())
	s.Same(parent, parent.Attach(left, right))
	s.Len(parent.Children, 2)
	s.Same(parent, parent.SetResult(box.String("ok")))
	s.Equal(box.String("ok"), parent.Result)
	s.Same(parent, parent.SetErr(nil))
	s.Empty(parent.Err)
	s.Same(parent, parent.SetErr(errors.New("boom")))
	s.Equal("boom", parent.Err)
}
