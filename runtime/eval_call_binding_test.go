// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"strings"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/index"
)

// Calls through a let-bound or parameter-bound callable (#161).

func (s *RuntimeTestSuite) TestCallLetBoundLambda() {
	ctx := s.T().Context()
	src := `namespace com/ex

derive twice = (n: number): number => { yield n * 2 }

policy pol {
  fact k?: number as k default 10
  let f = (x) => { yield x * 2 }
  let g = (x: number, y?: number) => { yield y is defined ? x + y : x + k }
  let count = (x) => { yield x + 100 }
  rule policyLet = { yield f(2) }
  rule optionalOmitted = { yield g(1) }
  rule optionalGiven = { yield g(1, 2) }
  rule blockLet = { let h = (x) => { yield x - 1 } yield h(5) }
  rule shadowsBuiltin = { yield count(1) }
  rule blockShadowsDerive = { let twice = (n) => { yield n * 3 } yield twice(2) }
  rule viaDerive = { yield twice(2) }
  rule higherOrder = { let apply = (fn, v) => { yield fn(v) } yield apply(f, 4) }
  rule derivePassed = { let apply = (fn, v) => { yield fn(v) } yield apply(twice, 5) }
  export decision of policyLet
  export decision of optionalOmitted
  export decision of optionalGiven
  export decision of blockLet
  export decision of shadowsBuiltin
  export decision of blockShadowsDerive
  export decision of viaDerive
  export decision of higherOrder
  export decision of derivePassed
}
`
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "letcall.sentrie", src: src})
	for rule, want := range map[string]float64{
		"policyLet":          4,
		"optionalOmitted":    11,
		"optionalGiven":      3,
		"blockLet":           4,
		"shadowsBuiltin":     101,
		"blockShadowsDerive": 6,
		"viaDerive":          4,
		"higherOrder":        8,
		"derivePassed":       10,
	} {
		s.Run(rule, func() {
			out, err := exec.ExecRule(ctx, "com/ex", "pol", rule, nil)
			s.Require().NoError(err)
			s.Equal(want, s.decisionNumber(out))
		})
	}
}

func (s *RuntimeTestSuite) TestCallLambdaInsideDeriveBody() {
	ctx := s.T().Context()
	src := `namespace com/ex

derive scaled = (n: number): number => {
  let mul = (x) => { yield x * n }
  let addOne = (x) => { yield mul(x) + 1 }
  yield addOne(3)
}

policy pol {
  rule r = { yield scaled(2) }
  export decision of r
}
`
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "derivelet.sentrie", src: src})
	out, err := exec.ExecRule(ctx, "com/ex", "pol", "r", nil)
	s.Require().NoError(err)
	s.Equal(float64(7), s.decisionNumber(out))
}

// bindingCallEC returns a context for a test policy with name bound to v as a
// local, inside a derive body when inDerive is set.
func (s *RuntimeTestSuite) bindingCallEC(inDerive bool, name string, v box.Value) (*ExecutionContext, *executorImpl, *index.Policy) {
	exec := &executorImpl{}
	p := newEvalTestPolicy()
	ec := NewExecutionContext(p, exec)
	if inDerive {
		ec = ec.DetachedChildContext()
		ec.evalDerive = attachNamespaceDerive(p, "caller", stubNumberDeriveLambda(nil, nil, ast.NewIntegerLiteral(1, stubRange())))
	}
	ec.SetLocal(name, v, true)
	return ec, exec, p
}

func identCall(name string, memoized bool, args ...ast.Expression) *ast.CallExpression {
	return ast.NewCallExpression(ast.NewIdentifier(name, stubRange()), args, memoized, nil, stubRange())
}

func doubleLambda() *ast.LambdaExpression {
	return stubLambda([]string{"x"}, ast.NewInfixExpression(
		ast.NewIdentifier("x", stubRange()), ast.NewIntegerLiteral(2, stubRange()), "*", stubRange()))
}

func (s *RuntimeTestSuite) TestGetTargetBindingLambdaArityErrors() {
	ctx := s.T().Context()
	ec, exec, p := s.bindingCallEC(false, "f", box.Undefined())
	ec.SetLocal("f", box.Callable(newLambdaCallable(doubleLambda(), ec)), true)

	target, err := getTarget(ctx, ec, exec, p, identCall("f", false))
	s.Require().NoError(err)
	out, err := target(ctx, box.Number(3))
	s.Require().NoError(err)
	n, _ := out.NumberValue()
	s.Equal(float64(6), n)

	_, err = target(ctx, box.Number(1), box.Number(2))
	s.Require().ErrorContains(err, "too many arguments: want at most 1, got 2")
	_, err = target(ctx)
	s.Require().ErrorContains(err, "not enough arguments: want at least 1, got 0")
}

func (s *RuntimeTestSuite) TestGetTargetBindingNotCallable() {
	ec, exec, p := s.bindingCallEC(false, "v", box.Number(5))
	_, err := getTarget(s.T().Context(), ec, exec, p, identCall("v", false))
	s.Require().ErrorContains(err, `cannot call "v": it is a number value, not a function`)
}

func (s *RuntimeTestSuite) TestGetTargetBindingRefusesMemoizedCall() {
	ec, exec, p := s.bindingCallEC(false, "f", box.Undefined())
	ec.SetLocal("f", box.Callable(newLambdaCallable(doubleLambda(), ec)), true)
	_, err := getTarget(s.T().Context(), ec, exec, p, identCall("f", true))
	s.Require().ErrorContains(err, `"f" is a local value and cannot be memoized`)
}

func (s *RuntimeTestSuite) TestGetTargetUnknownFunctionIsNotAnImportError() {
	ec, exec, p := s.bindingCallEC(false, "other", box.Number(1))
	_, err := getTarget(s.T().Context(), ec, exec, p, identCall("nope", false))
	s.Require().ErrorContains(err, `unknown function "nope"`)
	s.NotContains(err.Error(), "import")

	ec, exec, p = s.bindingCallEC(true, "other", box.Number(1))
	_, err = getTarget(s.T().Context(), ec, exec, p, identCall("nope", false))
	s.Require().ErrorContains(err, `unknown function "nope"`)
	s.NotContains(err.Error(), "import")
}

func (s *RuntimeTestSuite) TestGetTargetInsideDeriveRefusesOutsideLambda() {
	ctx := s.T().Context()
	outside := NewExecutionContext(newEvalTestPolicy(), &executorImpl{})
	ec, exec, p := s.bindingCallEC(true, "f", box.Callable(newLambdaCallable(doubleLambda(), outside)))
	_, err := getTarget(ctx, ec, exec, p, identCall("f", false))
	s.Require().ErrorContains(err, `calling "f" is not permitted inside a derive - the lambda was defined outside the derive`)
}

func (s *RuntimeTestSuite) TestGetTargetInsideDeriveAllowsDeriveLambdaAndDerive() {
	ctx := s.T().Context()
	ec, exec, p := s.bindingCallEC(true, "f", box.Undefined())
	// A lambda created while evaluating a derive body captures a context
	// with evalDerive set.
	ec.SetLocal("f", box.Callable(newLambdaCallable(doubleLambda(), ec)), true)
	target, err := getTarget(ctx, ec, exec, p, identCall("f", false))
	s.Require().NoError(err)
	out, err := target(ctx, box.Number(4))
	s.Require().NoError(err)
	n, _ := out.NumberValue()
	s.Equal(float64(8), n)

	d := attachNamespaceDerive(p, "seven", stubNumberDeriveLambda(nil, nil, ast.NewIntegerLiteral(7, stubRange())))
	ec.SetLocal("g", box.Callable(&deriveCallable{derive: d}), true)
	_, err = getTarget(ctx, ec, exec, p, identCall("g", false))
	s.Require().NoError(err)
}

// Inside a derive body a fact is not a binding, so a fact name falls through
// to derives and builtins like any unknown name.
func (s *RuntimeTestSuite) TestGetTargetInsideDeriveFactIsNotABinding() {
	ec, exec, p := s.bindingCallEC(true, "other", box.Number(1))
	p.Facts["count"] = ast.NewFactStatement("count", ast.NewNumberTypeRef(stubRange()), "count", nil, false, stubRange())
	target, err := getTarget(s.T().Context(), ec, exec, p, identCall("count", false))
	s.Require().NoError(err)
	out, err := target(s.T().Context(), box.List([]box.Value{box.Number(1), box.Number(2)}))
	s.Require().NoError(err)
	n, _ := out.NumberValue()
	s.Equal(float64(2), n)
}

// Unbounded recursion through a let-bound lambda stops at MaxCallDepth with an
// ordinary error instead of overflowing the Go stack. Neither form below is
// rejected statically: validate does not see a let's reference to itself, and
// self-application names no other let.
func (s *RuntimeTestSuite) TestLambdaRecursionStopsAtCallDepthLimit() {
	ctx := s.T().Context()
	src := `namespace com/ex

policy pol {
  let loop = (n) => { yield loop(n + 1) }
  let self = (fn) => { yield fn(fn) }
  let down = (n) => { yield n <= 0 ? 0 : down(n - 1) + 1 }
  rule selfCall = { yield loop(0) }
  rule selfApplication = { yield self(self) }
  rule blockSelfCall = { let spin = (n) => { yield spin(n) } yield spin(1) }
  rule belowLimit = { yield down(999) }
  export decision of selfCall
  export decision of selfApplication
  export decision of blockSelfCall
  export decision of belowLimit
}
`
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "recursion.sentrie", src: src})
	for _, rule := range []string{"selfCall", "selfApplication", "blockSelfCall"} {
		s.Run(rule, func() {
			_, err := exec.ExecRule(ctx, "com/ex", "pol", rule, nil)
			s.Require().Error(err)
			s.Require().ErrorIs(err, ErrCallDepthExceeded)
			s.Contains(err.Error(), "lambda call depth exceeded (limit 1000)")
			// The depth error is not wrapped once per level.
			s.Less(len(err.Error()), 500)
		})
	}

	// Recursion that ends inside the limit still evaluates.
	out, err := exec.ExecRule(ctx, "com/ex", "pol", "belowLimit", nil)
	s.Require().NoError(err)
	s.Equal(float64(999), s.decisionNumber(out))
}

// Recursion between let lambdas validates: a lambda body is not evaluated when
// its let is, so the reference cycle check does not follow it. Recursion that
// ends returns its result; recursion that does not stops at the depth limit.
func (s *RuntimeTestSuite) TestRecursiveLetLambdas() {
	ctx := s.T().Context()
	src := `namespace com/ex

policy pol {
  let fac = (n) => { let v = n <= 1 ? 1 : n * fac(n - 1)  yield v }
  let isEven = (n) => { yield n == 0 ? true : isOdd(n - 1) }
  let isOdd = (n) => { yield n == 0 ? false : isEven(n - 1) }
  let a = (n) => { yield b(n) }
  let b = (n) => { yield a(n) }
  rule factorial = { yield fac(5) }
  rule tenIsEven = { yield isEven(10) ? 1 : 0 }
  rule sevenIsEven = { yield isEven(7) ? 1 : 0 }
  rule blockMutual = { let r = (n) => { yield n <= 0 ? 0 : q(n - 1) + 1 } let q = (n) => { yield r(n) } yield r(6) }
  rule endless = { yield a(1) }
  export decision of factorial
  export decision of tenIsEven
  export decision of sevenIsEven
  export decision of blockMutual
  export decision of endless
}
`
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "mutual.sentrie", src: src})
	for rule, want := range map[string]float64{"factorial": 120, "tenIsEven": 1, "sevenIsEven": 0, "blockMutual": 6} {
		s.Run(rule, func() {
			out, err := exec.ExecRule(ctx, "com/ex", "pol", rule, nil)
			s.Require().NoError(err)
			s.Equal(want, s.decisionNumber(out))
		})
	}
	_, err := exec.ExecRule(ctx, "com/ex", "pol", "endless", nil)
	s.Require().ErrorIs(err, ErrCallDepthExceeded)
}

// One level of a recursive lambda costs Go stack in proportion to how deeply
// the recursive call is nested in its body, so MaxCallDepth alone would let a
// heavy body overflow the stack. The eval frame limit trips first.
func (s *RuntimeTestSuite) TestHeavyRecursiveLambdaBodyStopsAtEvalDepthLimit() {
	ctx := s.T().Context()
	body := strings.Repeat("(1 + ", 400) + "loop(n + 1)" + strings.Repeat(")", 400)
	src := "namespace com/ex\n\npolicy pol {\n  let loop = (n) => { yield " + body + " }\n" +
		"  rule r = { yield loop(0) }\n  export decision of r\n}\n"
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "heavy.sentrie", src: src})
	_, err := exec.ExecRule(ctx, "com/ex", "pol", "r", nil)
	s.Require().ErrorIs(err, ErrEvalDepthExceeded)
	s.Contains(err.Error(), "evaluation depth exceeded (limit 50000 nested evaluations)")
	s.Less(len(err.Error()), 500)
}

// An imported decision is evaluated in a fresh root context, which takes over
// the importing evaluation's call depth and eval frame counter.
func (s *RuntimeTestSuite) TestImportedDecisionInheritsDepthBudget() {
	ctx := s.T().Context()
	src := `namespace com/ex

policy target {
  let f = (n) => { yield n + 1 }
  rule r = { yield f(1) }
  export decision of r
}
`
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "target.sentrie", src: src})

	caller := NewExecutionContext(newEvalTestPolicy(), exec)
	out, err := exec.execRuleFrom(ctx, "com/ex", "target", "r", nil, caller)
	s.Require().NoError(err)
	s.Equal(float64(2), s.decisionNumber(out))

	caller.callDepth = MaxCallDepth
	_, err = exec.execRuleFrom(ctx, "com/ex", "target", "r", nil, caller)
	s.Require().ErrorIs(err, ErrCallDepthExceeded)

	caller = NewExecutionContext(newEvalTestPolicy(), exec)
	caller.evalDepth.n.Store(MaxEvalDepth)
	_, err = exec.execRuleFrom(ctx, "com/ex", "target", "r", nil, caller)
	s.Require().ErrorIs(err, ErrEvalDepthExceeded)
}

// The derive invocation path shares the depth budget.
func (s *RuntimeTestSuite) TestInvokeDeriveStopsAtCallDepthLimit() {
	ctx := s.T().Context()
	p := newEvalTestPolicy()
	exec := &executorImpl{}
	ec := NewExecutionContext(p, exec)
	ec.callDepth = MaxCallDepth
	d := attachNamespaceDerive(p, "seven", stubNumberDeriveLambda(nil, nil, ast.NewIntegerLiteral(7, stubRange())))
	_, err := invokeDerive(ctx, ec, exec, p, d, nil)
	s.Require().ErrorIs(err, ErrCallDepthExceeded)
	s.Contains(err.Error(), "derive call depth exceeded (limit 1000)")

	ec.callDepth = MaxCallDepth - 1
	out, err := invokeDerive(ctx, ec, exec, p, d, nil)
	s.Require().NoError(err)
	n, _ := out.NumberValue()
	s.Equal(float64(7), n)
}

// Each invocation sets its child's depth from the call site, and child
// contexts inherit it, so a nested call is counted even when the lambda was
// captured in a shallower context.
func (s *RuntimeTestSuite) TestLambdaCallDepthFollowsCallSite() {
	ctx := s.T().Context()
	p := newEvalTestPolicy()
	exec := &executorImpl{}
	capture := NewExecutionContext(p, exec)
	site := capture.AttachedChildContext().AttachedChildContext()
	site.callDepth = MaxCallDepth
	s.Equal(MaxCallDepth, site.AttachedChildContext().callDepth)
	s.Equal(MaxCallDepth, site.DetachedChildContext().callDepth)

	fn := newLambdaCallable(doubleLambda(), capture)
	_, err := fn.Invoke(ctx, &CallSite{EC: site, Exec: exec, Policy: p}, []box.Value{box.Number(1)})
	s.Require().ErrorIs(err, ErrCallDepthExceeded)

	site.callDepth = 0
	out, err := fn.Invoke(ctx, &CallSite{EC: site, Exec: exec, Policy: p}, []box.Value{box.Number(1)})
	s.Require().NoError(err)
	n, _ := out.NumberValue()
	s.Equal(float64(2), n)
}
