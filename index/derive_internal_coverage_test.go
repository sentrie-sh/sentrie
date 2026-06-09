// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"context"
	"fmt"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/parser"
	"github.com/sentrie-sh/sentrie/tokens"
	"github.com/sentrie-sh/sentrie/trinary"
)

func (s *IndexTestSuite) TestWalkDeriveExprDFSVisitsAllExpressionKinds() {
	r := deriveCovRng(1)
	root := ast.NewInfixExpression(
		ast.NewUnaryExpression("-", ast.NewIntegerLiteral(1, r), r),
		ast.NewTernaryExpression(
			ast.NewTrinaryLiteral(trinary.True, r),
			ast.NewIntegerLiteral(1, r),
			ast.NewIntegerLiteral(0, r),
			r,
		),
		"+",
		r,
	)
	var count int
	err := walkDeriveExprDFS(root, func(ast.Expression) error {
		count++
		return nil
	})
	s.NoError(err)
	s.Greater(count, 4)

	elvis := ast.NewTernaryElvis(ast.NewNullLiteral(r), ast.NewIntegerLiteral(2, r), r)
	s.NoError(walkDeriveExprDFS(elvis, func(ast.Expression) error { return nil }))

	block := ast.NewBlockExpression(
		[]ast.Statement{ast.NewVarDeclaration("x", nil, ast.NewIntegerLiteral(1, r), r)},
		ast.NewListLiteral([]ast.Expression{ast.NewIntegerLiteral(1, r)}, r),
		r,
	)
	s.NoError(walkDeriveExprDFS(block, func(ast.Expression) error { return nil }))

	complex := ast.NewCallExpression(
		ast.NewIdentifier("count", r),
		[]ast.Expression{
			ast.NewTransformExpression(ast.NewListLiteral(nil, r), ".length", r),
			ast.NewCastExpression(ast.NewIntegerLiteral(1, r), ast.NewNumberTypeRef(r), r),
			ast.NewIsDefinedExpression(ast.NewIdentifier("x", r), r),
			ast.NewIsEmptyExpression(ast.NewListLiteral(nil, r), r),
			ast.NewTrailingCommentExpression("// c", ast.NewIntegerLiteral(1, r), r),
			ast.NewPrecedingCommentExpression("// p", ast.NewIntegerLiteral(2, r), r),
		},
		false, nil, r,
	)
	s.NoError(walkDeriveExprDFS(complex, func(ast.Expression) error { return nil }))

	s.NoError(walkDeriveExprDFS(
		ast.NewFieldAccessExpression(ast.NewIdentifier("mod", r), "fn", r),
		func(ast.Expression) error { return nil },
	))
	s.NoError(walkDeriveExprDFS(
		ast.NewIndexAccessExpression(ast.NewListLiteral(nil, r), ast.NewIntegerLiteral(0, r), r),
		func(ast.Expression) error { return nil },
	))
	s.NoError(walkDeriveExprDFS(
		ast.NewMapLiteral([]ast.MapEntry{{Key: ast.NewStringLiteral("a", r), Value: ast.NewIntegerLiteral(1, r)}}, r),
		func(ast.Expression) error { return nil },
	))
}

func (s *IndexTestSuite) TestWalkDeriveExprDFSPropagatesVisitError() {
	r := deriveCovRng(2)
	err := walkDeriveExprDFS(ast.NewIntegerLiteral(1, r), func(ast.Expression) error {
		return fmt.Errorf("stop")
	})
	s.Error(err)
}

func (s *IndexTestSuite) TestScanLambdasOutsideCallsBranches() {
	r := deriveCovRng(3)
	s.NoError(scanLambdasOutsideCalls(ast.NewTernaryExpression(
		ast.NewTrinaryLiteral(trinary.True, r),
		ast.NewIntegerLiteral(1, r),
		ast.NewIntegerLiteral(0, r),
		r,
	), false))
	s.NoError(scanLambdasOutsideCalls(ast.NewTernaryElvis(ast.NewNullLiteral(r), ast.NewIntegerLiteral(1, r), r), false))

	inner := ast.NewLambdaExpression(
		[]string{"x"},
		ast.NewBlockExpression(nil, ast.NewIdentifier("x", r), r),
		r,
	)
	s.Error(scanLambdasOutsideCalls(inner, false))
	s.NoError(scanLambdasOutsideCalls(
		ast.NewCallExpression(ast.NewIdentifier("any", r), []ast.Expression{ast.NewListLiteral(nil, r), inner}, false, nil, r),
		false,
	))

	block := ast.NewBlockExpression(
		[]ast.Statement{ast.NewVarDeclaration("x", nil, ast.NewIntegerLiteral(1, r), r)},
		ast.NewInfixExpression(ast.NewIntegerLiteral(1, r), ast.NewIntegerLiteral(2, r), "+", r),
		r,
	)
	s.NoError(scanLambdasOutsideCalls(block, false))
}

func (s *IndexTestSuite) TestValidateDerivePureNilBodyAndWalkBranches() {
	idx := CreateIndex()
	nsFQN := ast.NewFQN([]string{"com", "ex"}, deriveCovRng(4))
	ns := &Namespace{FQN: nsFQN, Derives: map[string]*Derive{}}
	r := deriveCovRng(5)

	s.NoError(validateDerivePure(idx, &Derive{Lambda: nil}))
	s.NoError(validateDerivePure(idx, &Derive{Lambda: &ast.LambdaExpression{Body: nil}}))

	helper := &Derive{
		Name: "helper", FQN: ast.CreateFQN(nsFQN, "helper"), Namespace: ns,
		Lambda: ast.NewLambdaExpression(nil, ast.NewBlockExpression(nil, ast.NewIntegerLiteral(1, r), r), r),
	}
	ns.Derives["helper"] = helper

	d := &Derive{
		Name: "caller", FQN: ast.CreateFQN(nsFQN, "caller"), Namespace: ns,
		Lambda: ast.NewLambdaExpression(
			[]string{"n"},
			ast.NewBlockExpression(
				nil,
				ast.NewCallExpression(ast.NewIdentifier("helper", r), nil, false, nil, r),
				r,
			),
			r,
		),
		DefineShort: map[string]*Derive{"helper": helper},
		DefineFQN:   map[string]*Derive{"com/ex/helper": helper},
	}
	s.NoError(validateDerivePure(idx, d))

	paramCall := &Derive{
		Name: "paramCall", FQN: ast.CreateFQN(nsFQN, "paramCall"), Namespace: ns,
		Lambda: ast.NewLambdaExpression(
			[]string{"f"},
			ast.NewBlockExpression(
				nil,
				ast.NewCallExpression(ast.NewIdentifier("f", r), []ast.Expression{ast.NewIntegerLiteral(1, r)}, false, nil, r),
				r,
			),
			r,
		),
	}
	s.NoError(validateDerivePure(idx, paramCall))

	s.Error(validateDerivePure(idx, &Derive{
		Name: "badCall", FQN: ast.CreateFQN(nsFQN, "badCall"), Namespace: ns,
		Lambda: ast.NewLambdaExpression(
			nil,
			ast.NewBlockExpression(
				nil,
				ast.NewCallExpression(
					ast.NewInfixExpression(ast.NewIntegerLiteral(1, r), ast.NewIntegerLiteral(2, r), "+", r),
					[]ast.Expression{ast.NewIntegerLiteral(3, r)},
					false, nil, r,
				),
				r,
			),
			r,
		),
	}))
}

func (s *IndexTestSuite) TestIsAllowedDeriveCallbackArgDefaultBranch() {
	s.False(isAllowedDeriveCallbackArg("count", 1))
	s.True(isAllowedDeriveCallbackArg("distinct", 1))
}

func (s *IndexTestSuite) TestWalkDeriveBlockAndScanLambdasRejectUnsupportedStatements() {
	r := deriveCovRng(6)
	badStmt := ast.NewBlockExpression(
		[]ast.Statement{ast.NewCommentStatement("nope", r)},
		ast.NewIntegerLiteral(1, r),
		r,
	)
	err := walkDeriveBlock(CreateIndex(), &Derive{}, badStmt, map[string]struct{}{})
	s.Error(err)

	err = scanLambdasOutsideCalls(badStmt, false)
	s.Error(err)
}

func (s *IndexTestSuite) TestWalkDeriveExprSeenCyclicAndLambdaBody() {
	idx := CreateIndex()
	r := deriveCovRng(7)
	shared := ast.NewIdentifier("x", r)
	cyclicRoot := ast.NewInfixExpression(
		shared,
		ast.NewInfixExpression(shared, ast.NewIntegerLiteral(1, r), "+", r),
		"+",
		r,
	)
	d := &Derive{Name: "c", FQN: ast.NewFQN([]string{"com", "ex", "c"}, r)}
	s.Error(walkDeriveExprSeen(idx, d, cyclicRoot, map[string]struct{}{}, map[ast.Expression]struct{}{}))

	lam := ast.NewLambdaExpression(
		[]string{"p"},
		ast.NewBlockExpression(nil, ast.NewIdentifier("p", r), r),
		r,
	)
	s.NoError(walkDeriveExprSeen(idx, d, lam, map[string]struct{}{}, map[ast.Expression]struct{}{}))
}

func (s *IndexTestSuite) TestCheckDeriveIdentifierPolicyBindings() {
	r := deriveCovRng(8)
	nsFQN := ast.NewFQN([]string{"com", "ex"}, r)
	pol := &Policy{
		Name: "pol", FQN: ast.CreateFQN(nsFQN, "pol"), Namespace: &Namespace{FQN: nsFQN},
		Facts: map[string]*ast.FactStatement{"user": {}},
		Rules: map[string]*Rule{"gate": {}},
	}
	d := &Derive{Name: "d", FQN: ast.CreateFQN(nsFQN, "d"), Namespace: &Namespace{FQN: nsFQN}, Policy: pol}

	s.NoError(checkDeriveIdentifier(d, "param", map[string]struct{}{"param": {}}))
	s.Error(checkDeriveIdentifier(d, "user", map[string]struct{}{}))
	s.Error(checkDeriveIdentifier(d, "gate", map[string]struct{}{}))
	s.Error(checkDeriveIdentifier(d, "helper", map[string]struct{}{}))
}

func (s *IndexTestSuite) TestCheckDeriveCallDefineFQNVisibilityAndFilterCallback() {
	idx := CreateIndex()
	r := deriveCovRng(9)
	nsAlpha := ast.NewFQN([]string{"com", "alpha"}, r)
	nsBeta := ast.NewFQN([]string{"com", "beta"}, r)
	secret := &Derive{
		Name: "secret", FQN: ast.CreateFQN(ast.CreateFQN(nsAlpha, "polA"), "secret"),
		Namespace: &Namespace{FQN: nsAlpha},
		Policy:    &Policy{Name: "polA", FQN: ast.CreateFQN(nsAlpha, "polA"), Namespace: &Namespace{FQN: nsAlpha}},
		Lambda:    ast.NewLambdaExpression(nil, ast.NewBlockExpression(nil, ast.NewIntegerLiteral(1, r), r), r),
	}
	idx.DerivesByFQN[secret.FQN.String()] = secret

	caller := &Derive{
		Name: "caller", FQN: ast.CreateFQN(nsBeta, "caller"), Namespace: &Namespace{FQN: nsBeta},
		Lambda: ast.NewLambdaExpression(
			nil,
			ast.NewBlockExpression(
				nil,
				ast.NewCallExpression(slashCallee(r, "com", "alpha", "polA", "secret"), nil, false, nil, r),
				r,
			),
			r,
		),
		DefineFQN: map[string]*Derive{secret.FQN.String(): secret},
	}
	call := caller.Lambda.Body.Yield.(*ast.CallExpression)
	s.Error(checkDeriveCall(idx, caller, call, map[string]struct{}{}, map[ast.Expression]struct{}{}))

	isOne := &Derive{
		Name: "isOne", FQN: ast.CreateFQN(nsBeta, "isOne"), Namespace: &Namespace{FQN: nsBeta},
		Lambda: ast.NewLambdaExpression(
			[]string{"n"},
			ast.NewBlockExpression(nil, ast.NewTrinaryLiteral(trinary.True, r), r),
			r,
		),
	}
	caller.DefineShort = map[string]*Derive{"isOne": isOne}
	filterCall := ast.NewCallExpression(
		ast.NewIdentifier("filter", r),
		[]ast.Expression{
			ast.NewListLiteral([]ast.Expression{ast.NewIntegerLiteral(1, r)}, r),
			ast.NewIdentifier("isOne", r),
		},
		false, nil, r,
	)
	s.NoError(checkDeriveCall(idx, caller, filterCall, map[string]struct{}{}, map[ast.Expression]struct{}{}))
}

func (s *IndexTestSuite) TestForEachDeriveExprChildPropagatesChildError() {
	r := deriveCovRng(10)
	err := forEachDeriveExprChild(
		ast.NewInfixExpression(ast.NewIntegerLiteral(1, r), ast.NewIntegerLiteral(2, r), "+", r),
		func(ast.Expression) error { return fmt.Errorf("child fail") },
	)
	s.Error(err)
}

func (s *IndexTestSuite) TestWalkDeriveExprSeenComplexExpressionTree() {
	idx := CreateIndex()
	r := deriveCovRng(11)
	nsFQN := ast.NewFQN([]string{"com", "ex"}, r)
	d := &Derive{Name: "d", FQN: ast.CreateFQN(nsFQN, "d"), Namespace: &Namespace{FQN: nsFQN}}

	expr := ast.NewInfixExpression(
		ast.NewUnaryExpression("-", ast.NewIntegerLiteral(1, r), r),
		ast.NewTernaryExpression(
			ast.NewTrinaryLiteral(trinary.True, r),
			ast.NewIntegerLiteral(1, r),
			ast.NewIntegerLiteral(0, r),
			r,
		),
		"+",
		r,
	)
	scope := map[string]struct{}{}
	s.NoError(walkDeriveExprSeen(idx, d, expr, scope, map[ast.Expression]struct{}{}))

	rich := ast.NewCallExpression(
		ast.NewIdentifier("count", r),
		[]ast.Expression{
			ast.NewListLiteral([]ast.Expression{ast.NewIntegerLiteral(1, r)}, r),
			ast.NewIndexAccessExpression(ast.NewListLiteral([]ast.Expression{ast.NewIntegerLiteral(1, r)}, r), ast.NewIntegerLiteral(0, r), r),
			ast.NewCastExpression(ast.NewIntegerLiteral(1, r), ast.NewNumberTypeRef(r), r),
			ast.NewTransformExpression(ast.NewListLiteral(nil, r), ".length", r),
			ast.NewIsDefinedExpression(ast.NewListLiteral(nil, r), r),
			ast.NewIsEmptyExpression(ast.NewListLiteral(nil, r), r),
			ast.NewTrailingCommentExpression("// t", ast.NewIntegerLiteral(1, r), r),
			ast.NewPrecedingCommentExpression("// p", ast.NewIntegerLiteral(2, r), r),
			ast.NewMapLiteral([]ast.MapEntry{{Key: ast.NewStringLiteral("k", r), Value: ast.NewIntegerLiteral(1, r)}}, r),
		},
		false, nil, r,
	)
	s.NoError(walkDeriveExprSeen(idx, d, rich, scope, map[ast.Expression]struct{}{}))
}

func (s *IndexTestSuite) TestDeriveCycleValidationCancelled() {
	ctx := s.T().Context()
	idx := CreateIndex()
	src := `namespace com/ex
derive a = () => { yield com/ex/b() }
derive b = () => { yield com/ex/a() }
policy pol {
  let _s = 0
  rule r = { yield true }
  export decision of r
}
`
	prog, err := parser.NewParserFromString(src, "cyc_cancel.sentra").ParseProgram(ctx)
	s.Require().NoError(err)
	s.Require().NoError(idx.AddProgram(ctx, prog))
	vctx, cancel := context.WithCancel(ctx)
	cancel()
	err = idx.Validate(vctx)
	s.Error(err)
}

func (s *IndexTestSuite) TestAddProgramDuplicateShapeRejected() {
	ctx := s.T().Context()
	idx := CreateIndex()
	src := `namespace com/ex
shape Dup { value: number }
shape Dup { other: string }
policy p {
  let _s = 0
  rule r = { yield true }
  export decision of r
}
`
	prog, err := parser.NewParserFromString(src, "dup_shape.sentra").ParseProgram(ctx)
	s.Require().NoError(err)
	err = idx.AddProgram(ctx, prog)
	s.Error(err)
}

func (s *IndexTestSuite) TestValidateRuleGraphElvisTernary() {
	ctx := s.T().Context()
	idx := CreateIndex()
	src := `namespace com/ex
policy pol {
  let _seed = 0
  rule gate = { yield _seed ?: true }
  rule ternary = { yield true ? 1 : 0 }
  export decision of gate
  export decision of ternary
}
`
	prog, err := parser.NewParserFromString(src, "elvis_rule.sentra").ParseProgram(ctx)
	s.Require().NoError(err)
	s.Require().NoError(idx.AddProgram(ctx, prog))
	s.Require().NoError(idx.Validate(ctx))
}

func (s *IndexTestSuite) TestScanLambdasCoversRemainingExpressionKinds() {
	r := deriveCovRng(12)
	inner := ast.NewLambdaExpression(
		[]string{"x"},
		ast.NewBlockExpression(nil, ast.NewIdentifier("x", r), r),
		r,
	)
	cases := []ast.Expression{
		ast.NewFieldAccessExpression(ast.NewIntegerLiteral(1, r), "f", r),
		ast.NewIndexAccessExpression(ast.NewListLiteral(nil, r), ast.NewIntegerLiteral(0, r), r),
		ast.NewListLiteral([]ast.Expression{inner}, r),
		ast.NewMapLiteral([]ast.MapEntry{{Key: ast.NewStringLiteral("k", r), Value: inner}}, r),
		ast.NewCastExpression(inner, ast.NewNumberTypeRef(r), r),
		ast.NewTransformExpression(ast.NewListLiteral(nil, r), ".length", r),
		ast.NewIsDefinedExpression(ast.NewListLiteral(nil, r), r),
		ast.NewIsEmptyExpression(ast.NewListLiteral(nil, r), r),
		ast.NewTrailingCommentExpression("// c", inner, r),
		ast.NewPrecedingCommentExpression("// p", inner, r),
	}
	for _, expr := range cases {
		s.NoError(scanLambdasOutsideCalls(expr, true))
	}
}

func (s *IndexTestSuite) TestWalkDeriveExprSeenBlockExpression() {
	idx := CreateIndex()
	r := deriveCovRng(13)
	d := &Derive{Name: "d", FQN: ast.NewFQN([]string{"com", "ex", "d"}, r), Namespace: &Namespace{FQN: ast.NewFQN([]string{"com", "ex"}, r)}}
	block := ast.NewBlockExpression(
		[]ast.Statement{ast.NewVarDeclaration("x", nil, ast.NewIntegerLiteral(1, r), r)},
		ast.NewIntegerLiteral(1, r),
		r,
	)
	s.NoError(walkDeriveExprSeen(idx, d, block, map[string]struct{}{}, map[ast.Expression]struct{}{}))
}

func (s *IndexTestSuite) TestCheckDeriveCallUnknownSlashDeriveNotInIndex() {
	idx := CreateIndex()
	r := deriveCovRng(14)
	d := &Derive{Name: "d", FQN: ast.NewFQN([]string{"com", "ex", "d"}, r), Namespace: &Namespace{FQN: ast.NewFQN([]string{"com", "ex"}, r)}}
	call := ast.NewCallExpression(slashCallee(r, "com", "ex", "missing"), nil, false, nil, r)
	s.Error(checkDeriveCall(idx, d, call, map[string]struct{}{}, map[ast.Expression]struct{}{}))
}

func (s *IndexTestSuite) TestWalkDeriveExprDFSChildWalkErrors() {
	r := deriveCovRng(15)
	call := ast.NewCallExpression(
		ast.NewIdentifier("count", r),
		[]ast.Expression{ast.NewIntegerLiteral(1, r), ast.NewIntegerLiteral(2, r)},
		false, nil, r,
	)
	seen := 0
	err := walkDeriveExprDFS(call, func(e ast.Expression) error {
		seen++
		if seen == 2 {
			return fmt.Errorf("stop")
		}
		return nil
	})
	s.Error(err)
}

func (s *IndexTestSuite) TestAddProgramTopLevelCommentAndExportDeriveConflict() {
	ctx := s.T().Context()
	idx := CreateIndex()
	prog, err := parser.NewParserFromString(`namespace com/ex
derive x = () => { yield 1 }
export derive x
policy p {
  let _s = 0
  rule r = { yield true }
  export decision of r
}
`, "exp_conflict.sentra").ParseProgram(ctx)
	s.Require().NoError(err)
	err = idx.AddProgram(ctx, prog)
	s.Require().NoError(err)
	prog2, err := parser.NewParserFromString(`namespace com/ex
export derive x
policy p2 {
  let _s = 0
  rule r2 = { yield true }
  export decision of r2
}
`, "exp_conflict2.sentra").ParseProgram(ctx)
	s.Require().NoError(err)
	err = idx.AddProgram(ctx, prog2)
	s.Error(err)
}

func (s *IndexTestSuite) TestAddProgramRejectsUnsupportedTopLevelStatement() {
	ctx := s.T().Context()
	idx := CreateIndex()
	prog := &ast.Program{
		Reference: "bad.sentrie",
		Statements: []ast.Statement{
			ast.NewNamespaceStatement(ast.NewFQN([]string{"com", "ex"}, deriveCovRng(16)), deriveCovRng(16)),
			ast.NewPolicyStatement("orphan", nil, deriveCovRng(17)),
		},
	}
	err := idx.AddProgram(ctx, prog)
	s.Error(err)
}

func (s *IndexTestSuite) TestForEachDeriveExprChildBlockAndTernaryErrors() {
	r := deriveCovRng(17)
	s.Error(forEachDeriveExprChild(
		ast.NewBlockExpression(
			[]ast.Statement{ast.NewCommentStatement("bad", r)},
			ast.NewIntegerLiteral(1, r),
			r,
		),
		func(ast.Expression) error { return nil },
	))

	tern := ast.NewTernaryExpression(
		ast.NewTrinaryLiteral(trinary.True, r),
		ast.NewIntegerLiteral(1, r),
		ast.NewIntegerLiteral(0, r),
		r,
	)
	s.Error(forEachDeriveExprChild(tern, func(ast.Expression) error { return fmt.Errorf("fail") }))

	elvis := ast.NewTernaryElvis(ast.NewNullLiteral(r), ast.NewIntegerLiteral(1, r), r)
	s.Error(forEachDeriveExprChild(elvis, func(ast.Expression) error { return fmt.Errorf("fail") }))

	call := ast.NewCallExpression(
		ast.NewIdentifier("count", r),
		[]ast.Expression{ast.NewIntegerLiteral(1, r), ast.NewIntegerLiteral(2, r)},
		false, nil, r,
	)
	s.Error(forEachDeriveExprChild(call, func(e ast.Expression) error {
		if _, ok := e.(*ast.IntegerLiteral); ok {
			return fmt.Errorf("fail on literal")
		}
		return nil
	}))
}

func (s *IndexTestSuite) TestAddProgramSkipsTopLevelCommentStatement() {
	ctx := s.T().Context()
	idx := CreateIndex()
	base, err := parser.NewParserFromString(`namespace com/ex
derive d = () => { yield 1 }
policy p {
  let _s = 0
  rule r = { yield true }
  export decision of r
}
`, "base.sentra").ParseProgram(ctx)
	s.Require().NoError(err)
	r := deriveCovRng(18)
	stmts := make([]ast.Statement, 0, len(base.Statements)+1)
	stmts = append(stmts, base.Statements[0], ast.NewCommentStatement("header", r))
	stmts = append(stmts, base.Statements[1:]...)
	prog := &ast.Program{Reference: "with_comment.sentrie", Statements: stmts}
	s.Require().NoError(idx.AddProgram(ctx, prog))
}

func (s *IndexTestSuite) TestAddProgramRejectsUnsupportedNamespaceLevelFact() {
	ctx := s.T().Context()
	idx := CreateIndex()
	r := deriveCovRng(19)
	prog := &ast.Program{
		Reference: "bad_fact.sentrie",
		Statements: []ast.Statement{
			ast.NewNamespaceStatement(ast.NewFQN([]string{"com", "ex"}, r), r),
			ast.NewFactStatement("user", ast.NewStringTypeRef(r), "u", nil, true, r),
		},
	}
	err := idx.AddProgram(ctx, prog)
	s.Error(err)
}

func (s *IndexTestSuite) TestDetectDeriveCycleDirectContextCancelled() {
	ctx := s.T().Context()
	idx := CreateIndex()
	s.Require().NoError(idx.AddProgram(ctx, programWithDeriveCycleOnly()))
	vctx, cancel := context.WithCancel(ctx)
	cancel()
	err := idx.detectDeriveCycle(vctx)
	s.Error(err)
}

func programWithDeriveCycleOnly() *ast.Program {
	r := deriveCovRng(20)
	body := ast.NewCallExpression(
		slashCalleeForProgram(r, "com", "ex", "b"),
		nil, false, nil, r,
	)
	deriveA := ast.NewDeriveStatement("a", ast.NewLambdaExpression(nil, ast.NewBlockExpression(nil, body, r), r), r)
	bodyB := ast.NewCallExpression(
		slashCalleeForProgram(r, "com", "ex", "a"),
		nil, false, nil, r,
	)
	deriveB := ast.NewDeriveStatement("b", ast.NewLambdaExpression(nil, ast.NewBlockExpression(nil, bodyB, r), r), r)
	pol := ast.NewPolicyStatement("pol", []ast.Statement{
		ast.NewFactStatement("user", ast.NewStringTypeRef(r), "u", nil, true, r),
		ast.NewRuleStatement("r", nil, nil, ast.NewTrinaryLiteral(trinary.True, r), r),
		ast.NewRuleExportStatement("r", nil, r),
	}, r)
	return &ast.Program{
		Reference: "cyc_only.sentrie",
		Statements: []ast.Statement{
			ast.NewNamespaceStatement(ast.NewFQN([]string{"com", "ex"}, r), r),
			deriveA, deriveB, pol,
		},
	}
}

func slashCalleeForProgram(r tokens.Range, parts ...string) ast.Expression {
	var out ast.Expression = ast.NewIdentifier(parts[0], r)
	for _, p := range parts[1:] {
		out = ast.NewInfixExpression(out, ast.NewIdentifier(p, r), "/", r)
	}
	return out
}

func (s *IndexTestSuite) TestWalkDeriveExprDFSVisitErrorOnFirstNode() {
	r := deriveCovRng(21)
	err := walkDeriveExprDFS(
		ast.NewInfixExpression(ast.NewIntegerLiteral(1, r), ast.NewIntegerLiteral(2, r), "+", r),
		func(e ast.Expression) error {
			if _, ok := e.(*ast.InfixExpression); ok {
				return fmt.Errorf("visit fail")
			}
			return nil
		},
	)
	s.Error(err)
}

func (s *IndexTestSuite) TestForEachDeriveExprChildListAndMapErrors() {
	r := deriveCovRng(22)
	list := ast.NewListLiteral([]ast.Expression{ast.NewIntegerLiteral(1, r), ast.NewIntegerLiteral(2, r)}, r)
	s.Error(forEachDeriveExprChild(list, func(e ast.Expression) error {
		if _, ok := e.(*ast.IntegerLiteral); ok {
			return fmt.Errorf("list child fail")
		}
		return nil
	}))
	m := ast.NewMapLiteral([]ast.MapEntry{
		{Key: ast.NewStringLiteral("k", r), Value: ast.NewIntegerLiteral(1, r)},
	}, r)
	s.Error(forEachDeriveExprChild(m, func(e ast.Expression) error {
		if _, ok := e.(*ast.IntegerLiteral); ok {
			return fmt.Errorf("map value fail")
		}
		return nil
	}))
}

func (s *IndexTestSuite) TestDeriveCycleDuplicateDependencyEdge() {
	ctx := s.T().Context()
	idx := CreateIndex()
	src := `namespace com/ex
derive b = () => { yield 1 }
derive a = () => { yield com/ex/b() + com/ex/b() }
policy pol {
  let _s = 0
  rule r = { yield a() == 2 }
  export decision of r
}
`
	prog, err := parser.NewParserFromString(src, "dup_edge.sentra").ParseProgram(ctx)
	s.Require().NoError(err)
	s.Require().NoError(idx.AddProgram(ctx, prog))
	s.Require().NoError(idx.Validate(ctx))
}

func slashCallee(r tokens.Range, parts ...string) ast.Expression {
	if len(parts) == 0 {
		return nil
	}
	var out ast.Expression = ast.NewIdentifier(parts[0], r)
	for _, p := range parts[1:] {
		out = ast.NewInfixExpression(out, ast.NewIdentifier(p, r), "/", r)
	}
	return out
}