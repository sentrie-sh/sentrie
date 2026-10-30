// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"github.com/sentrie-sh/sentrie/parser"
)

func (suite *IndexTestSuite) TestValidate_ReferenceCycleThroughIsPredicates() {
	cases := map[string]string{
		"is defined":     "rule a = { yield b is defined }\n  rule b = { yield a is defined }",
		"is not defined": "rule a = { yield b is not defined }\n  rule b = { yield a is not defined }",
		"is empty":       "rule a = { yield b is empty }\n  rule b = { yield a is empty }",
		"let is empty":   "let x = y is empty\n  let y = x is empty\n  rule a = { yield x }",
	}
	for name, body := range cases {
		suite.Run(name, func() {
			ctx := suite.T().Context()
			idx := CreateIndex()
			src := "namespace com/ex\npolicy pol {\n  " + body + "\n  export decision of a\n}\n"
			prog, err := parser.NewParserFromString(src, "is_cycle.sentrie").ParseProgram(ctx)
			suite.Require().NoError(err)
			suite.Require().NoError(idx.AddProgram(ctx, prog))
			err = idx.Validate(ctx)
			suite.Require().Error(err)
			suite.Contains(err.Error(), "infinite recursion")
		})
	}
}
