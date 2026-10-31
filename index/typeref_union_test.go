// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import "github.com/sentrie-sh/sentrie/ast"

// TestValidateUnionMembers checks that a shape reference inside a union is
// resolved like any other: unknown members are rejected, known ones pass.
func (s *IndexTestSuite) TestValidateUnionMembers() {
	cases := map[string]struct {
		body string
		msg  string // empty means validation must pass
	}{
		"is with an unknown shape member":         {body: `rule r = { yield 1 is number | Usr }`, msg: `shape "Usr" is not visible here`},
		"is with an unknown member inside a list": {body: `rule r = { yield [1] is list[number | Usr] }`, msg: `shape "Usr" is not visible here`},
		"is with an unknown nullable member":      {body: `rule r = { yield 1 is number | Usr? }`, msg: `shape "Usr" is not visible here`},
		"is with an unexported qualified member":  {body: `rule r = { yield 1 is number | com/lib/Secret }`, msg: "com/lib/Secret"},
		"is with known members":                   {body: `rule r = { yield 1 is number | U | com/lib/Account | any }`},
		"is not with known members":               {body: `rule r = { yield 1 is not string | U }`},
		"is with union of any and list of unions": {body: `rule r = { yield [1] is any | list[string | number] }`},
	}
	for name, tc := range cases {
		s.Run(name, func() {
			src := "namespace com/app\nshape U { name: string }\npolicy pol {\n  " + tc.body + "\n  export decision of r\n}\n"
			_, err := s.validateIsSources(isLibSource, src)
			if tc.msg == "" {
				s.Require().NoError(err)
				return
			}
			s.Require().Error(err)
			s.Contains(err.Error(), tc.msg)
		})
	}
}

// TestTypeRefKindUnionAndAnyAreUnknown: a union or `any` admits more than one
// runtime kind, so the static kind check must not commit to one.
func (s *IndexTestSuite) TestTypeRefKindUnionAndAnyAreUnknown() {
	rng := testBuiltinKindRange()
	idx := CreateIndex()
	for _, ref := range []ast.TypeRef{
		ast.NewAnyTypeRef(rng),
		ast.NewUnionTypeRef([]ast.TypeRef{ast.NewStringTypeRef(rng), ast.NewNumberTypeRef(rng)}, rng),
		ast.NewUnionTypeRef([]ast.TypeRef{ast.NewNumberTypeRef(rng), ast.NewNumberTypeRef(rng)}, rng),
	} {
		_, ok := typeRefKind(idx, nil, ref)
		s.False(ok, ref.String())
	}
}

// TestValidateUnionTypedValueInBuiltinCall: a fact typed as a union or `any`
// is accepted as a builtin argument whichever member the builtin wants.
func (s *IndexTestSuite) TestValidateUnionTypedValueInBuiltinCall() {
	src := `namespace com/app
policy pol {
  fact f?: string | list[number]
  fact g?: any
  rule r = { yield count(f) > 0 and count(g) > 0 }
  export decision of r
}
`
	_, err := s.validateIsSources(src)
	s.Require().NoError(err)
}
