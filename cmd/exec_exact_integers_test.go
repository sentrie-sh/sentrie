// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const exactIntegersCmdPolicy = `namespace com/ids
policy pol {
  fact id: number
  fact other: number
  rule echo = { yield id }
  rule same = { yield id == other }
  rule differ = { yield id != other }
  rule successor = { yield id + 1 }
  export decision of echo
  export decision of same
  export decision of differ
  export decision of successor
}
`

// execDecisionValue runs `sentrie exec` for one rule with JSON output and
// returns the raw JSON text of the decision value, so the digits the user would
// see are compared, not a re-parsed float64.
func (s *CmdTestSuite) execDecisionValue(dir, rule string, extra ...string) string {
	s.T().Helper()
	args := append([]string{"com/ids/pol/" + rule, "--pack-location", dir, "--output", "json"}, extra...)
	var runErr error
	out := s.captureStdout(func() { runErr = runExecCLI(s.T().Context(), args) })
	s.Require().NoError(runErr, rule)

	var results []struct {
		Decision struct {
			Value json.RawMessage `json:"value"`
		} `json:"decision"`
	}
	s.Require().NoError(json.Unmarshal([]byte(out), &results), out)
	s.Require().Len(results, 1)
	return string(results[0].Decision.Value)
}

// TestExecKeepsLargeIntegerFacts covers #166 for --facts: two IDs one apart
// above 2^53 stay distinct and are echoed digit for digit.
func (s *CmdTestSuite) TestExecKeepsLargeIntegerFacts() {
	dir := s.writeTestPack(exactIntegersCmdPolicy)
	facts := []string{"--facts", `{"id": 9007199254740993, "other": 9007199254740992}`}

	s.Equal("9007199254740993", s.execDecisionValue(dir, "echo", facts...))
	s.Equal("false", s.execDecisionValue(dir, "same", facts...))
	s.Equal("true", s.execDecisionValue(dir, "differ", facts...))
	s.Equal("9007199254740994", s.execDecisionValue(dir, "successor", facts...))
}

// TestExecKeepsLargeIntegerFactFile covers #166 for --fact-file, and that a
// --facts value overrides the file with its digits intact.
func (s *CmdTestSuite) TestExecKeepsLargeIntegerFactFile() {
	dir := s.writeTestPack(exactIntegersCmdPolicy)
	factFile := filepath.Join(s.T().TempDir(), "facts.json")
	s.Require().NoError(os.WriteFile(factFile, []byte(`{"id": 9007199254740993, "other": 9007199254740992}`), 0o600))
	fromFile := []string{"--fact-file", factFile}

	s.Equal("9007199254740993", s.execDecisionValue(dir, "echo", fromFile...))
	s.Equal("false", s.execDecisionValue(dir, "same", fromFile...))
	s.Equal("true", s.execDecisionValue(dir, "differ", fromFile...))

	override := append(fromFile, "--facts", `{"other": 9007199254740993}`)
	s.Equal("true", s.execDecisionValue(dir, "same", override...))
	s.Equal("9007199254740993", s.execDecisionValue(dir, "echo", override...))
}

// TestExecLargeIntegerFactsInTableOutput checks the default table output also
// prints the exact digits and no exponent.
func (s *CmdTestSuite) TestExecLargeIntegerFactsInTableOutput() {
	dir := s.writeTestPack(exactIntegersCmdPolicy)
	var runErr error
	out := s.captureStdout(func() {
		runErr = runExecCLI(s.T().Context(), []string{"com/ids/pol/echo", "--pack-location", dir,
			"--facts", `{"id": 9007199254740993, "other": 1}`})
	})
	s.Require().NoError(runErr)
	s.Contains(out, "9007199254740993")
	s.NotContains(out, "e+")
}
