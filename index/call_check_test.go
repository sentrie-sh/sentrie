// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"errors"

	"github.com/sentrie-sh/sentrie/xerr"
)

// Static checks of a call's callee (#161).

// callPolicy wraps body lines in a namespace and policy with one exported rule.
func callPolicy(body string) string {
	return `namespace com/ex
derive twice = (n: number): number => { yield n * 2 }
policy p {
  fact k?: number as k default 1
  let f = (x) => { yield x * 2 }
  let g = (x: number, y?: number) => { yield x }
  let v = 5
` + body + `
  export decision of allow
}
`
}

func (s *IndexTestSuite) TestCallCheckAcceptsLetBoundCallables() {
	err := s.parseAndValidate(callPolicy(`
  let count = (a: number, b: number) => { yield a + b }
  rule allow = {
    let h = (x) => { yield f(x) + g(1) + g(1, 2) + twice(x) + count(1, 2) }
    let apply = (fn, z) => { yield fn(z) }
    yield h(1) + apply(f, 2) > 0
  }`), "f.sentrie")
	s.Require().NoError(err)
}

func (s *IndexTestSuite) TestCallCheckRejects() {
	cases := map[string]struct {
		body string
		want string
	}{
		"unknown function": {
			`  rule allow = { yield nope(1) }`,
			`unknown function "nope"`,
		},
		"non-callable let": {
			`  rule allow = { yield v(1) }`,
			`cannot call "v": it is a number value, not a function`,
		},
		"non-callable block let": {
			`  rule allow = { let s = "x" yield s() }`,
			`cannot call "s": it is a string value, not a function`,
		},
		"fact": {
			`  rule allow = { yield k(1) }`,
			`cannot call "k": it is a number value, not a function`,
		},
		"rule": {
			`  rule other = { yield true }
  rule allow = { yield other() }`,
			`cannot call rule "other"`,
		},
		"too many arguments": {
			`  rule allow = { yield f(1, 2) }`,
			`"f" takes at most 1 argument(s), got 2`,
		},
		"too few arguments": {
			`  rule allow = { yield g() }`,
			`"g" takes at least 1 argument(s), got 0`,
		},
		"memoized": {
			`  rule allow = { yield f(1)! }`,
			`"f" is a local value and cannot be memoized`,
		},
	}
	for name, c := range cases {
		s.Run(name, func() {
			err := s.parseAndValidate(callPolicy(c.body), "f.sentrie")
			s.Require().Error(err)
			s.Contains(err.Error(), c.want)
			var target xerr.CallTargetError
			s.True(errors.As(err, &target), err.Error())
			s.True(errors.Is(err, xerr.ErrIndex))
			s.NotContains(err.Error(), "import")
		})
	}
}

// The import hint is kept for a name the policy imports from a module, which
// has to be called through its alias.
func (s *IndexTestSuite) TestCallCheckImportedNameCalledBare() {
	err := s.parseAndValidate(`namespace com/ex
policy p {
  use { trim } from @sentrie/string as str
  rule allow = { yield trim(" a ") == "a" }
  export decision of allow
}
`, "f.sentrie")
	s.Require().Error(err)
	s.Contains(err.Error(), `"trim" is imported from module alias "str" and must be called as str.trim(...)`)
}

func (s *IndexTestSuite) TestCallCheckLetShadowsDeriveAndBuiltin() {
	// A let named like a derive or builtin is called instead of it, so its own
	// arity applies, not the derive's or builtin's.
	err := s.parseAndValidate(callPolicy(`
  rule allow = {
    let twice = (a, b) => { yield a + b }
    let now = (a) => { yield a }
    yield twice(1, 2) + now(1) > 0
  }`), "f.sentrie")
	s.Require().NoError(err)
}

func (s *IndexTestSuite) TestDerivePurityAllowsLambdaLetCall() {
	err := s.parseAndValidate(`namespace com/ex
derive scaled = (n: number): number => {
  let mul = (x) => { yield x * n }
  let addOne = (x) => { yield mul(x) + 1 }
  yield collect([1], (y) => { yield addOne(y) })[0]
}
policy p {
  rule allow = { yield scaled(2) > 0 }
  export decision of allow
}
`, "f.sentrie")
	s.Require().NoError(err)
}

func (s *IndexTestSuite) TestDerivePurityRejectsCallsThatMayLeaveTheDerive() {
	cases := map[string]struct {
		derive string
		want   string
	}{
		"derive parameter": {
			`derive d = (fn, x) => { yield fn(x) }`,
			`calling parameter "fn" is not permitted inside a derive`,
		},
		"lambda parameter": {
			`derive d = (x: number): number => {
  let apply = (fn, v) => { yield fn(v) }
  yield apply((y) => { yield y }, x)
}`,
			`calling parameter "fn" is not permitted inside a derive`,
		},
		"let aliasing a parameter": {
			`derive d = (fn, x) => {
  let h = fn
  yield h(x)
}`,
			`calling "h" is not permitted inside a derive - only a let bound directly to a lambda can be called`,
		},
		"lambda parameter shadowing a lambda let": {
			`derive d = (x: number): number => {
  let h = (v) => { yield v }
  yield collect([x], (h) => { yield h(1) })[0]
}`,
			`calling parameter "h" is not permitted inside a derive`,
		},
	}
	for name, c := range cases {
		s.Run(name, func() {
			err := s.parseAndValidate(`namespace com/ex
`+c.derive+`
policy p {
  rule allow = { yield true }
  export decision of allow
}
`, "f.sentrie")
			s.Require().Error(err)
			s.Contains(err.Error(), c.want)
		})
	}
}

// The runtime injects all of a block's lets before evaluating any, so an
// initializer may call the let itself or a later one. Validate agrees and
// leaves unbounded self-recursion to the runtime call depth limit.
func (s *IndexTestSuite) TestCallCheckBlockLetsSeeEachOther() {
	err := s.parseAndValidate(callPolicy(`
  rule allow = {
    let factorial = (n) => { yield n <= 1 ? 1 : n * factorial(n - 1) }
    let early = (n) => { yield later(n) }
    let later = (n) => { yield n + 1 }
    yield factorial(3) + early(1) > 0
  }`), "f.sentrie")
	s.Require().NoError(err)
}

// A lambda body is not evaluated when its let or rule is, so the reference
// cycle check does not walk into it: recursive let lambdas validate, and a
// body let does not alias a rule of the same name (#161).
func (s *IndexTestSuite) TestReferenceCycleCheckSkipsLambdaBodies() {
	err := s.parseAndValidate(`namespace com/ex
policy p {
  let fac = (n) => { let v = n <= 1 ? 1 : n * fac(n - 1)  yield v }
  let isEven = (n) => { yield n == 0 ? true : isOdd(n - 1) }
  let isOdd = (n) => { yield n == 0 ? false : isEven(n - 1) }
  let g = (n) => { let allow = n  yield allow }
  rule allow = { yield fac(3) > 0 and isEven(2) and g(1) > 0 }
  export decision of allow
}
`, "f.sentrie")
	s.Require().NoError(err)

	// A let whose value, not a lambda body, refers back is still a cycle.
	err = s.parseAndValidate(`namespace com/ex
policy p {
  let x = y + 1
  let y = x + 1
  rule allow = { yield x > 0 }
  export decision of allow
}
`, "f.sentrie")
	s.Require().ErrorContains(err, "infinite recursion")
}
