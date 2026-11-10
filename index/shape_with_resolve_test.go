// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

// withBaseLib is a namespace that exports shape A and keeps shape P private.
const withBaseLib = `namespace com/base
shape A { fa: number }
shape P { fp: number }
export shape A
`

const withBasePolicyTail = `
  rule r = { yield true }
  export decision of r
}
`

// A qualified base resolves only in the namespace it names, and that namespace
// must export it (#308).
func (s *IndexTestSuite) TestShapeWithQualifiedBaseRequiresExport() {
	cases := map[string]struct {
		src string
		msg string
	}{
		"exported":      {src: "namespace com/ex\nshape B with com/base/A { fb: number }\n"},
		"not exported":  {src: "namespace com/ex\nshape B with com/base/P { fb: number }\n", msg: "shape 'com/base/P' not exported"},
		"missing shape": {src: "namespace com/ex\nshape B with com/base/Nope { fb: number }\n", msg: "shape 'com/base/Nope' not found"},
		"missing ns":    {src: "namespace com/ex\nshape B with com/none/A { fb: number }\n", msg: "shape 'com/none/A' not found"},
		"policy shape exported": {
			src: "namespace com/ex\npolicy pol {\n  shape B with com/base/A { fb: number }" + withBasePolicyTail,
		},
		"policy shape not exported": {
			src: "namespace com/ex\npolicy pol {\n  shape B with com/base/P { fb: number }" + withBasePolicyTail,
			msg: "shape 'com/base/P' not exported",
		},
	}
	for name, tc := range cases {
		s.Run(name, func() {
			_, err := s.validateIsSources(withBaseLib, tc.src)
			if tc.msg == "" {
				s.Require().NoError(err)
				return
			}
			s.Require().Error(err)
			s.Contains(err.Error(), tc.msg)
		})
	}
}

// A qualified base that names the shape's own namespace needs no export.
func (s *IndexTestSuite) TestShapeWithQualifiedOwnNamespaceNeedsNoExport() {
	idx, err := s.validateIsSources(`namespace t/a
shape A { fa: number }
shape B with t/a/A { fb: number }
policy pol {
  shape L with t/a/A { fl: number }
  rule r = { yield true }
  export decision of r
}
`)
	s.Require().NoError(err)
	ns := idx.Namespaces["t/a"]
	s.Equal([]string{"fa", "fb"}, shapeFieldNames(ns.Shapes["B"]))
	s.Equal([]string{"fa", "fl"}, shapeFieldNames(ns.Policies["pol"].Shapes["L"]))

	// Another namespace naming the same unexported shape still fails.
	_, err = s.validateIsSources(`namespace t/a
shape A { fa: number }
`, `namespace t/b
shape B with t/a/A { fb: number }
`)
	s.Require().Error(err)
	s.Contains(err.Error(), "shape 't/a/A' not exported")
}

// An unqualified base resolves in the policy, then the namespace; the first
// match wins.
func (s *IndexTestSuite) TestShapeWithUnqualifiedBasePolicyThenNamespace() {
	idx, err := s.validateIsSources(`namespace com/ex
shape A { nsf: number }
shape N { nf: number }
policy pol {
  shape A { pf: number }
  shape ShadowUser with A { u: number }
  shape NsUser with N { u: number }
  rule r = { yield true }
  export decision of r
}
shape OutsideUser with A { u: number }
`)
	s.Require().NoError(err)
	ns := idx.Namespaces["com/ex"]
	pol := ns.Policies["pol"]
	s.Equal([]string{"pf", "u"}, shapeFieldNames(pol.Shapes["ShadowUser"]), "policy shape shadows the namespace shape")
	s.Equal([]string{"nf", "u"}, shapeFieldNames(pol.Shapes["NsUser"]), "namespace shape reached from a policy shape")
	s.Equal([]string{"nsf", "u"}, shapeFieldNames(ns.Shapes["OutsideUser"]), "namespace shape outside the policy")
}

// An unqualified base found only as another namespace's export must be
// qualified; one found nowhere is not found.
func (s *IndexTestSuite) TestShapeWithUnqualifiedBaseElsewhere() {
	cases := map[string]struct {
		src string
		msg string
	}{
		"only another namespace export": {
			src: "namespace com/ex\nshape B with A { fb: number }\n",
			msg: "must be qualified: use com/base/A",
		},
		"policy shape, only another namespace export": {
			src: "namespace com/ex\npolicy pol {\n  shape B with A { fb: number }" + withBasePolicyTail,
			msg: "must be qualified: use com/base/A",
		},
		"another namespace shape that is not exported": {
			src: "namespace com/ex\nshape B with P { fb: number }\n",
			msg: "shape 'P' not found",
		},
		"nowhere": {
			src: "namespace com/ex\nshape B with Nope { fb: number }\n",
			msg: "shape 'Nope' not found",
		},
	}
	for name, tc := range cases {
		s.Run(name, func() {
			_, err := s.validateIsSources(withBaseLib, tc.src)
			s.Require().Error(err)
			s.Contains(err.Error(), tc.msg)
		})
	}
}

// #307: a policy shape composes with a sibling in its policy, in the
// namespace, and by a qualified name.
func (s *IndexTestSuite) TestShapeWithPolicyShapeBases307() {
	idx, err := s.validateIsSources(withBaseLib, `namespace com/ex
shape NsBase { fn: number }
policy pol {
  shape Sib { fs: number }
  shape OnSibling with Sib { f1: number }
  shape OnNamespace with NsBase { f2: number }
  shape OnQualified with com/base/A { f3: number }
  shape OnOwnQualified with com/ex/NsBase { f4: number }
  rule r = { yield true }
  export decision of r
}
`)
	s.Require().NoError(err)
	pol := idx.Namespaces["com/ex"].Policies["pol"]
	s.Equal([]string{"f1", "fs"}, shapeFieldNames(pol.Shapes["OnSibling"]))
	s.Equal([]string{"f2", "fn"}, shapeFieldNames(pol.Shapes["OnNamespace"]))
	s.Equal([]string{"f3", "fa"}, shapeFieldNames(pol.Shapes["OnQualified"]))
	s.Equal([]string{"f4", "fn"}, shapeFieldNames(pol.Shapes["OnOwnQualified"]))
}

// #308: `B with com/base/A` takes its fields from com/base/A even when the
// namespace declares its own A, and no cycle is reported through the local A.
func (s *IndexTestSuite) TestShapeWithQualifiedBaseIgnoresLocalShape308() {
	idx, err := s.validateIsSources(withBaseLib, `namespace com/ex
shape A with B { local: number }
shape B with com/base/A { fb: number }
policy pol {
  shape A { pa: number }
  shape C with com/base/A { fc: number }
  rule r = { yield true }
  export decision of r
}
`)
	s.Require().NoError(err, "hydration and the cycle check must both pick com/base/A")
	ns := idx.Namespaces["com/ex"]
	s.Equal([]string{"fa", "fb"}, shapeFieldNames(ns.Shapes["B"]))
	s.Equal([]string{"fa", "fb", "local"}, shapeFieldNames(ns.Shapes["A"]))
	s.Equal([]string{"fa", "fc"}, shapeFieldNames(ns.Policies["pol"].Shapes["C"]))
}

// Cycle detection still catches cycles among policy shapes, including through
// a qualified own-namespace base, and an unqualified base resolves the way
// hydration does (policy first).
func (s *IndexTestSuite) TestShapeWithCycleDetection() {
	cases := map[string]string{
		"policy shapes":                   "namespace com/ex\npolicy pol {\n  shape X with Y { x: number }\n  shape Y with X { y: number }" + withBasePolicyTail,
		"namespace shapes, qualified own": "namespace com/ex\nshape X with com/ex/Y { x: number }\nshape Y with com/ex/X { y: number }\n",
		"policy shadow closes the cycle":  "namespace com/ex\nshape A { a: number }\npolicy pol {\n  shape A with B { pa: number }\n  shape B with A { pb: number }" + withBasePolicyTail,
	}
	for name, src := range cases {
		s.Run(name, func() {
			_, err := s.validateIsSources(withBaseLib, src)
			s.Require().Error(err)
			s.Contains(err.Error(), "cyclic dependencies in shapes")
		})
	}

	s.Run("policy self reference is a self-loop", func() {
		_, err := s.validateIsSources(withBaseLib, "namespace com/ex\npolicy pol {\n  shape X with X { x: number }"+withBasePolicyTail)
		s.Require().Error(err)
		s.Contains(err.Error(), "self-loop")
	})

	s.Run("policy shape with the namespace shape of its own name is no cycle", func() {
		idx, err := s.validateIsSources(withBaseLib, `namespace com/ex
shape A { a: number }
policy pol {
  shape B with A { b: number }
  shape A2 with B { c: number }`+withBasePolicyTail)
		s.Require().NoError(err)
		s.Equal([]string{"a", "b", "c"}, shapeFieldNames(idx.Namespaces["com/ex"].Policies["pol"].Shapes["A2"]))
	})
}
