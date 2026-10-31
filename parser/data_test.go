// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"github.com/sentrie-sh/sentrie/ast"
)

func (s *ParserTestSuite) parseOnePolicy(src string) (*ast.PolicyStatement, error) {
	prog, err := NewParserFromString(src, "test.sentrie").ParseProgram(s.T().Context())
	if err != nil {
		return nil, err
	}
	for _, stmt := range prog.Statements {
		if p, ok := stmt.(*ast.PolicyStatement); ok {
			return p, nil
		}
	}
	s.FailNow("no policy in program")
	return nil, nil
}

func (s *ParserTestSuite) TestParseDataStatementForms() {
	cases := []struct {
		src       string
		name      string
		alias     string
		typed     bool
		shapeName string
	}{
		{src: "data roles", name: "roles", alias: "roles"},
		{src: "data roles as r", name: "roles", alias: "r"},
		{src: "data roles: dict[list[string]]", name: "roles", alias: "roles", typed: true},
		{src: "data roles: RoleTable as r", name: "roles", alias: "r", typed: true, shapeName: "RoleTable"},
	}
	for _, tc := range cases {
		s.Run(tc.src, func() {
			pol, err := s.parseOnePolicy("namespace a\npolicy p {\n" + tc.src + "\n}\n")
			s.Require().NoError(err)
			s.Require().Len(pol.Statements, 1)
			d, ok := pol.Statements[0].(*ast.DataStatement)
			s.Require().True(ok, "got %T", pol.Statements[0])
			s.Equal(tc.name, d.Name)
			s.Equal(tc.alias, d.Alias)
			s.Equal("data", d.Kind())
			if !tc.typed {
				s.Nil(d.Type)
				return
			}
			s.Require().NotNil(d.Type)
			if tc.shapeName != "" {
				st, ok := d.Type.(*ast.ShapeTypeRef)
				s.Require().True(ok, "got %T", d.Type)
				s.Equal(tc.shapeName, st.Ref.String())
			}
		})
	}
}

// `data` is contextual: outside the head of a policy statement it remains an
// ordinary identifier.
func (s *ParserTestSuite) TestParseDataStaysAnIdentifierElsewhere() {
	src := `namespace a
policy p {
  fact data: dict[string]
  data roles as data_roles
  let x = data.y
  let f = (data) => { yield data }
  rule r = data["k"] == "v"
  export decision of r
}
`
	pol, err := s.parseOnePolicy(src)
	s.Require().NoError(err)
	f, ok := pol.Statements[0].(*ast.FactStatement)
	s.Require().True(ok)
	s.Equal("data", f.Name)
	d, ok := pol.Statements[1].(*ast.DataStatement)
	s.Require().True(ok)
	s.Equal("data_roles", d.Alias)
}

func (s *ParserTestSuite) TestParseDataStatementErrors() {
	for _, src := range []string{
		"data",
		"data 3",
		"data roles:",
		"data roles as",
		"data roles as 3",
	} {
		s.Run(src, func() {
			_, err := NewParserFromString("namespace a\npolicy p {\n"+src+"\n}\n", "test.sentrie").ParseProgram(s.T().Context())
			s.Error(err)
		})
	}
}
