// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"os"
	"path/filepath"
)

const numberRangeCmdPolicy = `namespace com/range
policy pol {
  fact a: number
  rule echo = { yield a }
  rule overflow = { yield a * 10 }
  export decision of echo
  export decision of overflow
}
`

// TestExecRejectsOutOfRangeNumberFacts: with UseNumber a JSON number float64
// cannot hold no longer fails in the decoder, so exec must reject it itself,
// as main did, instead of evaluating it as an infinity.
func (s *CmdTestSuite) TestExecRejectsOutOfRangeNumberFacts() {
	dir := s.writeTestPack(numberRangeCmdPolicy)
	ctx := s.T().Context()

	err := runExecCLI(ctx, []string{"com/range/pol/echo", "--pack-location", dir, "--output", "json",
		"--facts", `{"a": 1e400}`})
	s.Require().Error(err)
	s.Contains(err.Error(), "number 1e400 is out of range")

	factFile := filepath.Join(s.T().TempDir(), "facts.json")
	s.Require().NoError(os.WriteFile(factFile, []byte(`{"a": [-1e500]}`), 0o600))
	err = runExecCLI(ctx, []string{"com/range/pol/echo", "--pack-location", dir, "--output", "json",
		"--fact-file", factFile})
	s.Require().Error(err)
	s.Contains(err.Error(), "out of range")
}

// TestExecJSONOutputSurfacesEncodingErrors: a non-finite result from
// arithmetic cannot be encoded as JSON, and exec must say so rather than print
// nothing and exit zero.
func (s *CmdTestSuite) TestExecJSONOutputSurfacesEncodingErrors() {
	dir := s.writeTestPack(numberRangeCmdPolicy)
	var err error
	out := s.captureStdout(func() {
		err = runExecCLI(s.T().Context(), []string{"com/range/pol/overflow", "--pack-location", dir, "--output", "json",
			"--facts", `{"a": 1e308}`})
	})
	s.Require().Error(err, out)
	s.Contains(err.Error(), "encoding output")
}
