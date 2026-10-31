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
