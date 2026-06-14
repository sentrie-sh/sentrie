// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"context"
	"testing"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/tokens"
	"github.com/stretchr/testify/require"
)

func testBuiltinKindRange() tokens.Range {
	return tokens.Range{File: "bk.sentrie", From: tokens.Pos{Line: 1, Column: 0}, To: tokens.Pos{Line: 1, Column: 1}}
}

func TestTypeRefKindScalarsAndContainers(t *testing.T) {
	t.Parallel()
	idx := CreateIndex()
	rng := testBuiltinKindRange()
	cases := []struct {
		ref  ast.TypeRef
		kind box.ValueKind
	}{
		{ast.NewStringTypeRef(rng), box.ValueString},
		{ast.NewNumberTypeRef(rng), box.ValueNumber},
		{ast.NewTrinaryTypeRef(rng), box.ValueTrinary},
		{ast.NewListTypeRef(ast.NewNumberTypeRef(rng), rng), box.ValueList},
		{ast.NewDictTypeRef(ast.NewStringTypeRef(rng), rng), box.ValueDict},
		{ast.NewDocumentTypeRef(rng), box.ValueDocument},
		{ast.NewRecordTypeRef([]ast.TypeRef{ast.NewNumberTypeRef(rng)}, rng), box.ValueList},
	}
	for _, tc := range cases {
		kind, ok := typeRefKind(idx, nil, tc.ref)
		require.True(t, ok)
		require.Equal(t, tc.kind, kind)
	}
	_, ok := typeRefKind(idx, nil, ast.NewNullableTypeRef(ast.NewStringTypeRef(rng), rng))
	require.False(t, ok)
	_, ok = typeRefKind(idx, nil, nil)
	require.False(t, ok)
}

func TestResolveShapeReadOnlyExportedNamespace(t *testing.T) {
	t.Parallel()
	idx := CreateIndex()
	rng := testBuiltinKindRange()
	nsFQN := ast.NewFQN([]string{"com", "ex"}, rng)
	ns := createNamespace(ast.NewNamespaceStatement(nsFQN, rng))
	idx.Namespaces[ns.FQN.String()] = ns

	shape, err := createShape(ns, nil, ast.NewShapeStatement("Remote", ast.NewStringTypeRef(rng), nil, rng))
	require.NoError(t, err)
	ns.Shapes["Remote"] = shape
	require.NoError(t, ns.addShapeExport(&ExportedShape{Name: "Remote", Statement: ast.NewShapeExportStatement("Remote", rng)}))

	got := resolveShapeReadOnly(idx, nil, ast.NewFQN([]string{"com", "ex", "Remote"}, rng).Ptr())
	require.Same(t, shape, got)
}

func TestLookupShapeFieldTypeReadOnlyWithChain(t *testing.T) {
	t.Parallel()
	idx := CreateIndex()
	rng := testBuiltinKindRange()
	nsFQN := ast.NewFQN([]string{"com", "ex"}, rng)
	ns := createNamespace(ast.NewNamespaceStatement(nsFQN, rng))
	idx.Namespaces[ns.FQN.String()] = ns

	base, err := createShape(ns, nil, ast.NewShapeStatement("Base", nil, &ast.Cmplx{
		Range: rng,
		Fields: map[string]*ast.ShapeField{
			"items": {Name: "items", Type: ast.NewListTypeRef(ast.NewNumberTypeRef(rng), rng), Range: rng},
		},
	}, rng))
	require.NoError(t, err)
	ns.Shapes["Base"] = base

	child, err := createShape(ns, nil, ast.NewShapeStatement("Child", nil, &ast.Cmplx{
		Range:  rng,
		With:   ast.NewFQN([]string{"Base"}, rng).Ptr(),
		Fields: map[string]*ast.ShapeField{},
	}, rng))
	require.NoError(t, err)

	policy := &Policy{Namespace: ns}
	ref, ok := lookupShapeFieldTypeReadOnly(idx, policy, child, "items")
	require.True(t, ok)
	_, ok = ref.(*ast.ListTypeRef)
	require.True(t, ok)
}

func TestResolveCallableArityDeriveIdentifier(t *testing.T) {
	t.Parallel()
	idx := CreateIndex()
	rng := testBuiltinKindRange()
	nsFQN := ast.NewFQN([]string{"com", "ex"}, rng)
	ns := createNamespace(ast.NewNamespaceStatement(nsFQN, rng))
	idx.Namespaces[ns.FQN.String()] = ns

	lam := ast.NewLambdaExpressionFull([]string{"a", "b"}, nil, nil, nil, ast.NewBlockExpression(nil, ast.NewIntegerLiteral(1, rng), rng), rng)
	derive := &Derive{Name: "pred", Lambda: lam, Namespace: ns}
	k := &kindCheckCtx{idx: idx, policy: &Policy{Namespace: ns, Derives: map[string]*Derive{"pred": derive}}}
	n, ok := k.resolveCallableArity(ast.NewIdentifier("pred", rng))
	require.True(t, ok)
	require.Equal(t, 2, n)
}

func TestResolveIdentKindDirectWrapper(t *testing.T) {
	t.Parallel()
	rng := testBuiltinKindRange()
	k := &kindCheckCtx{scope: map[string]bindingInfo{
		"x": {kindKnown: true, kind: box.ValueNumber},
	}}
	kind, ok := k.resolveIdentKind(ast.NewIdentifier("x", rng))
	require.True(t, ok)
	require.Equal(t, box.ValueNumber, kind)
}

func TestCheckBuiltinCallsContextCancellation(t *testing.T) {
	t.Parallel()
	idx := CreateIndex()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := idx.checkBuiltinCalls(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "validation cancelled")
}

func TestCheckBuiltinCallExpectedCallableFallback(t *testing.T) {
	t.Parallel()
	rng := testBuiltinKindRange()
	k := &kindCheckCtx{}
	call := ast.NewCallExpression(
		ast.NewIdentifier("filter", rng),
		[]ast.Expression{ast.NewListLiteral([]ast.Expression{ast.NewIntegerLiteral(1, rng)}, rng), ast.NewIntegerLiteral(1, rng)},
		false,
		nil,
		rng,
	)
	errs := k.checkBuiltinCall(call)
	require.Len(t, errs, 1)
	require.Contains(t, errs[0].Error(), "expected callable, got number")
}

func TestFieldTypeRefHopDocumentReturnsUnknown(t *testing.T) {
	t.Parallel()
	idx := CreateIndex()
	rng := testBuiltinKindRange()
	_, ok := fieldTypeRefHop(idx, nil, ast.NewDocumentTypeRef(rng), "x")
	require.False(t, ok)
}
