// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"context"
	"testing"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/builtins"
	"github.com/sentrie-sh/sentrie/tokens"
	"github.com/sentrie-sh/sentrie/trinary"
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

func TestResolveShapeReadOnlyNilAndEmpty(t *testing.T) {
	t.Parallel()
	idx := CreateIndex()
	rng := testBuiltinKindRange()
	require.Nil(t, resolveShapeReadOnly(idx, nil, nil))
	empty := ast.NewFQN([]string{}, rng)
	require.True(t, empty.IsEmpty())
	require.Nil(t, resolveShapeReadOnly(idx, nil, empty.Ptr()))
}

func TestResolveKindNilExpression(t *testing.T) {
	t.Parallel()
	k := &kindCheckCtx{}
	_, ok := k.resolveKind(nil)
	require.False(t, ok)
}

func TestLookupShapeFieldTypeReadOnlyMissingWithParent(t *testing.T) {
	t.Parallel()
	idx := CreateIndex()
	rng := testBuiltinKindRange()
	ns := testNamespace("com/ex")
	child, err := createShape(ns, nil, ast.NewShapeStatement("Child", nil, &ast.Cmplx{
		Range:  rng,
		With:   ast.NewFQN([]string{"Missing"}, rng).Ptr(),
		Fields: map[string]*ast.ShapeField{},
	}, rng))
	require.NoError(t, err)
	_, ok := lookupShapeFieldTypeReadOnly(idx, nil, child, "items")
	require.False(t, ok)
}

func TestResolveShapeReadOnlyExportedShapeInNamespace(t *testing.T) {
	t.Parallel()
	idx := CreateIndex()
	rng := testBuiltinKindRange()
	nsFQN := ast.NewFQN([]string{"com", "ex", "remote"}, rng)
	ns := createNamespace(ast.NewNamespaceStatement(nsFQN, rng))
	idx.Namespaces[ns.FQN.String()] = ns

	shape, err := createShape(ns, nil, ast.NewShapeStatement("Exported", ast.NewStringTypeRef(rng), nil, rng))
	require.NoError(t, err)
	ns.Shapes["Exported"] = shape
	require.NoError(t, ns.addShapeExport(&ExportedShape{Name: "Exported", Statement: ast.NewShapeExportStatement("Exported", rng)}))

	got := resolveShapeReadOnly(idx, nil, ast.NewFQN([]string{"com", "ex", "remote", "Exported"}, rng).Ptr())
	require.Same(t, shape, got)
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

func TestCheckBuiltinCallsCancelledDuringLoops(t *testing.T) {
	t.Parallel()
	rng := testBuiltinKindRange()
	idx := CreateIndex()
	ns := testNamespace("com/ex")
	p := testPolicy(ns, "p")
	p.Rules["allow"] = testRule(p, "allow", ast.NewBlockExpression(nil, ast.NewIntegerLiteral(1, rng), rng))
	ns.Policies[p.Name] = p
	idx.Namespaces[ns.FQN.String()] = ns

	derive := &Derive{
		Name:      "pred",
		FQN:       ast.CreateFQN(ns.FQN, "pred"),
		Namespace: ns,
		Lambda: ast.NewLambdaExpression(
			nil,
			ast.NewBlockExpression(nil, ast.NewIntegerLiteral(1, rng), rng),
			rng,
		),
	}
	idx.DerivesByFQN[derive.FQN.String()] = derive

	cases := []struct {
		name string
		at   int
	}{
		{name: "namespace loop", at: 2},
		{name: "policy loop", at: 3},
		{name: "derive loop", at: 4},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := &scriptedCtx{Context: t.Context(), at: map[int]error{tc.at: context.Canceled}}
			err := idx.checkBuiltinCalls(ctx)
			require.Error(t, err)
			require.Contains(t, err.Error(), "validation cancelled")
		})
	}
}

func TestLookupDeriveDefineShortAndNamespace(t *testing.T) {
	t.Parallel()
	ns := testNamespace("com/ex")
	helper := &Derive{Name: "helper", FQN: ast.CreateFQN(ns.FQN, "helper"), Namespace: ns}
	nsDerive := &Derive{Name: "nsPred", FQN: ast.CreateFQN(ns.FQN, "nsPred"), Namespace: ns}
	ns.Derives["nsPred"] = nsDerive

	defineCtx := &kindCheckCtx{
		derive: &Derive{DefineShort: map[string]*Derive{"helper": helper}},
	}
	require.Same(t, helper, defineCtx.lookupDerive("helper"))

	policyCtx := &kindCheckCtx{
		policy: &Policy{
			Namespace: ns,
			Derives:   map[string]*Derive{},
		},
	}
	require.Same(t, nsDerive, policyCtx.lookupDerive("nsPred"))
}

func TestIsBuiltinCallShadowedByDerive(t *testing.T) {
	t.Parallel()
	rng := testBuiltinKindRange()
	ns := testNamespace("com/ex")
	derive := &Derive{
		Name: "count",
		Lambda: ast.NewLambdaExpression(
			[]string{"x"},
			ast.NewBlockExpression(nil, ast.NewIntegerLiteral(1, rng), rng),
			rng,
		),
	}
	k := &kindCheckCtx{
		policy: &Policy{Namespace: ns, Derives: map[string]*Derive{"count": derive}},
		scope:  map[string]bindingInfo{},
	}
	call := ast.NewCallExpression(
		ast.NewIdentifier("count", rng),
		[]ast.Expression{ast.NewListLiteral(nil, rng)},
		false,
		nil,
		rng,
	)
	_, ok := k.isBuiltinCall(call)
	require.False(t, ok)
}

func TestLookupShapeKindNilAndEmptyComplex(t *testing.T) {
	t.Parallel()
	idx := CreateIndex()
	rng := testBuiltinKindRange()
	ns := testNamespace("com/ex")
	idx.Namespaces[ns.FQN.String()] = ns

	_, ok := lookupShapeKind(idx, nil, nil)
	require.False(t, ok)

	empty, err := createShape(ns, nil, ast.NewShapeStatement("Empty", nil, &ast.Cmplx{
		Range:  rng,
		Fields: map[string]*ast.ShapeField{},
	}, rng))
	require.NoError(t, err)
	ns.Shapes["Empty"] = empty
	_, ok = lookupShapeKind(idx, &Policy{Namespace: ns}, ast.NewShapeTypeRef(ast.NewFQN([]string{"com", "ex", "Empty"}, rng).Ptr(), rng))
	require.False(t, ok)
}

func TestResolveShapeReadOnlyFailurePaths(t *testing.T) {
	t.Parallel()
	idx := CreateIndex()
	rng := testBuiltinKindRange()
	nsFQN := ast.NewFQN([]string{"com", "ex"}, rng)
	ns := createNamespace(ast.NewNamespaceStatement(nsFQN, rng))
	idx.Namespaces[ns.FQN.String()] = ns

	require.Nil(t, resolveShapeReadOnly(idx, nil, ast.NewFQN([]string{"com", "missing", "Shape"}, rng).Ptr()))
	require.Nil(t, resolveShapeReadOnly(idx, nil, ast.NewFQN([]string{"com", "ex", "Missing"}, rng).Ptr()))

	require.NoError(t, ns.addShapeExport(&ExportedShape{Name: "Ghost", Statement: ast.NewShapeExportStatement("Ghost", rng)}))
	require.Nil(t, resolveShapeReadOnly(idx, nil, ast.NewFQN([]string{"com", "ex", "Ghost"}, rng).Ptr()))
}

func TestLookupShapeFieldTypeReadOnlyNilShape(t *testing.T) {
	t.Parallel()
	_, ok := lookupShapeFieldTypeReadOnly(CreateIndex(), nil, nil, "x")
	require.False(t, ok)
}

func TestFieldTypeRefHopMissingShape(t *testing.T) {
	t.Parallel()
	idx := CreateIndex()
	rng := testBuiltinKindRange()
	_, ok := fieldTypeRefHop(idx, nil, ast.NewShapeTypeRef(ast.NewFQN([]string{"com", "ex", "Missing"}, rng).Ptr(), rng), "x")
	require.False(t, ok)
	_, ok = fieldTypeRefHop(idx, nil, nil, "x")
	require.False(t, ok)
}

func TestResolveKindLiteralVariants(t *testing.T) {
	t.Parallel()
	rng := testBuiltinKindRange()
	k := &kindCheckCtx{}

	kind, ok := k.resolveKind(ast.NewMapLiteral(nil, rng))
	require.True(t, ok)
	require.Equal(t, box.ValueDict, kind)

	kind, ok = k.resolveKind(ast.NewTrinaryLiteral(trinary.True, rng))
	require.True(t, ok)
	require.Equal(t, box.ValueTrinary, kind)

	_, ok = k.resolveKind(ast.NewNullLiteral(rng))
	require.False(t, ok)
}

func TestResolveIdentKindScopeBindings(t *testing.T) {
	t.Parallel()
	rng := testBuiltinKindRange()
	lam := ast.NewLambdaExpression([]string{"x"}, ast.NewBlockExpression(nil, ast.NewIntegerLiteral(1, rng), rng), rng)
	derive := &Derive{Name: "pred", Lambda: lam}

	k := &kindCheckCtx{scope: map[string]bindingInfo{
		"d": {derive: derive},
		"f": {lambda: lam},
		"s": {typeRef: ast.NewStringTypeRef(rng)},
	}}

	kind, ok := k.resolveIdentKind(ast.NewIdentifier("d", rng))
	require.True(t, ok)
	require.Equal(t, box.ValueCallable, kind)

	kind, ok = k.resolveIdentKind(ast.NewIdentifier("f", rng))
	require.True(t, ok)
	require.Equal(t, box.ValueCallable, kind)

	kind, ok = k.resolveIdentKind(ast.NewIdentifier("s", rng))
	require.True(t, ok)
	require.Equal(t, box.ValueString, kind)
}

func TestResolveIdentKindFromPolicyDerive(t *testing.T) {
	t.Parallel()
	rng := testBuiltinKindRange()
	derive := &Derive{Name: "pred", Lambda: ast.NewLambdaExpression(nil, ast.NewBlockExpression(nil, ast.NewIntegerLiteral(1, rng), rng), rng)}
	k := &kindCheckCtx{policy: &Policy{Derives: map[string]*Derive{"pred": derive}}}
	kind, ok := k.resolveIdentKind(ast.NewIdentifier("pred", rng))
	require.True(t, ok)
	require.Equal(t, box.ValueCallable, kind)
}

func TestResolveFieldAccessChainFailures(t *testing.T) {
	t.Parallel()
	rng := testBuiltinKindRange()
	idx := CreateIndex()
	ns := testNamespace("com/ex")
	rowShape, err := createShape(ns, nil, ast.NewShapeStatement("Row", nil, &ast.Cmplx{
		Range: rng,
		Fields: map[string]*ast.ShapeField{
			"items": {Name: "items", Type: ast.NewListTypeRef(ast.NewNumberTypeRef(rng), rng), Range: rng},
		},
	}, rng))
	require.NoError(t, err)

	k := &kindCheckCtx{
		idx:    idx,
		policy: &Policy{Namespace: ns},
		scope: map[string]bindingInfo{
			"row": {typeRef: ast.NewShapeTypeRef(rowShape.FQN.Ptr(), rng)},
			"raw": {letInit: ast.NewListLiteral(nil, rng)},
		},
	}

	_, ok := k.resolveFieldAccessChain(ast.NewFieldAccessExpression(ast.NewIntegerLiteral(1, rng), "x", rng))
	require.False(t, ok)

	_, ok = k.resolveFieldAccessChain(ast.NewFieldAccessExpression(ast.NewIdentifier("missing", rng), "x", rng))
	require.False(t, ok)

	_, ok = k.resolveFieldAccessChain(ast.NewFieldAccessExpression(ast.NewIdentifier("raw", rng), "x", rng))
	require.False(t, ok)

	_, ok = k.resolveFieldAccessChain(ast.NewFieldAccessExpression(ast.NewIdentifier("row", rng), "missing", rng))
	require.False(t, ok)
}

func TestResolveCallableArityLambdaFromLet(t *testing.T) {
	t.Parallel()
	rng := testBuiltinKindRange()
	lam := ast.NewLambdaExpressionFull([]string{"a", "b"}, nil, []bool{true, false}, nil, ast.NewBlockExpression(nil, ast.NewIntegerLiteral(1, rng), rng), rng)
	k := &kindCheckCtx{scope: map[string]bindingInfo{"f": {lambda: lam}}}

	n, ok := k.resolveCallableArity(nil)
	require.False(t, ok)
	require.Equal(t, 0, n)

	n, ok = k.resolveCallableArity(ast.NewIdentifier("f", rng))
	require.True(t, ok)
	require.Equal(t, 1, n)
}

func TestBuiltinParamSigAtVariadic(t *testing.T) {
	t.Parallel()
	rng := testBuiltinKindRange()
	k := &kindCheckCtx{}
	call := ast.NewCallExpression(
		ast.NewIdentifier("error", rng),
		[]ast.Expression{ast.NewStringLiteral("a", rng), ast.NewStringLiteral("b", rng)},
		false,
		nil,
		rng,
	)
	errs := k.checkBuiltinCall(call)
	require.Empty(t, errs)
	ps, ok := builtinParamSigAt(builtins.Table["error"].Sig, 1)
	require.True(t, ok)
	require.Equal(t, "args", ps.Name)
}
