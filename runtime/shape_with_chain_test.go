// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"fmt"
	"strings"

	"github.com/sentrie-sh/sentrie/trinary"
)

// shapeChainDecls are the shape declarations of the `with` chain tests. Each
// shape adds one field named after itself and composes with the previous one.
var shapeChainDecls = map[string]string{
	"A": "shape A { fa: number }",
	"B": "shape B with A { fb: number }",
	"C": "shape C with B { fc: number }",
	"D": "shape D with C { fd: number }",
}

func chainValue(names ...string) string {
	parts := make([]string, 0, len(names))
	for _, n := range names {
		parts = append(parts, fmt.Sprintf(`"f%s": 1`, strings.ToLower(n)))
	}
	return "{ " + strings.Join(parts, ", ") + " }"
}

// A value must carry the fields of every shape up a `with` chain, however deep
// the chain is and in whatever order the shapes are declared.
func (s *RuntimeTestSuite) TestShapeWithChainHydratesAncestors() {
	ctx := s.T().Context()

	cases := map[string]struct {
		order []string
		top   string
		all   []string
	}{
		"depth 3 in order":      {order: []string{"A", "B", "C"}, top: "C", all: []string{"A", "B", "C"}},
		"depth 3 reversed":      {order: []string{"C", "B", "A"}, top: "C", all: []string{"A", "B", "C"}},
		"depth 3 shuffled":      {order: []string{"B", "C", "A"}, top: "C", all: []string{"A", "B", "C"}},
		"depth 4 in order":      {order: []string{"A", "B", "C", "D"}, top: "D", all: []string{"A", "B", "C", "D"}},
		"depth 4 reversed":      {order: []string{"D", "C", "B", "A"}, top: "D", all: []string{"A", "B", "C", "D"}},
		"depth 4 shuffled":      {order: []string{"C", "A", "D", "B"}, top: "D", all: []string{"A", "B", "C", "D"}},
		"depth 4 other shuffle": {order: []string{"B", "D", "A", "C"}, top: "D", all: []string{"A", "B", "C", "D"}},
	}

	for name, tc := range cases {
		s.Run(name, func() {
			decls := make([]string, 0, len(tc.order))
			for _, n := range tc.order {
				decls = append(decls, shapeChainDecls[n])
			}
			src := "namespace com/app\n" + strings.Join(decls, "\n") + fmt.Sprintf(`
policy pol {
  let full = %s
  rule complete = { yield full is %s }
  export decision of complete
`, chainValue(tc.all...), tc.top)
			// One rule per field: dropping that field alone must fail the check.
			for _, n := range tc.all {
				without := make([]string, 0, len(tc.all)-1)
				for _, m := range tc.all {
					if m != n {
						without = append(without, m)
					}
				}
				src += fmt.Sprintf("  rule missing_%s = { yield %s is %s }\n  export decision of missing_%s\n",
					n, chainValue(without...), tc.top, n)
			}
			src += "}\n"

			_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "chain.sentrie", src: src})

			out, err := exec.ExecRule(ctx, "com/app", "pol", "complete", nil)
			s.Require().NoError(err)
			s.Equal(trinary.True, out.Decision.State)

			for _, n := range tc.all {
				out, err := exec.ExecRule(ctx, "com/app", "pol", "missing_"+n, nil)
				s.Require().NoError(err)
				s.Equal(trinary.False, out.Decision.State, "a value missing field f%s must not conform to %s", strings.ToLower(n), tc.top)
			}
		})
	}
}

// Two children of one parent, each with a child of its own: both branches must
// carry the shared grandparent's fields. A shape has a single `with` parent, so
// this is the closest a diamond gets.
func (s *RuntimeTestSuite) TestShapeWithChainSharedGrandparent() {
	ctx := s.T().Context()
	src := `namespace com/app
shape C1 with B1 { fc1: number }
shape C2 with B2 { fc2: number }
shape B2 with A { fb2: number }
shape B1 with A { fb1: number }
shape A { fa: number }

policy pol {
  rule c1_ok = { yield { "fa": 1, "fb1": 1, "fc1": 1 } is C1 }
  rule c1_no_a = { yield { "fb1": 1, "fc1": 1 } is C1 }
  rule c2_ok = { yield { "fa": 1, "fb2": 1, "fc2": 1 } is C2 }
  rule c2_no_a = { yield { "fb2": 1, "fc2": 1 } is C2 }
  rule c1_not_c2 = { yield { "fa": 1, "fb1": 1, "fc1": 1 } is C2 }
  export decision of c1_ok
  export decision of c1_no_a
  export decision of c2_ok
  export decision of c2_no_a
  export decision of c1_not_c2
}
`
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "diamond.sentrie", src: src})
	want := map[string]trinary.Value{
		"c1_ok":     trinary.True,
		"c1_no_a":   trinary.False,
		"c2_ok":     trinary.True,
		"c2_no_a":   trinary.False,
		"c1_not_c2": trinary.False,
	}
	for rule, state := range want {
		s.Run(rule, func() {
			out, err := exec.ExecRule(ctx, "com/app", "pol", rule, nil)
			s.Require().NoError(err)
			s.Equal(state, out.Decision.State)
		})
	}
}

// A chain whose links live in different namespaces, with the middle and top
// shapes declared before the exported base is seen.
func (s *RuntimeTestSuite) TestShapeWithChainAcrossNamespaces() {
	ctx := s.T().Context()
	base := `namespace com/base
shape A { fa: number }
export shape A
`
	mid := `namespace com/mid
shape B with com/base/A { fb: number }
export shape B
`
	app := `namespace com/app
shape C with com/mid/B { fc: number }

policy pol {
  rule ok = { yield { "fa": 1, "fb": 1, "fc": 1 } is C }
  rule no_a = { yield { "fb": 1, "fc": 1 } is C }
  rule no_b = { yield { "fa": 1, "fc": 1 } is C }
  export decision of ok
  export decision of no_a
  export decision of no_b
}
`
	want := map[string]trinary.Value{"ok": trinary.True, "no_a": trinary.False, "no_b": trinary.False}

	orders := map[string][]deriveTestProgram{
		"base first": {{name: "base.sentrie", src: base}, {name: "mid.sentrie", src: mid}, {name: "app.sentrie", src: app}},
		"app first":  {{name: "app.sentrie", src: app}, {name: "mid.sentrie", src: mid}, {name: "base.sentrie", src: base}},
	}
	for name, programs := range orders {
		s.Run(name, func() {
			_, exec := s.mustBuildDeriveExecutor(ctx, programs...)
			for rule, state := range want {
				out, err := exec.ExecRule(ctx, "com/app", "pol", rule, nil)
				s.Require().NoError(err)
				s.Equal(state, out.Decision.State, rule)
			}
		})
	}
}
