// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2026 Binaek Sarkar
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

package trace

import (
	"context"
	"errors"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/tokens"
)

func (s *TraceTestSuite) TestNewAndDoneSetsDuration() {
	ident := ast.NewIdentifier("x", tokens.Range{File: "test.sentra"})
	_, node, done := New(context.Background(), ident, "ident", map[string]any{"key": "value"})
	s.Equal(ident.Kind(), node.Kind)
	s.Equal("ident", node.Op)
	s.Equal("value", node.Meta["key"])
	s.Zero(node.Duration)
	done()
	s.NotZero(node.Duration)
}

func (s *TraceTestSuite) TestIgnoredAndUnsupportedNodeKinds() {
	ident := ast.NewIdentifier("x", tokens.Range{File: "test.sentra"})
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
