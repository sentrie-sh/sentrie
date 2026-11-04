// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"fmt"

	"github.com/sentrie-sh/sentrie/parser"
	"github.com/stretchr/testify/require"
)

// indexPolicyBody indexes a namespace with one policy whose body is the given
// statements and returns the error from AddProgram.
func (s *IndexTestSuite) indexPolicyBody(name, body string) (*Index, error) {
	ctx := s.T().Context()
	idx := CreateIndex()
	src := fmt.Sprintf("namespace com/ex\npolicy pol {\n%s\n  rule r = { yield true }\n  export decision of r\n}\n", body)
	prog, err := parser.NewParserFromString(src, name).ParseProgram(ctx)
	require.NoError(s.T(), err)
	return idx, idx.AddProgram(ctx, prog)
}

const deriveLine = "  derive d = (): number => { yield 1 }"

func (s *IndexTestSuite) TestPolicyDeriveIsFirstBodyStatement() {
	idx, err := s.indexPolicyBody("first.sentra", deriveLine)
	s.Require().NoError(err)
	s.Contains(idx.Namespaces["com/ex"].Policies["pol"].Derives, "d")
}

func (s *IndexTestSuite) TestPolicyDeriveAfterMetaHeader() {
	idx, err := s.indexPolicyBody("after_meta.sentra", "  title \"t\"\n  description \"d\"\n"+deriveLine)
	s.Require().NoError(err)
	s.Contains(idx.Namespaces["com/ex"].Policies["pol"].Derives, "d")
}

func (s *IndexTestSuite) TestPolicyDeriveAfterFacts() {
	idx, err := s.indexPolicyBody("after_facts.sentra", "  fact f: string\n"+deriveLine)
	s.Require().NoError(err)
	p := idx.Namespaces["com/ex"].Policies["pol"]
	s.Contains(p.Derives, "d")
	s.Contains(p.Facts, "f")
}

func (s *IndexTestSuite) TestPolicyDeriveAfterUses() {
	idx, err := s.indexPolicyBody("after_uses.sentra", "  use { sha256 } from @sentrie/hash as o\n"+deriveLine)
	s.Require().NoError(err)
	p := idx.Namespaces["com/ex"].Policies["pol"]
	s.Contains(p.Derives, "d")
	s.Contains(p.Uses, "o")
}

func (s *IndexTestSuite) TestPolicyFactAfterDeriveIsLateHeader() {
	_, err := s.indexPolicyBody("fact_after_derive.sentra", deriveLine+"\n  fact f: string")
	s.Require().Error(err)
	s.Contains(err.Error(), "'fact' must appear before rules, exports, lets, derives, and shapes")
	s.NotContains(err.Error(), "'derive'")
}

func (s *IndexTestSuite) TestPolicyUseAfterDeriveIsLateHeader() {
	_, err := s.indexPolicyBody("use_after_derive.sentra", deriveLine+"\n  use { sha256 } from @sentrie/hash as o")
	s.Require().Error(err)
	s.Contains(err.Error(), "'use' must appear before rules, exports, lets, derives, and shapes")
	s.NotContains(err.Error(), "'derive'")
}

func (s *IndexTestSuite) TestPolicyMetadataAfterDeriveIsLateHeader() {
	_, err := s.indexPolicyBody("title_after_derive.sentra", deriveLine+"\n  title \"t\"")
	s.Require().Error(err)
	s.Contains(err.Error(), "'title' must appear before rules, exports, lets, derives, and shapes")
}
