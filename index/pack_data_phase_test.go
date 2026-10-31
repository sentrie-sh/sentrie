// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
)

// A data declaration is a fact-phase header statement: it classifies as one,
// and it may not appear once the policy body has begun, like a fact.
func (suite *IndexTestSuite) TestPackDataIsAFactPhaseStatement() {
	r := tokens.Range{File: "t.sentrie"}
	suite.Equal(policyStmtFact, policyStmtKindOf(ast.NewDataStatement("roles", nil, "roles", r)))

	cases := []struct {
		name string
		body string
		ok   bool
	}{
		{"data before use", "  data roles\n  use { now } from @sentrie/time as t\n  rule r = count(roles) > 0\n", true},
		{"data interleaved with facts", "  fact u: string\n  data roles\n  fact v: string\n  rule r = count(roles) > 0\n", true},
		{"data after use, as for a fact", "  use { now } from @sentrie/time as t\n  data roles\n  rule r = count(roles) > 0\n", true},
		{"data after a rule", "  rule q = 1 > 0\n  data roles\n  rule r = count(roles) > 0\n", false},
	}
	for _, tc := range cases {
		suite.Run(tc.name, func() {
			src := "namespace a\npolicy p {\n" + tc.body + "  export decision of r\n}\n"
			addErr, valErr := suite.indexData(dataTestPack(), src)
			if tc.ok {
				suite.Require().NoError(addErr)
				suite.Require().NoError(valErr)
				return
			}
			suite.Require().Error(addErr)
		})
	}
}
