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

package trace

import (
	"context"
	"errors"
	"testing"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/tokens"
	"github.com/stretchr/testify/require"
)

func TestNewAndDoneSetsDuration(t *testing.T) {
	ident := ast.NewIdentifier("x", tokens.Range{File: "test.sentra"})
	_, node, done := New(context.Background(), ident, "ident", map[string]any{"key": "value"})

	require.Equal(t, ident.Kind(), node.Kind)
	require.Equal(t, "ident", node.Op)
	require.Equal(t, "value", node.Meta["key"])
	require.Zero(t, node.Duration)

	done()
	require.NotZero(t, node.Duration)
}

func TestIgnoredAndUnsupportedNodeKinds(t *testing.T) {
	ident := ast.NewIdentifier("x", tokens.Range{File: "test.sentra"})

	ignored := IgnoredStmt(ident)
	require.Equal(t, "stmt-ignored", ignored.Kind)
	require.Contains(t, ignored.Meta["type"], "ast.Identifier")

	unsupported := UnsupportedExpression(ident)
	require.Equal(t, "unsupported", unsupported.Kind)
	require.Contains(t, unsupported.Meta["type"], "ast.Identifier")
}

func TestAttachSetResultSetErr(t *testing.T) {
	parent := &Node{Kind: "root"}
	left := &Node{Kind: "left"}
	right := &Node{Kind: "right"}

	require.Same(t, parent, parent.Attach())
	require.Same(t, parent, parent.Attach(left, right))
	require.Len(t, parent.Children, 2)

	require.Same(t, parent, parent.SetResult(box.String("ok")))
	require.Equal(t, box.String("ok"), parent.Result)

	require.Same(t, parent, parent.SetErr(nil))
	require.Empty(t, parent.Err)

	require.Same(t, parent, parent.SetErr(errors.New("boom")))
	require.Equal(t, "boom", parent.Err)
}
