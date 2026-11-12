// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"sync/atomic"
	"testing"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
	"github.com/sentrie-sh/sentrie/trinary"
	"github.com/stretchr/testify/require"
)

func rng(line int) tokens.Range {
	return tokens.Range{
		File: "res.sentrie",
		From: tokens.Pos{Line: line, Column: 0, Offset: 0},
		To:   tokens.Pos{Line: line, Column: 1, Offset: 1},
	}
}

func TestCreateShapeSkipsBlankFieldNameEntries(t *testing.T) {
	nsStmt := ast.NewNamespaceStatement(
		ast.NewFQN([]string{"com", "example"}, rng(1)),
		rng(1),
	)
	ns := createNamespace(nsStmt)
	cm := &ast.Cmplx{
		Range: rng(2),
		With:  nil,
		Fields: map[string]*ast.ShapeField{
			"":   {Range: rng(3), Name: "", Type: ast.NewStringTypeRef(rng(3))},
			"id": {Range: rng(4), Name: "id", Optional: false, Type: ast.NewStringTypeRef(rng(4))},
		},
	}
	stmt := ast.NewShapeStatement("S", nil, cm, rng(2))
	shape, err := createShape(ns, nil, stmt)
	require.NoError(t, err)
	require.Len(t, shape.Model.Fields, 1)
	require.Contains(t, shape.Model.Fields, "id")
}

func TestShapeResolveDependency_NoModelEarlyReturn(t *testing.T) {
	nsStmt := ast.NewNamespaceStatement(
		ast.NewFQN([]string{"com", "example"}, rng(1)),
		rng(1),
	)
	ns := createNamespace(nsStmt)
	stmt := ast.NewShapeStatement("Alias", ast.NewStringTypeRef(rng(2)), nil, rng(2))
	shape, err := createShape(ns, nil, stmt)
	require.NoError(t, err)
	idx := CreateIndex()
	require.NoError(t, shape.resolveDependency(idx))
}

func TestShapeResolveDependency_AlreadyHydratedShortCircuits(t *testing.T) {
	nsStmt := ast.NewNamespaceStatement(
		ast.NewFQN([]string{"com", "example"}, rng(1)),
		rng(1),
	)
	ns := createNamespace(nsStmt)
	stmt := ast.NewShapeStatement("Alias", ast.NewStringTypeRef(rng(2)), nil, rng(2))
	shape, err := createShape(ns, nil, stmt)
	require.NoError(t, err)
	atomic.StoreUint32(&shape.hydrated, 1)
	idx := CreateIndex()
	require.NoError(t, shape.resolveDependency(idx))
}

func TestShapeResolveDependency_ComposeFromAliasBaseErrors(t *testing.T) {
	ctx := t.Context()
	idx := CreateIndex()
	nsStmt := ast.NewNamespaceStatement(
		ast.NewFQN([]string{"com", "example"}, rng(1)),
		rng(1),
	)
	ns, err := idx.ensureNamespace(ctx, nsStmt)
	require.NoError(t, err)

	baseStmt := ast.NewShapeStatement(
		"AliasBase",
		ast.NewStringTypeRef(rng(2)),
		nil,
		rng(2),
	)
	base, err := createShape(ns, nil, baseStmt)
	require.NoError(t, err)
	require.NoError(t, ns.addShape(base))

	with := ast.NewFQN([]string{"AliasBase"}, rng(5)).Ptr()
	childStmt := ast.NewShapeStatement(
		"Child",
		nil,
		&ast.Cmplx{
			Range: rng(5),
			With:  with,
			Fields: map[string]*ast.ShapeField{
				"x": {Range: rng(6), Name: "x", Optional: false, Type: ast.NewStringTypeRef(rng(6))},
			},
		},
		rng(5),
	)
	polStmt := ast.NewPolicyStatement(
		"P",
		[]ast.Statement{
			ast.NewFactStatement("user", ast.NewStringTypeRef(rng(10)), "u", nil, true, rng(10)),
			childStmt,
			ast.NewRuleStatement("r", nil, nil, ast.NewTrinaryLiteral(trinary.True, rng(11)), rng(11)),
			ast.NewRuleExportStatement("r", nil, rng(12)),
		},
		rng(9),
	)
	prog := &ast.Program{
		Reference: "p.sentrie",
		Statements: []ast.Statement{
			nsStmt,
			polStmt,
		},
	}
	policy, err := createPolicy(ns, polStmt, prog, nil, nil)
	require.NoError(t, err)
	child := policy.Shapes["Child"]
	require.NotNil(t, child)

	err = child.resolveDependency(idx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot compose")
}

func TestShapeResolveDependency_DuplicateFieldFromComposedBaseErrors(t *testing.T) {
	ctx := t.Context()
	idx := CreateIndex()
	nsStmt := ast.NewNamespaceStatement(
		ast.NewFQN([]string{"com", "example"}, rng(1)),
		rng(1),
	)
	ns, err := idx.ensureNamespace(ctx, nsStmt)
	require.NoError(t, err)

	baseStmt := ast.NewShapeStatement(
		"BaseEnt",
		nil,
		&ast.Cmplx{
			Range: rng(2),
			With:  nil,
			Fields: map[string]*ast.ShapeField{
				"id": {Range: rng(3), Name: "id", Optional: false, Type: ast.NewStringTypeRef(rng(3))},
			},
		},
		rng(2),
	)
	base, err := createShape(ns, nil, baseStmt)
	require.NoError(t, err)
	require.NoError(t, ns.addShape(base))

	with := ast.NewFQN([]string{"BaseEnt"}, rng(10)).Ptr()
	childStmt := ast.NewShapeStatement(
		"ChildDup",
		nil,
		&ast.Cmplx{
			Range: rng(10),
			With:  with,
			Fields: map[string]*ast.ShapeField{
				"id": {Range: rng(11), Name: "id", Optional: false, Type: ast.NewNumberTypeRef(rng(11))},
			},
		},
		rng(10),
	)
	polStmt := ast.NewPolicyStatement(
		"P2",
		[]ast.Statement{
			ast.NewFactStatement("user", ast.NewStringTypeRef(rng(20)), "u", nil, true, rng(20)),
			childStmt,
			ast.NewRuleStatement("r", nil, nil, ast.NewTrinaryLiteral(trinary.True, rng(21)), rng(21)),
			ast.NewRuleExportStatement("r", nil, rng(22)),
		},
		rng(19),
	)
	prog := &ast.Program{
		Reference: "p2.sentrie",
		Statements: []ast.Statement{
			nsStmt,
			polStmt,
		},
	}
	policy, err := createPolicy(ns, polStmt, prog, nil, nil)
	require.NoError(t, err)
	child := policy.Shapes["ChildDup"]
	require.NotNil(t, child)

	err = child.resolveDependency(idx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot compose with duplicate shape field")
}

func TestShapeResolveDependency_PolicyLocalComposeUsesInPolicyShapes(t *testing.T) {
	ctx := t.Context()
	idx := CreateIndex()
	nsStmt := ast.NewNamespaceStatement(
		ast.NewFQN([]string{"com", "example"}, rng(1)),
		rng(1),
	)
	ns, err := idx.ensureNamespace(ctx, nsStmt)
	require.NoError(t, err)

	baseStmt := ast.NewShapeStatement(
		"PBase",
		nil,
		&ast.Cmplx{
			Range: rng(2),
			With:  nil,
			Fields: map[string]*ast.ShapeField{
				"n": {Range: rng(3), Name: "n", Optional: false, Type: ast.NewStringTypeRef(rng(3))},
			},
		},
		rng(2),
	)
	with := ast.NewFQN([]string{"PBase"}, rng(10)).Ptr()
	extStmt := ast.NewShapeStatement(
		"PExt",
		nil,
		&ast.Cmplx{
			Range:  rng(10),
			With:   with,
			Fields: map[string]*ast.ShapeField{},
		},
		rng(10),
	)
	polStmt := ast.NewPolicyStatement(
		"P3",
		[]ast.Statement{
			ast.NewFactStatement("user", ast.NewStringTypeRef(rng(20)), "u", nil, true, rng(20)),
			baseStmt,
			extStmt,
			ast.NewRuleStatement("r", nil, nil, ast.NewTrinaryLiteral(trinary.True, rng(21)), rng(21)),
			ast.NewRuleExportStatement("r", nil, rng(22)),
		},
		rng(19),
	)
	prog := &ast.Program{
		Reference: "p3.sentrie",
		Statements: []ast.Statement{
			nsStmt,
			polStmt,
		},
	}
	policy, err := createPolicy(ns, polStmt, prog, nil, nil)
	require.NoError(t, err)
	ext := policy.Shapes["PExt"]
	require.NotNil(t, ext)

	err = ext.resolveDependency(idx)
	require.NoError(t, err)
	require.Contains(t, ext.Model.Fields, "n")
}
