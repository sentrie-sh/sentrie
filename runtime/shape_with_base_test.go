// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import "github.com/sentrie-sh/sentrie/trinary"

// shapeWithBaseLib exports com/base/A { fa: number }.
const shapeWithBaseLib = `namespace com/base
shape A { fa: number }
export shape A
`

// runWithBaseRules builds the programs and runs each rule of com/ex's policy
// p against the given facts, returning each rule's decision state.
func (s *RuntimeTestSuite) runWithBaseRules(src string, facts map[string]any, rules ...string) map[string]trinary.Value {
	ctx := s.T().Context()
	_, exec := s.mustBuildDeriveExecutor(ctx,
		deriveTestProgram{name: "base.sentrie", src: shapeWithBaseLib},
		deriveTestProgram{name: "ex.sentrie", src: src},
	)
	out := make(map[string]trinary.Value, len(rules))
	for _, rule := range rules {
		res, err := exec.ExecRule(ctx, "com/ex", "p", rule, facts)
		s.Require().NoError(err, rule)
		out[rule] = res.Decision.State
	}
	return out
}

// TestShapeWithBaseResolutionEvaluates307 checks that a policy shape composed
// with a sibling, a namespace shape, a qualified shape and a qualified
// own-namespace shape validates and checks values with the base's fields.
func (s *RuntimeTestSuite) TestShapeWithBaseResolutionEvaluates307() {
	src := `namespace com/ex
shape NsBase { fn: number }
policy p {
  fact v: document
  shape Sib { fs: number }
  shape OnSibling with Sib { f1: number }
  shape OnNamespace with NsBase { f2: number }
  shape OnQualified with com/base/A { f3: number }
  shape OnOwn with com/ex/NsBase { f4: number }
  rule sibling = { yield v is OnSibling }
  rule inns = { yield v is OnNamespace }
  rule qualified = { yield v is OnQualified }
  rule own = { yield v is OnOwn }
  export decision of sibling
  export decision of inns
  export decision of qualified
  export decision of own
}
`
	cases := []struct {
		rule string
		good map[string]any
		bad  map[string]any // the shape's own field only, the base's field missing
	}{
		{"sibling", map[string]any{"fs": 1, "f1": 1}, map[string]any{"f1": 1}},
		{"inns", map[string]any{"fn": 1, "f2": 1}, map[string]any{"f2": 1}},
		{"qualified", map[string]any{"fa": 1, "f3": 1}, map[string]any{"f3": 1}},
		{"own", map[string]any{"fn": 1, "f4": 1}, map[string]any{"f4": 1}},
	}
	for _, tc := range cases {
		got := s.runWithBaseRules(src, map[string]any{"v": tc.good}, tc.rule)
		s.Equal(trinary.True, got[tc.rule], tc.rule)

		got = s.runWithBaseRules(src, map[string]any{"v": tc.bad}, tc.rule)
		s.Equal(trinary.False, got[tc.rule], tc.rule+" without the base field")
	}
}

// TestShapeWithQualifiedBaseIgnoresLocalShape308 checks that `B with
// com/base/A` takes its fields from com/base/A, not from the namespace's own
// A, and that a policy shape named A shadows only unqualified references.
func (s *RuntimeTestSuite) TestShapeWithQualifiedBaseIgnoresLocalShape308() {
	src := `namespace com/ex
shape A { loc: number }
shape B with com/base/A { fb: number }
policy p {
  fact v: document
  shape A { pol: number }
  shape C with A { fc: number }
  shape D with com/base/A { fd: number }
  rule b = { yield v is B }
  rule c = { yield v is C }
  rule d = { yield v is D }
  export decision of b
  export decision of c
  export decision of d
}
`
	// fa and fb: the base's field and B's own. loc is the local A's field.
	got := s.runWithBaseRules(src, map[string]any{"v": map[string]any{"fa": 1, "fb": 1}}, "b", "d")
	s.Equal(trinary.True, got["b"], "B has com/base/A's fa")
	s.Equal(trinary.False, got["d"], "D lacks fd")

	got = s.runWithBaseRules(src, map[string]any{"v": map[string]any{"loc": 1, "fb": 1}}, "b")
	s.Equal(trinary.False, got["b"], "B does not take the local A's loc")

	got = s.runWithBaseRules(src, map[string]any{"v": map[string]any{"fa": 1, "fd": 1}}, "d")
	s.Equal(trinary.True, got["d"], "policy shape D with com/base/A")

	// C's unqualified A is the policy's A (pol), not the namespace's or the base's.
	got = s.runWithBaseRules(src, map[string]any{"v": map[string]any{"pol": 1, "fc": 1}}, "c")
	s.Equal(trinary.True, got["c"], "unqualified A in the policy is the policy's A")
	got = s.runWithBaseRules(src, map[string]any{"v": map[string]any{"loc": 1, "fc": 1}}, "c")
	s.Equal(trinary.False, got["c"], "policy A shadows the namespace A")
}
