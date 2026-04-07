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
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
	"github.com/sentrie-sh/sentrie/trinary"
	"github.com/sentrie-sh/sentrie/xerr"
	"github.com/stretchr/testify/require"
)

func resolveRn(line int) tokens.Range {
	return tokens.Range{
		File: "resolve.sentrie",
		From: tokens.Pos{Line: line, Column: 0, Offset: 0},
		To:   tokens.Pos{Line: line, Column: 1, Offset: 1},
	}
}

func TestResolveNamespace_NotFound(t *testing.T) {
	idx := CreateIndex()
	ns, err := idx.ResolveNamespace("missing/ns")
	require.Nil(t, ns)
	require.Error(t, err)
	require.True(t, errors.Is(err, xerr.NotFoundError{}), "got %v", err)
	require.Contains(t, err.Error(), "namespace: missing/ns")
}

func TestResolveNamespace_OK(t *testing.T) {
	ctx := context.Background()
	idx := CreateIndex()
	rn := resolveRn(1)
	nsStmt := ast.NewNamespaceStatement(ast.NewFQN([]string{"com", "example"}, rn), rn)
	_, err := idx.ensureNamespace(ctx, nsStmt)
	require.NoError(t, err)

	got, err := idx.ResolveNamespace("com/example")
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, "com/example", got.FQN.String())
}

func TestResolvePolicy_NotFoundNamespace(t *testing.T) {
	idx := CreateIndex()
	p, err := idx.ResolvePolicy("com/none", "p")
	require.Nil(t, p)
	require.Error(t, err)
	require.True(t, errors.Is(err, xerr.NotFoundError{}))
}

func TestResolvePolicy_NotFoundPolicy(t *testing.T) {
	ctx := context.Background()
	idx := CreateIndex()
	rn := resolveRn(1)
	nsStmt := ast.NewNamespaceStatement(ast.NewFQN([]string{"com", "example"}, rn), rn)
	_, err := idx.ensureNamespace(ctx, nsStmt)
	require.NoError(t, err)

	p, err := idx.ResolvePolicy("com/example", "noSuchPolicy")
	require.Nil(t, p)
	require.Error(t, err)
	require.True(t, errors.Is(err, xerr.NotFoundError{}))
	require.Contains(t, err.Error(), filepath.Join("com/example", "noSuchPolicy"))
}

func TestResolvePolicy_OK(t *testing.T) {
	ctx := context.Background()
	idx := CreateIndex()
	rn0 := resolveRn(1)
	program := &ast.Program{
		Reference: "resolve.sentrie",
		Statements: []ast.Statement{
			ast.NewNamespaceStatement(ast.NewFQN([]string{"com", "example"}, rn0), rn0),
			ast.NewPolicyStatement(
				"auth",
				[]ast.Statement{
					ast.NewFactStatement("user", ast.NewStringTypeRef(resolveRn(3)), "u", nil, true, resolveRn(3)),
					ast.NewRuleStatement("allow", nil, ast.NewTrinaryLiteral(trinary.True, resolveRn(4)), nil, resolveRn(4)),
					ast.NewRuleExportStatement("allow", nil, resolveRn(5)),
				},
				resolveRn(2),
			),
		},
	}
	require.NoError(t, idx.AddProgram(ctx, program))

	p, err := idx.ResolvePolicy("com/example", "auth")
	require.NoError(t, err)
	require.NotNil(t, p)
	require.Equal(t, "auth", p.Name)
	require.Equal(t, "com/example/auth", p.FQN.String())
}

func TestResolveShape_NotFoundNamespace(t *testing.T) {
	idx := CreateIndex()
	s, err := idx.ResolveShape("com/none", "S")
	require.Nil(t, s)
	require.Error(t, err)
	require.True(t, errors.Is(err, xerr.NotFoundError{}))
}

func TestResolveShape_NotFoundShape(t *testing.T) {
	ctx := context.Background()
	idx := CreateIndex()
	rn := resolveRn(1)
	nsStmt := ast.NewNamespaceStatement(ast.NewFQN([]string{"com", "example"}, rn), rn)
	_, err := idx.ensureNamespace(ctx, nsStmt)
	require.NoError(t, err)

	s, err := idx.ResolveShape("com/example", "MissingShape")
	require.Nil(t, s)
	require.Error(t, err)
	require.True(t, errors.Is(err, xerr.NotFoundError{}))
	require.Contains(t, err.Error(), filepath.Join("com/example", "MissingShape"))
}

func TestResolveShape_OK(t *testing.T) {
	ctx := context.Background()
	idx := CreateIndex()
	rn0 := resolveRn(1)
	program := &ast.Program{
		Reference: "resolve.sentrie",
		Statements: []ast.Statement{
			ast.NewNamespaceStatement(ast.NewFQN([]string{"com", "example"}, rn0), rn0),
			ast.NewShapeStatement("User", ast.NewStringTypeRef(resolveRn(2)), nil, resolveRn(2)),
			ast.NewPolicyStatement(
				"p",
				[]ast.Statement{
					ast.NewFactStatement("user", ast.NewStringTypeRef(resolveRn(3)), "u", nil, true, resolveRn(3)),
					ast.NewRuleStatement("allow", nil, ast.NewTrinaryLiteral(trinary.True, resolveRn(4)), nil, resolveRn(4)),
					ast.NewRuleExportStatement("allow", nil, resolveRn(5)),
				},
				resolveRn(2),
			),
		},
	}
	require.NoError(t, idx.AddProgram(ctx, program))

	s, err := idx.ResolveShape("com/example", "User")
	require.NoError(t, err)
	require.NotNil(t, s)
	require.Equal(t, "User", s.Name)
	require.Equal(t, "com/example/User", s.FQN.String())
}

func TestVerifyRuleExported_OK(t *testing.T) {
	ctx := context.Background()
	idx := CreateIndex()
	rn0 := resolveRn(1)
	program := &ast.Program{
		Reference: "resolve.sentrie",
		Statements: []ast.Statement{
			ast.NewNamespaceStatement(ast.NewFQN([]string{"com", "example"}, rn0), rn0),
			ast.NewPolicyStatement(
				"auth",
				[]ast.Statement{
					ast.NewFactStatement("user", ast.NewStringTypeRef(resolveRn(3)), "u", nil, true, resolveRn(3)),
					ast.NewRuleStatement("allow", nil, ast.NewTrinaryLiteral(trinary.True, resolveRn(4)), nil, resolveRn(4)),
					ast.NewRuleExportStatement("allow", nil, resolveRn(5)),
				},
				resolveRn(2),
			),
		},
	}
	require.NoError(t, idx.AddProgram(ctx, program))
	p, err := idx.ResolvePolicy("com/example", "auth")
	require.NoError(t, err)

	require.NoError(t, p.VerifyRuleExported("allow"))
}

func TestVerifyRuleExported_NotExported(t *testing.T) {
	ctx := context.Background()
	idx := CreateIndex()
	rn0 := resolveRn(1)
	program := &ast.Program{
		Reference: "resolve.sentrie",
		Statements: []ast.Statement{
			ast.NewNamespaceStatement(ast.NewFQN([]string{"com", "example"}, rn0), rn0),
			ast.NewPolicyStatement(
				"auth",
				[]ast.Statement{
					ast.NewFactStatement("user", ast.NewStringTypeRef(resolveRn(3)), "u", nil, true, resolveRn(3)),
					ast.NewRuleStatement("allow", nil, ast.NewTrinaryLiteral(trinary.True, resolveRn(4)), nil, resolveRn(4)),
					ast.NewRuleStatement("deny", nil, ast.NewTrinaryLiteral(trinary.False, resolveRn(5)), nil, resolveRn(5)),
					ast.NewRuleExportStatement("allow", nil, resolveRn(6)),
				},
				resolveRn(2),
			),
		},
	}
	require.NoError(t, idx.AddProgram(ctx, program))
	p, err := idx.ResolvePolicy("com/example", "auth")
	require.NoError(t, err)

	err = p.VerifyRuleExported("deny")
	require.Error(t, err)
	require.True(t, errors.Is(err, xerr.NotExportedError{}))
	require.Contains(t, err.Error(), RuleFQN("com/example", "auth", "deny"))
}

func TestVerifyShapeExported_OK(t *testing.T) {
	ctx := context.Background()
	idx := CreateIndex()
	rn0 := resolveRn(1)
	program := &ast.Program{
		Reference: "resolve.sentrie",
		Statements: []ast.Statement{
			ast.NewNamespaceStatement(ast.NewFQN([]string{"com", "example"}, rn0), rn0),
			ast.NewShapeStatement("User", ast.NewStringTypeRef(resolveRn(2)), nil, resolveRn(2)),
			ast.NewShapeExportStatement("User", resolveRn(3)),
			ast.NewPolicyStatement(
				"p",
				[]ast.Statement{
					ast.NewFactStatement("user", ast.NewStringTypeRef(resolveRn(4)), "u", nil, true, resolveRn(4)),
					ast.NewRuleStatement("allow", nil, ast.NewTrinaryLiteral(trinary.True, resolveRn(5)), nil, resolveRn(5)),
					ast.NewRuleExportStatement("allow", nil, resolveRn(6)),
				},
				resolveRn(2),
			),
		},
	}
	require.NoError(t, idx.AddProgram(ctx, program))
	ns, err := idx.ResolveNamespace("com/example")
	require.NoError(t, err)

	require.NoError(t, ns.VerifyShapeExported("User"))
}

func TestVerifyShapeExported_NotExported(t *testing.T) {
	ctx := context.Background()
	idx := CreateIndex()
	rn0 := resolveRn(1)
	program := &ast.Program{
		Reference: "resolve.sentrie",
		Statements: []ast.Statement{
			ast.NewNamespaceStatement(ast.NewFQN([]string{"com", "example"}, rn0), rn0),
			ast.NewShapeStatement("User", ast.NewStringTypeRef(resolveRn(2)), nil, resolveRn(2)),
			ast.NewPolicyStatement(
				"p",
				[]ast.Statement{
					ast.NewFactStatement("user", ast.NewStringTypeRef(resolveRn(3)), "u", nil, true, resolveRn(3)),
					ast.NewRuleStatement("allow", nil, ast.NewTrinaryLiteral(trinary.True, resolveRn(4)), nil, resolveRn(4)),
					ast.NewRuleExportStatement("allow", nil, resolveRn(5)),
				},
				resolveRn(2),
			),
		},
	}
	require.NoError(t, idx.AddProgram(ctx, program))
	ns, err := idx.ResolveNamespace("com/example")
	require.NoError(t, err)

	err = ns.VerifyShapeExported("User")
	require.Error(t, err)
	require.True(t, errors.Is(err, xerr.NotExportedError{}))
	require.Contains(t, err.Error(), ShapeFQN("com/example", "User"))
}

func TestRuleFQN(t *testing.T) {
	require.Equal(t, "com/example/auth/allow", RuleFQN("com/example", "auth", "allow"))
}

func TestShapeFQN(t *testing.T) {
	require.Equal(t, "com/example/User", ShapeFQN("com/example", "User"))
}
