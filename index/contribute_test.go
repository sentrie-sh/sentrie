// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/parser"
	"github.com/sentrie-sh/sentrie/tokens"
	"github.com/sentrie-sh/sentrie/xerr"
)

// contributeIndexErr indexes and validates a one-policy program built from body.
func (s *IndexTestSuite) contributeIndexErr(body string) (*Index, error) {
	ctx := s.T().Context()
	src := "namespace com/ex\npolicy p {\n  fact user: document as user\n" + body + "\n}\n"
	prog, err := parser.NewParserFromString(src, "contribute.sentrie").ParseProgram(ctx)
	s.Require().NoError(err)
	idx := CreateIndex()
	if err := idx.AddProgram(ctx, prog); err != nil {
		return idx, err
	}
	return idx, idx.Validate(ctx)
}

func (s *IndexTestSuite) TestContributionsAttachInDeclarationOrder() {
	idx, err := s.contributeIndexErr(`
  rule deny: list[string] = []
  contribute deny when user.disabled { yield "disabled" }
  contribute deny { yield "always" }
  rule allow: boolean = false
  contribute allow when user.admin { yield true }
  export decision of deny
  export decision of allow`)
	s.Require().NoError(err)

	p, err := idx.ResolvePolicy("com/ex", "p")
	s.Require().NoError(err)
	deny := p.Rules["deny"]
	s.Require().Len(deny.Contributions, 2)
	s.NotNil(deny.Contributions[0].When)
	s.Nil(deny.Contributions[1].When)
	s.Same(deny, deny.Contributions[0].Rule)
	s.Len(p.Rules["allow"].Contributions, 1)
}

func (s *IndexTestSuite) TestContributionIndexErrors() {
	cases := map[string]struct {
		body string
		want string
	}{
		"undeclared rule": {
			body: `rule a = true
  contribute deny { yield "x" }
  export decision of a`,
			want: "cannot contribute to unknown rule 'deny'",
		},
		"rule declared after the contribution": {
			body: `contribute deny { yield "x" }
  rule deny: list[string] = []
  export decision of deny`,
			want: "cannot contribute to unknown rule 'deny'",
		},
		"untyped rule": {
			body: `rule deny = []
  contribute deny { yield "x" }
  export decision of deny`,
			want: "the rule declares no type",
		},
		"non-list, non-boolean rule": {
			body: `rule deny: string = "x"
  contribute deny { yield "y" }
  export decision of deny`,
			want: "the rule declares type string",
		},
		"literal yield of the wrong element type": {
			body: `rule deny: list[string] = []
  contribute deny { yield 42 }
  export decision of deny`,
			want: "yields a number, but the rule takes string",
		},
		"literal yield of the wrong boolean type": {
			body: `rule allow: boolean = false
  contribute allow when user.admin "yes"
  export decision of allow`,
			want: "yields a string, but the rule takes trinary",
		},
		"dict literal yield to a list of records": {
			body: `rule pairs: list[record[string, number]] = []
  contribute pairs { yield {"a": 1} }
  export decision of pairs`,
			want: "yields a dict, but the rule takes record",
		},
		"contribution guard reads its own rule": {
			body: `rule deny: list[string] = []
  contribute deny when count(deny) > 0 { yield "loop" }
  export decision of deny`,
			want: "reads the rule itself",
		},
		"contribution body reads its own rule": {
			body: `rule deny: list[string] = []
  contribute deny { yield count(deny) }
  export decision of deny`,
			want: "reads the rule itself",
		},
		"null yield to a non-nullable element": {
			body: `rule deny: list[string] = []
  contribute deny { yield null }
  export decision of deny`,
			want: "yields a null",
		},
	}
	for name, c := range cases {
		s.Run(name, func() {
			_, err := s.contributeIndexErr(c.body)
			s.Require().Error(err)
			s.ErrorIs(err, xerr.ErrIndex)
			s.Contains(err.Error(), c.want)
		})
	}
}

func (s *IndexTestSuite) TestContributionToImportedRuleIsRejected() {
	ctx := s.T().Context()
	src := `namespace com/ex
policy other {
  rule allow: boolean = true
  export decision of allow
}
policy p {
  rule allow: boolean = import decision allow from com/ex/other
  contribute allow { yield true }
  export decision of allow
}
`
	prog, err := parser.NewParserFromString(src, "contribute.sentrie").ParseProgram(ctx)
	s.Require().NoError(err)
	err = CreateIndex().AddProgram(ctx, prog)
	s.Require().Error(err)
	s.Contains(err.Error(), "imports its decision")
}

func (s *IndexTestSuite) TestContributionNullableAndDocumentElementsPassStatically() {
	_, err := s.contributeIndexErr(`
  rule maybe: list[string?] = []
  contribute maybe { yield null }
  rule docs: list[document] = []
  contribute docs { yield {"a": 1} }
  export decision of maybe
  export decision of docs`)
	s.NoError(err)
}

func (s *IndexTestSuite) TestContributionReferenceCycleIsDetected() {
	_, err := s.contributeIndexErr(`
  rule deny: list[string] = []
  rule seen = count(deny)
  contribute deny when seen > 0 { yield "loop" }
  export decision of deny`)
	s.Require().Error(err)
	s.Contains(err.Error(), "deny")
	s.Contains(err.Error(), "seen")
}

func (s *IndexTestSuite) TestContributionsThroughEachOtherAreAReferenceCycle() {
	_, err := s.contributeIndexErr(`
  rule a: list[string] = []
  rule b: list[string] = []
  contribute a { yield count(b) }
  contribute b { yield count(a) }
  export decision of a
  export decision of b`)
	s.Require().Error(err)
	s.Contains(err.Error(), "infinite recursion")
}

func (s *IndexTestSuite) TestContributionMayReadAnotherRuleWithoutACycle() {
	_, err := s.contributeIndexErr(`
  rule limit = 3
  rule deny: list[string] = []
  contribute deny when user.n > limit { yield "over" }
  rule allow: boolean = count(deny) == 0
  export decision of deny
  export decision of allow`)
	s.NoError(err)
}

func (s *IndexTestSuite) TestContributionToTrinaryAndNullableRulesAttaches() {
	idx, err := s.contributeIndexErr(`
  rule t: trinary = false
  contribute t { yield unknown }
  rule n: boolean? = false
  contribute n { yield true }
  rule l: list[string]? = []
  contribute l { yield "x" }
  export decision of t
  export decision of n
  export decision of l`)
	s.Require().NoError(err)
	p, err := idx.ResolvePolicy("com/ex", "p")
	s.Require().NoError(err)
	for _, name := range []string{"t", "n", "l"} {
		s.Len(p.Rules[name].Contributions, 1, name)
	}
}

func (s *IndexTestSuite) TestContributeIsAReservedWord() {
	ctx := s.T().Context()
	for name, body := range map[string]string{
		"as a rule name": "rule contribute = true",
		"as a let name":  "let contribute = 1",
	} {
		s.Run(name, func() {
			src := "namespace com/ex\npolicy p {\n" + body + "\n}\n"
			_, err := parser.NewParserFromString(src, "reserved.sentrie").ParseProgram(ctx)
			s.Error(err)
		})
	}
}

func (s *IndexTestSuite) TestContributeIsAPolicyBodyStatement() {
	stmt := ast.NewContributeStatement("deny", nil, ast.NewNullLiteral(tokens.Range{}), tokens.Range{})
	s.Equal(policyStmtBody, policyStmtKindOf(stmt))
	s.True(isBodyStmt(stmt))
}

func (s *IndexTestSuite) TestBuiltinCallsInContributionsAreChecked() {
	cases := map[string]string{
		"guard": `
  rule deny: list[string] = []
  contribute deny when count() > 0 { yield "x" }
  export decision of deny`,
		"body": `
  rule deny: list[string] = []
  contribute deny { yield count() }
  export decision of deny`,
	}
	for name, body := range cases {
		s.Run(name, func() {
			_, err := s.contributeIndexErr(body)
			s.Require().Error(err)
			s.ErrorIs(err, xerr.ErrIndex)
			s.Contains(err.Error(), "count")
		})
	}
}

func (s *IndexTestSuite) TestContributionRecordAndShadowedNamesPass() {
	_, err := s.contributeIndexErr(`
  rule pairs: list[record[string, number]] = []
  contribute pairs { yield ["a", 1] }
  rule deny: list[string] = []
  contribute deny when count(filter(["x"], (deny) => { yield deny == "x" })) > 0 { yield "shadow" }
  export decision of pairs
  export decision of deny`)
	s.NoError(err)
}
