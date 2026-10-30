// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"strings"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/parser"
)

// validateIsSources indexes and validates the programs together.
func (s *IndexTestSuite) validateIsSources(srcs ...string) (*Index, error) {
	ctx := s.T().Context()
	idx := CreateIndex()
	for i, src := range srcs {
		prog, err := parser.NewParserFromString(src, "is"+string(rune('a'+i))+".sentrie").ParseProgram(ctx)
		s.Require().NoError(err)
		s.Require().NoError(idx.AddProgram(ctx, prog))
	}
	return idx, idx.Validate(ctx)
}

const isLibSource = `namespace com/lib
shape Account { id: string }
export shape Account
shape Secret { id: string }
`

func (s *IndexTestSuite) TestResolveIsNameErrors() {
	cases := map[string]struct {
		body string
		msg  string
	}{
		"qualified namespace without the shape": {body: `rule r = { yield 1 is com/lib/Missing }`, msg: "com/lib/Missing"},
		"qualified shape not exported":          {body: `rule r = { yield 1 is com/lib/Secret }`, msg: "com/lib/Secret"},
		"misspelt shape in a list type":         {body: `rule r = { yield [1] is list[Usr] }`, msg: `shape "Usr" is not visible here`},
		"misspelt shape with nullable suffix":   {body: `rule r = { yield 1 is Usr? }`, msg: `shape "Usr" is not visible here`},
		"block let is both":                     {body: `rule r = { let U = 1  yield 1 is U }`, msg: `"U" names both a shape and a value in scope`},
		"lambda parameter is both":              {body: `rule r = { yield any([1], (U) => { yield 1 is U }) }`, msg: `"U" names both a shape and a value in scope`},
		"rule is both":                          {body: "rule U = { yield true }\n  rule r = { yield 1 is U }", msg: `"U" names both a shape and a value in scope`},
	}
	for name, tc := range cases {
		s.Run(name, func() {
			src := "namespace com/app\nshape U { name: string }\npolicy pol {\n  " + tc.body + "\n  export decision of r\n}\n"
			_, err := s.validateIsSources(isLibSource, src)
			s.Require().Error(err)
			s.Contains(err.Error(), tc.msg)
		})
	}
}

func (s *IndexTestSuite) TestResolveIsLocalValuesStayEquality() {
	src := `namespace com/app
policy pol {
  rule block_let = { let v = 1  yield 1 is v }
  rule lambda_param = { yield any([1], (v) => { yield 1 is v }) }
  export decision of block_let
  export decision of lambda_param
}
`
	idx, err := s.validateIsSources(src)
	s.Require().NoError(err)
	var seen int
	s.Require().NoError(idx.forEachScopedExpr(s.T().Context(), func(_ *kindCheckCtx, e ast.Expression) {
		if is, ok := e.(*ast.IsExpression); ok {
			seen++
			s.Nil(is.Type)
			s.NotNil(is.Right)
		}
	}))
	s.Equal(2, seen)
}

func (s *IndexTestSuite) TestResolveIsFactInsideDerive() {
	src := `namespace com/app
policy pol {
  fact f?: document
  let seed = 0
  derive d = (v: document): trinary => { yield v is f }
  rule r = { yield d({}) }
  export decision of r
}
`
	_, err := s.validateIsSources(src)
	s.Require().Error(err)
	s.Contains(err.Error(), `"f" is neither a shape nor a value in scope`)
}

func (s *IndexTestSuite) TestResolveIsErrorsInSourceOrder() {
	src := `namespace com/app
policy pol {
  let a = 1 is Zz
  let b = 1 is Yy
  let c = 1 is Xx
  rule r = { yield true }
  export decision of r
}
`
	_, err := s.validateIsSources(src)
	s.Require().Error(err)
	msg := err.Error()
	z, y, x := strings.Index(msg, `"Zz"`), strings.Index(msg, `"Yy"`), strings.Index(msg, `"Xx"`)
	s.Require().True(z >= 0 && y >= 0 && x >= 0, msg)
	s.Less(z, y)
	s.Less(y, x)
}
