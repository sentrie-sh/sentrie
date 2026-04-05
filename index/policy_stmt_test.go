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

package index

import (
	"testing"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
	"github.com/sentrie-sh/sentrie/trinary"
	"github.com/stretchr/testify/require"
)

func TestPolicyStmtKindClassification(t *testing.T) {
	r := tokens.Range{File: "t.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 1}}

	tests := []struct {
		stmt ast.Statement
		want policyStmtKind
	}{
		{ast.NewCommentStatement("-- x", r), policyStmtComment},
		{ast.NewTitleStatement("x", r), policyStmtMetadata},
		{ast.NewDescriptionStatement("x", r), policyStmtMetadata},
		{ast.NewVersionStatement("1.0.0", r), policyStmtMetadata},
		{ast.NewTagStatement("k", "v", r), policyStmtMetadata},
		{ast.NewFactStatement("f", ast.NewStringTypeRef(r), "f", nil, true, r), policyStmtFact},
		{ast.NewUseStatement([]string{"x"}, "", []string{"m"}, "m", r), policyStmtUse},
		{ast.NewVarDeclaration("n", nil, ast.NewTrinaryLiteral(trinary.True, r), r), policyStmtBody},
		{ast.NewRuleStatement("rule", nil, ast.NewTrinaryLiteral(trinary.True, r), nil, r), policyStmtBody},
		{ast.NewRuleExportStatement("rule", nil, r), policyStmtBody},
		{ast.NewShapeStatement("S", nil, nil, r), policyStmtBody},
	}
	for _, tc := range tests {
		got := policyStmtKindOf(tc.stmt)
		require.Equal(t, tc.want, got, "%T", tc.stmt)
	}

	ns := ast.NewNamespaceStatement(ast.NewFQN([]string{"x"}, r), r)
	require.Equal(t, policyStmtUnknown, policyStmtKindOf(ns))

	require.True(t, isMetadataStmt(ast.NewTitleStatement("t", r)))
	require.True(t, isFactStmt(ast.NewFactStatement("f", ast.NewStringTypeRef(r), "f", nil, true, r)))
	require.True(t, isUseStmt(ast.NewUseStatement([]string{"x"}, "", []string{"m"}, "m", r)))
	require.True(t, isBodyStmt(ast.NewRuleStatement("rule", nil, ast.NewTrinaryLiteral(trinary.True, r), nil, r)))
}

func TestBuildTagsByKey(t *testing.T) {
	require.Nil(t, buildTagsByKey(nil))
	require.Nil(t, buildTagsByKey([]PolicyTagPair{}))
	m := buildTagsByKey([]PolicyTagPair{
		{Key: "a", Value: "1"},
		{Key: "a", Value: "2"},
		{Key: "b", Value: ""},
	})
	require.Equal(t, []string{"1", "2"}, m["a"])
	require.Equal(t, []string{""}, m["b"])
}
