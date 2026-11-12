// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"fmt"
)

// TestRejectNotATypeNameInEveryTypePosition: `bool`, `int` and `float` are
// parse errors in every type position, with a hint at the type to use (#144).
func (s *ParserTestSuite) TestRejectNotATypeNameInEveryTypePosition() {
	hints := map[string]string{"bool": "boolean", "int": "number", "float": "number"}
	positions := map[string]string{
		"fact":             "namespace n\npolicy p {\n  fact f: %s\n}\n",
		"let":              "namespace n\npolicy p {\n  let x: %s = 1\n}\n",
		"rule":             "namespace n\npolicy p {\n  rule r: %s = true\n}\n",
		"shape field":      "namespace n\nshape S { a: %s }\n",
		"shape alias":      "namespace n\nshape S list[%s]\n",
		"shape with":       "namespace n\nshape S with %s { a: string }\n",
		"lambda param":     "namespace n\npolicy p {\n  let f = (a: %s) => { yield a }\n}\n",
		"lambda return":    "namespace n\npolicy p {\n  let f = (a: number): %s => { yield a }\n}\n",
		"derive param":     "namespace n\nderive d = (a: %s) => { yield a }\n",
		"derive return":    "namespace n\nderive d = (a: number): %s => { yield a }\n",
		"is":               "namespace n\npolicy p {\n  let x = 1 is %s?\n}\n",
		"is union":         "namespace n\npolicy p {\n  let x = 1 is string | %s\n}\n",
		"cast":             "namespace n\npolicy p {\n  let x = 1 as %s\n}\n",
		"list element":     "namespace n\npolicy p {\n  let x: list[%s] = []\n}\n",
		"dict element":     "namespace n\npolicy p {\n  let x: dict[%s] = {}\n}\n",
		"nullable":         "namespace n\npolicy p {\n  fact f?: %s?\n}\n",
		"union member":     "namespace n\npolicy p {\n  fact f: string | %s\n}\n",
		"nested container": "namespace n\npolicy p {\n  fact f: list[dict[%s]]\n}\n",
	}
	for pos, tmpl := range positions {
		for name, hint := range hints {
			s.Run(pos+"/"+name, func() {
				src := fmt.Sprintf(tmpl, name)
				_, err := NewParserFromString(src, "nottype.sentrie").ParseProgram(s.T().Context())
				s.Require().Error(err)
				s.Contains(err.Error(), fmt.Sprintf("unknown type %q", name))
				s.Contains(err.Error(), fmt.Sprintf("did you mean %q", hint))
			})
		}
	}
}

// TestNotATypeNamesStayIdentifiers: they are not keywords, so they remain
// usable as lets, fields, parameters and values (#144).
func (s *ParserTestSuite) TestNotATypeNamesStayIdentifiers() {
	src := `namespace n
shape S { bool: boolean  int: number  float: number }
derive d = (int: number, float: number): number => { yield int + float }
policy p {
  fact s: S
  let bool = s.bool
  let int = 5
  let float = 1.5
  let ok = int is defined and s.int == int
  let g = (bool) => { yield bool }
  let t: boolean = bool
  rule r = { yield float > int }
  export decision of r
}
`
	prog, err := NewParserFromString(src, "idents.sentrie").ParseProgram(s.T().Context())
	s.Require().NoError(err)
	s.Require().NotNil(prog)
}

// TestRejectShapeNamedNotAType: a shape named bool, int or float could never
// be referenced, so declaring one is a parse error (PROVISIONAL, #144).
func (s *ParserTestSuite) TestRejectShapeNamedNotAType() {
	for name, hint := range map[string]string{"bool": "boolean", "int": "number", "float": "number"} {
		s.Run(name, func() {
			src := fmt.Sprintf("namespace n\nshape %s { a: string }\n", name)
			_, err := NewParserFromString(src, "shapename.sentrie").ParseProgram(s.T().Context())
			s.Require().Error(err)
			s.Contains(err.Error(), fmt.Sprintf("cannot declare a shape named %q", name))
			s.Contains(err.Error(), fmt.Sprintf("did you mean %q", hint))
		})
	}
	// a name that merely contains one is fine
	_, err := NewParserFromString("namespace n\nshape Integer { a: string }\nshape boolish string\n", "ok.sentrie").ParseProgram(s.T().Context())
	s.Require().NoError(err)
}
