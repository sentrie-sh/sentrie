// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2026 Binaek Sarkar

package index

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
)

func testRange() tokens.Range {
	return tokens.NewRange("test.sent", tokens.Pos{Line: 0, Column: 1}, tokens.Pos{Line: 0, Column: 1})
}

func testFQN(parts ...string) ast.FQN {
	return ast.NewFQN(parts, testRange())
}

func testNamespace(name string) *Namespace {
	stmt := ast.NewNamespaceStatement(testFQN(name), testRange())
	return createNamespace(stmt)
}

func testPolicy(ns *Namespace, name string) *Policy {
	stmt := ast.NewPolicyStatement(name, nil, testRange())
	return &Policy{
		Statement:   stmt,
		Namespace:   ns,
		Name:        name,
		FQN:         ast.CreateFQN(ns.FQN, name),
		Rules:       map[string]*Rule{},
		RuleExports: map[string]*ExportedRule{},
		Shapes:      map[string]*Shape{},
	}
}

func testRule(p *Policy, name string, body ast.Expression) *Rule {
	node := ast.NewRuleStatement(name, nil, nil, body, testRange())
	return &Rule{
		Node:   node,
		Policy: p,
		Name:   name,
		FQN:    ast.CreateFQN(p.FQN, name),
		Body:   body,
	}
}

func testShape(ns *Namespace, p *Policy, name string, with *ast.FQN) *Shape {
	stmt := ast.NewShapeStatement(name, nil, &ast.Cmplx{
		Range:  testRange(),
		With:   with,
		Fields: map[string]*ast.ShapeField{},
	}, testRange())
	return &Shape{
		Statement: stmt,
		Namespace: ns,
		Policy:    p,
		Name:      name,
		FQN:       ast.CreateFQN(cmpShapeBaseFQN(ns, p), name),
		Model: &ShapeModel{
			WithFQN: with,
			Fields:  map[string]*ShapeModelField{},
		},
	}
}

type scriptedCtx struct {
	context.Context
	call int
	at   map[int]error
}

func (s *scriptedCtx) Deadline() (time.Time, bool) { return time.Time{}, false }
func (s *scriptedCtx) Done() <-chan struct{}       { return nil }
func (s *scriptedCtx) Value(key any) any           { return nil }
func (s *scriptedCtx) Err() error {
	s.call++
	if err, ok := s.at[s.call]; ok {
		return err
	}
	return nil
}

func cmpShapeBaseFQN(ns *Namespace, p *Policy) ast.FQN {
	if p != nil {
		return p.FQN
	}
	return ns.FQN
}

func TestDetectReferenceCycleCancelledBeforeRulesLoop(t *testing.T) {
	idx := CreateIndex()
	ns := testNamespace("n")
	p := testPolicy(ns, "p")
	p.Rules["r"] = testRule(p, "r", nil)
	ns.Policies[p.Name] = p
	idx.Namespaces[ns.FQN.String()] = ns

	ctx := &scriptedCtx{Context: context.Background(), at: map[int]error{1: errors.New("cancelled")}}
	err := idx.detectReferenceCycle(ctx)
	if err == nil {
		t.Fatalf("expected canceled reference cycle detection, got %v", err)
	}
}

func TestDetectReferenceCycleCancelledInsideRulesLoop(t *testing.T) {
	idx := CreateIndex()
	ns := testNamespace("n")
	p := testPolicy(ns, "p")
	r := testRule(p, "r", nil)
	p.Rules[r.Name] = r
	ns.Policies[p.Name] = p
	idx.Namespaces[ns.FQN.String()] = ns

	ctx := &scriptedCtx{Context: context.Background(), at: map[int]error{2: errors.New("cancelled")}}
	err := idx.detectReferenceCycle(ctx)
	if err == nil {
		t.Fatalf("expected canceled reference cycle detection in rules loop, got %v", err)
	}
}

func TestDetectRuleCycleCancelledAtNamespaceLoop(t *testing.T) {
	idx := CreateIndex()
	ns := testNamespace("n")
	idx.Namespaces[ns.FQN.String()] = ns

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := idx.detectRuleCycle(ctx)
	if err == nil {
		t.Fatalf("expected canceled rule cycle detection, got %v", err)
	}
}

func TestDetectRuleCycleCancelledAtPolicyLoop(t *testing.T) {
	idx := CreateIndex()
	ns := testNamespace("n")
	p := testPolicy(ns, "p")
	p.Rules["r"] = testRule(p, "r", nil)
	ns.Policies[p.Name] = p
	idx.Namespaces[ns.FQN.String()] = ns

	ctx := &scriptedCtx{Context: context.Background(), at: map[int]error{1: errors.New("cancelled")}}
	_, err := idx.detectRuleCycle(ctx)
	if err == nil {
		t.Fatalf("expected canceled rule cycle detection at policy loop, got %v", err)
	}
}

func TestDetectRuleCycleCancelledInSecondPassChecks(t *testing.T) {
	makeIdx := func() *Index {
		idx := CreateIndex()
		ns := testNamespace("n")
		p := testPolicy(ns, "p")
		importExpr := ast.NewImportClause("imported", ast.NewFQN([]string{"other", "policy"}, testRange()).Ptr(), nil, testRange())
		p.Rules["r"] = testRule(p, "r", importExpr)
		ns.Policies[p.Name] = p
		idx.Namespaces[ns.FQN.String()] = ns
		return idx
	}

	if _, err := makeIdx().detectRuleCycle(&scriptedCtx{Context: context.Background(), at: map[int]error{2: errors.New("cancelled")}}); err == nil {
		t.Fatal("expected cancellation at second-pass namespace check")
	}
	if _, err := makeIdx().detectRuleCycle(&scriptedCtx{Context: context.Background(), at: map[int]error{3: errors.New("cancelled")}}); err == nil {
		t.Fatal("expected cancellation at second-pass policy check")
	}
	if _, err := makeIdx().detectRuleCycle(&scriptedCtx{Context: context.Background(), at: map[int]error{4: errors.New("cancelled")}}); err == nil {
		t.Fatal("expected cancellation at second-pass rule check")
	}
}

func TestDetectShapeCycleCancelledAtNamespaceLoop(t *testing.T) {
	idx := CreateIndex()
	ns := testNamespace("n")
	idx.Namespaces[ns.FQN.String()] = ns

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := idx.detectShapeCycle(ctx)
	if err == nil {
		t.Fatalf("expected canceled shape cycle detection, got %v", err)
	}
}

func TestDetectShapeCycleCancelledInChecks(t *testing.T) {
	makeIdx := func() *Index {
		idx := CreateIndex()
		ns := testNamespace("n")
		s := testShape(ns, nil, "base", nil)
		ns.Shapes[s.Name] = s
		p := testPolicy(ns, "p")
		p.Shapes["s"] = testShape(ns, p, "s", nil)
		ns.Policies[p.Name] = p
		idx.Namespaces[ns.FQN.String()] = ns
		return idx
	}
	for _, call := range []int{1, 2, 3, 4, 5} {
		if _, err := makeIdx().detectShapeCycle(&scriptedCtx{Context: context.Background(), at: map[int]error{call: errors.New("cancelled")}}); err == nil {
			t.Fatalf("expected cancellation at detectShapeCycle err-check call %d", call)
		}
	}
}

func TestDetectShapeCyclePolicyShapeSelfReferenceAddEdgeError(t *testing.T) {
	idx := CreateIndex()
	ns := testNamespace("n")
	p := testPolicy(ns, "p")
	shape := testShape(ns, p, "self", ast.NewFQN([]string{"self"}, testRange()).Ptr())
	p.Shapes[shape.Name] = shape
	ns.Policies[p.Name] = p
	ns.Shapes["self"] = shape
	idx.Namespaces[ns.FQN.String()] = ns

	_, err := idx.detectShapeCycle(context.Background())
	if err == nil {
		t.Fatal("expected self-loop add edge error")
	}
	if got := err.Error(); got == "" || !strings.Contains(got, "error adding edge") {
		t.Fatalf("expected add-edge error, got %v", err)
	}
}

// Policy-branch AddEdge wrapping is reachable when the composing shape exists only on the policy,
// the dependency is a namespace alias skipped by the first edge pass, and both share the same FQN string.
func TestDetectShapeCyclePolicyBranchAddEdgeErrorDuplicateFQN(t *testing.T) {
	rng := testRange()
	ns := testNamespace("n")
	p := testPolicy(ns, "pol")

	stmtA := ast.NewShapeStatement("dup", ast.NewStringTypeRef(rng), nil, rng)
	shapeA, err := createShape(ns, nil, stmtA)
	if err != nil {
		t.Fatal(err)
	}

	cmplx := &ast.Cmplx{
		Range:  rng,
		With:   ast.NewFQN([]string{"dup"}, rng).Ptr(),
		Fields: map[string]*ast.ShapeField{},
	}
	stmtB := ast.NewShapeStatement("dup", nil, cmplx, rng)
	shapeB, err := createShape(ns, p, stmtB)
	if err != nil {
		t.Fatal(err)
	}

	shared := ast.CreateFQN(p.FQN, "dup")
	shapeA.FQN = shared
	shapeB.FQN = shared

	ns.Shapes["dup"] = shapeA
	p.Shapes["dup"] = shapeB
	ns.Policies[p.Name] = p

	idx := CreateIndex()
	idx.Namespaces[ns.FQN.String()] = ns

	_, err = idx.detectShapeCycle(context.Background())
	if err == nil {
		t.Fatal("expected policy-branch add-edge error")
	}
	if !strings.Contains(err.Error(), "error adding edge") {
		t.Fatalf("unexpected error: %v", err)
	}
}
