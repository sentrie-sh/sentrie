// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"os"
	"path/filepath"
)

const keywordNamespacePolicy = `namespace constraints/in

shape Item {
  name: string
}
export shape Item

policy p {
  fact item: Item
  rule ok = { yield item.name == "a" }
  export decision of ok
}
`

// TestValidateCmdAcceptsKeywordSegmentNamespace is the reproduction from #297:
// validate used to fail with "expected Ident, got in" on the namespace line.
func (s *CmdTestSuite) TestValidateCmdAcceptsKeywordSegmentNamespace() {
	dir := s.writeTestPack(keywordNamespacePolicy)
	s.Require().NoError(runValidateCLI(s.T().Context(), []string{"constraints/in/p", "--pack-location", dir}))
}

// TestValidateCmdAcceptsKeywordSegmentNamespaceImport checks a keyword-segment
// namespace through the whole CLI path when another file imports its decision
// and uses its shape.
func (s *CmdTestSuite) TestValidateCmdAcceptsKeywordSegmentNamespaceImport() {
	dir := s.writeTestPack(keywordNamespacePolicy)
	s.Require().NoError(os.WriteFile(filepath.Join(dir, "other.sentrie"), []byte(`namespace text/string

policy q {
  fact i: constraints/in/Item
  rule imported = import decision ok from constraints/in/p with item as i
  export decision of imported
}
`), 0o600))
	s.Require().NoError(runValidateCLI(s.T().Context(), []string{"text/string/q", "--pack-location", dir}))
}

// TestValidateCmdRejectsKeywordFirstSegment checks the error a user sees when
// the first segment is a keyword.
func (s *CmdTestSuite) TestValidateCmdRejectsKeywordFirstSegment() {
	dir := s.writeTestPack("namespace in/x\npolicy p {\n  rule ok = true\n  export decision of ok\n}\n")
	err := runValidateCLI(s.T().Context(), []string{"in/x/p", "--pack-location", dir})
	s.Require().Error(err)
	s.Contains(err.Error(), "the first segment of a name cannot be a keyword")
}

// TestValidateCmdReportsKeywordSegmentCalleeInsteadOfPanicking is the panic
// from #297: a call through a slash path whose segment is a keyword that starts
// an expression (`not`) used to crash validate with a nil pointer in the call
// parser. It must be a reported error.
func (s *CmdTestSuite) TestValidateCmdReportsKeywordSegmentCalleeInsteadOfPanicking() {
	for _, call := range []string{"a/not/f(1)", "a/not/b/f(1)", "x and a/not/f(1)"} {
		dir := s.writeTestPack(`namespace app
policy p {
  rule ok = ` + call + `
  export decision of ok
}
`)
		var err error
		s.NotPanics(func() {
			err = runValidateCLI(s.T().Context(), []string{"app/p", "--pack-location", dir})
		}, "call: %s", call)
		s.Require().Error(err, "call: %s", call)
	}
}
