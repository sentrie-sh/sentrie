// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"
)

// End-to-end coverage of calls through let-bound callables (#161), including
// the paths validate normally stops, which are run here with validate bypassed.

// execRuleUnvalidated runs rule without index validation, so runtime-only
// error paths stay reachable.
func (s *RuntimeTestSuite) execRuleUnvalidated(ctx context.Context, src, rule string) (*ExecutorOutput, error) {
	idx := s.mustBuildDeriveIndex(ctx, deriveTestProgram{name: "unvalidated.sentrie", src: src})
	raw, err := NewExecutor(idx)
	s.Require().NoError(err)
	return raw.ExecRule(ctx, "com/ex", "pol", rule, nil)
}

func (s *RuntimeTestSuite) TestCallLetLambdaClosuresAndChaining() {
	ctx := s.T().Context()
	src := `namespace com/ex

policy pol {
  fact base?: number as base default 100
  let offset = 7
  let addOffset = (x) => { yield x + offset }
  let addBase = (x) => { yield addOffset(x) + base }
  rule outerCapture = { yield addOffset(1) }
  rule chained = { yield addBase(1) }
  rule blockCapture = {
    let local = 10
    let addLocal = (x) => { yield x + local + offset }
    yield addLocal(1)
  }
  rule returnsLambda = {
    let make = (n) => { yield (x) => { yield x * n } }
    let triple = make(3)
    yield triple(4)
  }
  export decision of outerCapture
  export decision of chained
  export decision of blockCapture
  export decision of returnsLambda
}
`
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "closures.sentrie", src: src})
	for rule, want := range map[string]float64{
		"outerCapture":  8,
		"chained":       108,
		"blockCapture":  18,
		"returnsLambda": 12,
	} {
		s.Run(rule, func() {
			out, err := exec.ExecRule(ctx, "com/ex", "pol", rule, nil)
			s.Require().NoError(err)
			s.Equal(want, s.decisionNumber(out))
		})
	}
}

// A fact is a value binding, so it wins over a derive of the same name and
// calling it is an error rather than a call of the derive.
func (s *RuntimeTestSuite) TestCallFactShadowsDeriveIsNotCallable() {
	ctx := s.T().Context()
	src := `namespace com/ex

derive twice = (n: number): number => { yield n * 2 }

policy pol {
  fact twice?: number as twice default 1
  rule r = { yield twice(2) }
  export decision of r
}
`
	_, err := s.execRuleUnvalidated(ctx, src, "r")
	s.Require().ErrorContains(err, `cannot call "twice": it is a number value, not a function`)
}

func (s *RuntimeTestSuite) TestCallRuntimeArityMismatchWithoutValidate() {
	ctx := s.T().Context()
	src := `namespace com/ex

policy pol {
  let f = (x) => { yield x * 2 }
  rule tooMany = { yield f(1, 2) }
  rule tooFew = { yield f() }
  export decision of tooMany
  export decision of tooFew
}
`
	_, err := s.execRuleUnvalidated(ctx, src, "tooMany")
	s.Require().ErrorContains(err, "too many arguments: want at most 1, got 2")
	_, err = s.execRuleUnvalidated(ctx, src, "tooFew")
	s.Require().ErrorContains(err, "not enough arguments: want at least 1, got 0")
}

func (s *RuntimeTestSuite) TestCallMemoizedBindingRefusedAtRuntime() {
	ctx := s.T().Context()
	src := `namespace com/ex

policy pol {
  let f = (x) => { yield x * 2 }
  rule r = { yield f(2)! }
  export decision of r
}
`
	_, err := s.execRuleUnvalidated(ctx, src, "r")
	s.Require().ErrorContains(err, `"f" is a local value and cannot be memoized`)
}

// An imported name called without its alias reports the alias, not a bare
// unresolved-import error.
func (s *RuntimeTestSuite) TestCallImportedNameWithoutAliasNamesAlias() {
	ctx := s.T().Context()
	src := `namespace com/ex

policy pol {
  use { sha256 } from @sentrie/hash as h
  rule r = { yield sha256("a") }
  export decision of r
}
`
	_, err := s.execRuleUnvalidated(ctx, src, "r")
	s.Require().Error(err)
	s.Contains(err.Error(), `"h"`)
	s.Contains(err.Error(), "h.sha256")
}

// A calling name that nothing defines is an unknown function, not an import.
func (s *RuntimeTestSuite) TestCallUnknownFunctionAtRuntime() {
	ctx := s.T().Context()
	src := `namespace com/ex

policy pol {
  rule r = { yield nope(1) }
  export decision of r
}
`
	_, err := s.execRuleUnvalidated(ctx, src, "r")
	s.Require().ErrorContains(err, `unknown function "nope"`)
	s.NotContains(err.Error(), "unable to resolve import")
}

// A derive may call a lambda defined in its own body, but a lambda handed in
// as an argument cannot be called from the derive body at runtime.
func (s *RuntimeTestSuite) TestCallDeriveBodyLambdaPurityAtRuntime() {
	ctx := s.T().Context()
	src := `namespace com/ex

derive okDerive = (n: number): number => {
  let inc = (x) => { yield x + 1 }
  yield inc(n)
}
derive viaParam = (fn, n) => { yield fn(n) }

policy pol {
  fact k?: number as k default 5
  let outside = (x) => { yield x + k }
  rule ok = { yield okDerive(1) }
  rule leaks = { yield viaParam(outside, 1) }
  export decision of ok
  export decision of leaks
}
`
	_, err := s.execRuleUnvalidated(ctx, src, "leaks")
	s.Require().ErrorContains(err, "not permitted inside a derive")
	out, err := s.execRuleUnvalidated(ctx, src, "ok")
	s.Require().NoError(err)
	s.Equal(float64(2), s.decisionNumber(out))
}

// Recursion through a policy-level let lambda resolves the name at call time,
// so a terminating recursion evaluates.
func (s *RuntimeTestSuite) TestCallRecursiveLetLambdaTerminating() {
	ctx := s.T().Context()
	src := `namespace com/ex

policy pol {
  let fac = (n) => { yield n <= 1 ? 1 : n * fac(n - 1) }
  rule bounded = { yield fac(5) }
  export decision of bounded
}
`
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "rec.sentrie", src: src})
	out, err := exec.ExecRule(ctx, "com/ex", "pol", "bounded", nil)
	s.Require().NoError(err)
	s.Equal(float64(120), s.decisionNumber(out))
}
