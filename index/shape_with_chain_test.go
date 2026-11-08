// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"slices"
	"strings"
)

func shapeFieldNames(sh *Shape) []string {
	names := make([]string, 0, len(sh.Model.Fields))
	for name := range sh.Model.Fields {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

// After the index is validated every shape in a `with` chain holds the fields
// of all its ancestors, whatever order the shapes were declared in.
func (s *IndexTestSuite) TestShapeWithChainHydratedFieldSets() {
	decls := map[string]string{
		"A": "shape A { fa: number }",
		"B": "shape B with A { fb: number }",
		"C": "shape C with B { fc: number }",
		"D": "shape D with C { fd: number }",
	}
	orders := map[string][]string{
		"in order": {"A", "B", "C", "D"},
		"reversed": {"D", "C", "B", "A"},
		"shuffled": {"C", "A", "D", "B"},
	}
	want := map[string][]string{
		"A": {"fa"},
		"B": {"fa", "fb"},
		"C": {"fa", "fb", "fc"},
		"D": {"fa", "fb", "fc", "fd"},
	}
	for name, order := range orders {
		s.Run(name, func() {
			src := "namespace com/ex\n"
			for _, n := range order {
				src += decls[n] + "\n"
			}
			idx, err := s.validateIsSources(src)
			s.Require().NoError(err)
			for shape, fields := range want {
				s.Equal(fields, shapeFieldNames(idx.Namespaces["com/ex"].Shapes[shape]), shape)
			}
		})
	}
}

// A policy-local shape composed with a namespace-level chain gets every field
// of the chain.
func (s *IndexTestSuite) TestShapeWithChainPolicyLocalShape() {
	idx, err := s.validateIsSources(`namespace com/ex
shape C with B { fc: number }
shape B with A { fb: number }
shape A { fa: number }
policy pol {
  shape L with C { fl: number }
  rule r = { yield true }
  export decision of r
}
`)
	s.Require().NoError(err)
	local := idx.Namespaces["com/ex"].Policies["pol"].Shapes["L"]
	s.Equal([]string{"fa", "fb", "fc", "fl"}, shapeFieldNames(local))
}

// A field redeclared by a shape and an ancestor more than one level up is a
// duplicate, as it is one level up (the "parent" case).
func (s *IndexTestSuite) TestShapeWithChainDuplicateFieldFromGrandparent() {
	cases := map[string]string{
		"parent":      "shape A { x: number }\nshape B with A { x: number }\n",
		"grandparent": "shape A { x: number }\nshape B with A { y: number }\nshape C with B { x: number }\n",
		"reversed":    "shape C with B { x: number }\nshape B with A { y: number }\nshape A { x: number }\n",
		"depth four":  "shape D with C { x: number }\nshape C with B { z: number }\nshape B with A { y: number }\nshape A { x: number }\n",
	}
	for name, decls := range cases {
		s.Run(name, func() {
			_, err := s.validateIsSources("namespace com/ex\n" + decls)
			s.Require().Error(err)
			s.True(strings.Contains(err.Error(), "duplicate shape field 'x'"), err.Error())
		})
	}
}

// The static kind of a field reached through a `with` chain is known, so the
// builtin checks see it. This guards the behaviour; it does not depend on the
// hydration order.
func (s *IndexTestSuite) TestShapeWithChainInheritedFieldKindIsChecked() {
	shapes := `namespace com/ex
shape C with B { z: number }
shape B with A { y: number }
shape A { xs: list[number]  n: number }
`
	policy := func(rule string) string {
		return shapes + "policy p {\n  fact v: C as v\n  rule r = { yield " + rule + " }\n  export decision of r\n}\n"
	}

	s.Require().NoError(s.parseAndValidate(policy("count(v.xs) > 0"), "ok.sentrie"))

	err := s.parseAndValidate(policy("filter(v.n, (x) => { yield x })"), "bad.sentrie")
	s.Require().Error(err)
	s.Contains(err.Error(), "first argument must be a list")
}
