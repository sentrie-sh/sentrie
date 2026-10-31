// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/trinary"
)

// deriveShapeScopeLib exports, from t/d, one derive per type-check site in a
// derive (parameter type, return type, cast, typed let, and a parameter type
// of a lambda written in the body), each naming t/d's own shape Foo.
const deriveShapeScopeLib = `namespace t/d
shape Foo { a: string }
derive viaParam = (v: Foo): string => { yield v.a }
derive viaReturn = (v: document): Foo => { yield v }
derive viaCast = (v: document): string => { let c = v as Foo  yield c.a }
derive viaLet = (v: document): string => { let f: Foo = v  yield f.a }
derive viaLambda = (xs: list[document]): number => { yield count(filter(xs, (x: Foo): trinary => { yield true })) }
export derive viaParam
export derive viaReturn
export derive viaCast
export derive viaLet
export derive viaLambda
`

// TestDeriveTypeChecksUseDeriveScope checks that every type check inside a
// derive resolves shapes in the derive's namespace, as the index does, when
// the derive is called from another namespace: with no shape of that name in
// the caller, and with a different one that must not be picked up.
func (s *RuntimeTestSuite) TestDeriveTypeChecksUseDeriveScope() {
	ctx := s.T().Context()
	sites := []string{"viaParam", "viaReturn", "viaCast", "viaLet", "viaLambda"}
	for name, callerShapes := range map[string]string{
		"caller without the shape":              "",
		"caller with its own shape of the name": "shape Foo { b: number }\n",
	} {
		s.Run(name, func() {
			caller := "namespace t/e\n" + callerShapes + `policy p {
  let dFoo = { "a": "x" }
  let eFoo = { "b": 1 }
  rule viaParam = { yield t/d/viaParam(dFoo) }
  rule viaReturn = { yield t/d/viaReturn(dFoo) }
  rule viaCast = { yield t/d/viaCast(dFoo) }
  rule viaLet = { yield t/d/viaLet(dFoo) }
  rule viaLambda = { yield t/d/viaLambda([dFoo]) }
  rule viaParamBad = { yield t/d/viaParam(eFoo) }
  rule viaReturnBad = { yield t/d/viaReturn(eFoo) }
  rule viaCastBad = { yield t/d/viaCast(eFoo) }
  rule viaLetBad = { yield t/d/viaLet(eFoo) }
  rule viaLambdaBad = { yield t/d/viaLambda([eFoo]) }
  export decision of viaParam
  export decision of viaReturn
  export decision of viaCast
  export decision of viaLet
  export decision of viaLambda
  export decision of viaParamBad
  export decision of viaReturnBad
  export decision of viaCastBad
  export decision of viaLetBad
  export decision of viaLambdaBad
}
`
			_, exec := s.mustBuildDeriveExecutor(ctx,
				deriveTestProgram{name: "d.sentrie", src: deriveShapeScopeLib},
				deriveTestProgram{name: "e.sentrie", src: caller},
			)
			for _, site := range sites {
				out, err := exec.ExecRule(ctx, "t/e", "p", site, nil)
				s.Require().NoError(err, site)
				s.Equal(trinary.True, out.Decision.State, site)

				// A value that is only the caller's Foo fails t/d's Foo.
				_, err = exec.ExecRule(ctx, "t/e", "p", site+"Bad", nil)
				s.Require().Error(err, site)
				s.Contains(err.Error(), "field a is required", site)
			}
		})
	}
}

// TestPolicyLambdaThroughDeriveUsesPolicyScope checks that a lambda written in
// a policy keeps resolving its shapes in that policy when a derive in another
// namespace, which declares a shape of the same name, applies it.
func (s *RuntimeTestSuite) TestPolicyLambdaThroughDeriveUsesPolicyScope() {
	ctx := s.T().Context()
	lib := `namespace t/d
shape Foo { a: string }
derive keep = (xs: list[document], pred): list[document] => { yield filter(xs, pred) }
export derive keep
`
	caller := `namespace t/e
shape Foo { b: number }
policy p {
  let xs = [{ "a": "x" }, { "b": 1 }]
  rule byIs = { yield t/d/keep(xs, (x: document): trinary => { yield x is Foo }) }
  rule byParam = { yield t/d/keep([{ "b": 2 }], (x: Foo): trinary => { yield true }) }
  rule byCast = { yield t/d/keep([{ "b": 3 }], (x: document): trinary => { let c = x as Foo  yield true }) }
  rule byLet = { yield t/d/keep([{ "b": 4 }], (x: document): trinary => { let f: Foo = x  yield f.b > 0 }) }
  export decision of byIs
  export decision of byParam
  export decision of byCast
  export decision of byLet
}
`
	_, exec := s.mustBuildDeriveExecutor(ctx,
		deriveTestProgram{name: "d.sentrie", src: lib},
		deriveTestProgram{name: "e.sentrie", src: caller},
	)
	want := map[string]float64{"byIs": 1, "byParam": 2, "byCast": 3, "byLet": 4}
	for rule, b := range want {
		out, err := exec.ExecRule(ctx, "t/e", "p", rule, nil)
		s.Require().NoError(err, rule)
		kept, ok := out.Decision.Value.ListValue()
		s.Require().True(ok, rule)
		s.Require().Len(kept, 1, rule)
		doc, ok := kept[0].DictValue()
		s.Require().True(ok, rule)
		s.Equal(box.Number(b), doc["b"], rule)
	}
}

// deriveShapeScopeRun builds the programs and runs each rule of t/e's policy
// p, returning the decision state or the error per rule.
type deriveShapeScopeRun struct {
	state trinary.Value
	value box.Value
	err   error
}

func (s *RuntimeTestSuite) runDeriveShapeScopeRules(programs []deriveTestProgram, rules ...string) map[string]deriveShapeScopeRun {
	ctx := s.T().Context()
	_, exec := s.mustBuildDeriveExecutor(ctx, programs...)
	out := make(map[string]deriveShapeScopeRun, len(rules))
	for _, rule := range rules {
		res, err := exec.ExecRule(ctx, "t/e", "p", rule, nil)
		run := deriveShapeScopeRun{err: err}
		if err == nil {
			run.state = res.Decision.State
			run.value = res.Decision.Value
		}
		out[rule] = run
	}
	return out
}

// deriveShapeScopeCaller builds t/e's source: its own shapes, then policy p
// with one exported rule per "name = expr" pair.
func deriveShapeScopeCaller(callerShapes string, rules ...[2]string) string {
	src := "namespace t/e\n" + callerShapes + "policy p {\n"
	for _, r := range rules {
		src += "  rule " + r[0] + " = { yield " + r[1] + " }\n  export decision of " + r[0] + "\n"
	}
	return src + "}\n"
}

// TestDeriveNestedAndContainerShapesUseDeriveScope checks that a shape named
// only inside another shape's field, or as the element type of a list
// or record, resolves in the derive's namespace: the caller has no such shape
// or a different one of the same name.
func (s *RuntimeTestSuite) TestDeriveNestedAndContainerShapesUseDeriveScope() {
	lib := `namespace t/d
shape Inner { n: number }
shape Outer { i: Inner }
derive viaNested = (v: Outer): number => { yield v.i.n }
derive viaList = (v: list[Inner]): number => { yield count(v) }
derive viaRecord = (v: record[string, Inner]): number => { yield 1 }
derive viaCastList = (v: list[document]): number => { let l = v as list[Inner]  yield count(l) }
export derive viaNested
export derive viaList
export derive viaRecord
export derive viaCastList
`
	for name, callerShapes := range map[string]string{
		"caller without the shapes":     "",
		"caller with its own Inner":     "shape Inner { m: string }\n",
		"caller with its own Outer too": "shape Inner { m: string }\nshape Outer { i: Inner }\n",
	} {
		s.Run(name, func() {
			caller := deriveShapeScopeCaller(callerShapes,
				[2]string{"nested", `t/d/viaNested({ "i": { "n": 1 } })`},
				[2]string{"inList", `t/d/viaList([{ "n": 1 }, { "n": 2 }])`},
				[2]string{"inRecord", `t/d/viaRecord(["k", { "n": 1 }])`},
				[2]string{"castList", `t/d/viaCastList([{ "n": 1 }, { "n": 2 }])`},
				[2]string{"nestedBad", `t/d/viaNested({ "i": { "m": "x" } })`},
				[2]string{"listBad", `t/d/viaList([{ "m": "x" }])`},
				[2]string{"recordBad", `t/d/viaRecord(["k", { "m": "x" }])`},
				[2]string{"castListBad", `t/d/viaCastList([{ "m": "x" }])`},
			)
			got := s.runDeriveShapeScopeRules([]deriveTestProgram{
				{name: "d.sentrie", src: lib},
				{name: "e.sentrie", src: caller},
			}, "nested", "inList", "inRecord", "castList", "nestedBad", "listBad", "recordBad", "castListBad")
			for _, rule := range []string{"nested", "inList", "inRecord", "castList"} {
				s.Require().NoError(got[rule].err, rule)
			}
			for _, rule := range []string{"nestedBad", "listBad", "recordBad", "castListBad"} {
				s.Require().Error(got[rule].err, rule)
				s.Contains(got[rule].err.Error(), "field n is required", rule)
			}
		})
	}
}

// TestDeriveCallingDeriveChecksEachInItsOwnScope checks a derive in t/d that
// calls a derive in t/f, called from t/e, when all three declare a different
// Foo: each derive checks its parameter against its own namespace's Foo.
func (s *RuntimeTestSuite) TestDeriveCallingDeriveChecksEachInItsOwnScope() {
	fLib := `namespace t/f
shape Foo { a: string }
derive inner = (v: Foo): string => { yield v.a }
export derive inner
`
	dLib := `namespace t/d
shape Foo { b: number }
derive outer = (v: Foo): string => { yield t/f/inner({ "a": "x" }) }
derive outerBad = (v: Foo): string => { yield t/f/inner({ "b": 1 }) }
export derive outer
export derive outerBad
`
	caller := deriveShapeScopeCaller("shape Foo { c: boolean }\n",
		[2]string{"ok", `t/d/outer({ "b": 1 })`},
		[2]string{"outerParamBad", `t/d/outer({ "c": true })`},
		[2]string{"innerParamBad", `t/d/outerBad({ "b": 1 })`},
	)
	got := s.runDeriveShapeScopeRules([]deriveTestProgram{
		{name: "f.sentrie", src: fLib},
		{name: "d.sentrie", src: dLib},
		{name: "e.sentrie", src: caller},
	}, "ok", "outerParamBad", "innerParamBad")

	s.Require().NoError(got["ok"].err)
	s.Equal(box.String("x"), got["ok"].value)

	// {"c": true} is only the caller's Foo: t/d's Foo wants b.
	s.Require().Error(got["outerParamBad"].err)
	s.Contains(got["outerParamBad"].err.Error(), "field b is required")

	// {"b": 1} passes t/d's Foo but is not t/f's Foo, which inner checks.
	s.Require().Error(got["innerParamBad"].err)
	s.Contains(got["innerParamBad"].err.Error(), "field a is required")
}

// TestMemoizedDeriveCallUsesDeriveScopeForEveryCaller checks that a memoized
// call of a derive whose body casts to its namespace's Foo gives the same
// answer from two policies that declare different Foos, and on repeats, so a
// result cached under one caller's shapes is never wrong for another.
func (s *RuntimeTestSuite) TestMemoizedDeriveCallUsesDeriveScopeForEveryCaller() {
	ctx := s.T().Context()
	lib := `namespace t/d
shape Foo { a: string }
derive viaCast = (v: document): string => { let c = v as Foo  yield c.a }
export derive viaCast
`
	callerE := `namespace t/e
shape Foo { b: number }
policy p {
  let dFoo = { "a": "x" }
  let eFoo = { "b": 1 }
  rule ok = { yield t/d/viaCast(dFoo)!60 }
  rule bad = { yield t/d/viaCast(eFoo)!60 }
  export decision of ok
  export decision of bad
}
`
	callerG := `namespace t/g
policy p {
  let dFoo = { "a": "x" }
  let eFoo = { "b": 1 }
  rule ok = { yield t/d/viaCast(dFoo)!60 }
  rule bad = { yield t/d/viaCast(eFoo)!60 }
  export decision of ok
  export decision of bad
}
`
	_, exec := s.mustBuildDeriveExecutor(ctx,
		deriveTestProgram{name: "d.sentrie", src: lib},
		deriveTestProgram{name: "e.sentrie", src: callerE},
		deriveTestProgram{name: "g.sentrie", src: callerG},
	)
	for round := range 2 {
		for _, ns := range []string{"t/e", "t/g"} {
			out, err := exec.ExecRule(ctx, ns, "p", "ok", nil)
			s.Require().NoError(err, "%s round %d", ns, round)
			s.Equal(box.String("x"), out.Decision.Value, "%s round %d", ns, round)

			_, err = exec.ExecRule(ctx, ns, "p", "bad", nil)
			s.Require().Error(err, "%s round %d", ns, round)
			s.Contains(err.Error(), "field a is required", "%s round %d", ns, round)
		}
	}
}

// TestDeriveReachedThroughPipelineUsesDeriveScope checks that a derive applied
// with the pipeline operator inside another derive of its namespace checks its
// parameter in that namespace, whoever called the outer derive.
func (s *RuntimeTestSuite) TestDeriveReachedThroughPipelineUsesDeriveScope() {
	lib := `namespace t/d
shape Foo { a: string }
derive inner = (v: document): string => { let c = v as Foo  yield c.a }
derive viaPipe = (v: document): string => { yield v |> inner() }
export derive viaPipe
`
	for name, callerShapes := range map[string]string{
		"caller without the shape":              "",
		"caller with its own shape of the name": "shape Foo { b: number }\n",
	} {
		s.Run(name, func() {
			caller := deriveShapeScopeCaller(callerShapes,
				[2]string{"ok", `t/d/viaPipe({ "a": "x" })`},
				[2]string{"bad", `t/d/viaPipe({ "b": 1 })`},
			)
			got := s.runDeriveShapeScopeRules([]deriveTestProgram{
				{name: "d.sentrie", src: lib},
				{name: "e.sentrie", src: caller},
			}, "ok", "bad")
			s.Require().NoError(got["ok"].err)
			s.Equal(box.String("x"), got["ok"].value)
			s.Require().Error(got["bad"].err)
			s.Contains(got["bad"].err.Error(), "field a is required")
		})
	}
}

// TestDeriveMissingShapeInDeriveScopeIsAnErrorNotACallerFallback checks that
// when the derive's own namespace lacks a shape the derive names, the caller
// declaring it does not make the check pass: the error names the shape.
func (s *RuntimeTestSuite) TestDeriveMissingShapeInDeriveScopeIsAnErrorNotACallerFallback() {
	lib := `namespace t/d
derive viaParam = (v: Nope): string => { yield "x" }
derive viaCast = (v: document): document => { let c = v as Nope  yield c }
derive viaLet = (v: document): document => { let c: Nope = v  yield c }
derive viaReturn = (v: document): Nope => { yield v }
export derive viaParam
export derive viaCast
export derive viaLet
export derive viaReturn
`
	caller := deriveShapeScopeCaller("shape Nope { b: number }\n",
		[2]string{"param", `t/d/viaParam({ "b": 1 })`},
		[2]string{"cast", `t/d/viaCast({ "b": 1 })`},
		[2]string{"typedLet", `t/d/viaLet({ "b": 1 })`},
		[2]string{"return", `t/d/viaReturn({ "b": 1 })`},
	)
	got := s.runDeriveShapeScopeRules([]deriveTestProgram{
		{name: "d.sentrie", src: lib},
		{name: "e.sentrie", src: caller},
	}, "param", "cast", "typedLet", "return")
	for rule, run := range got {
		s.Require().Error(run.err, rule)
		s.Contains(run.err.Error(), "Nope", rule)
	}
}
