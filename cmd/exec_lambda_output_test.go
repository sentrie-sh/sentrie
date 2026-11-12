// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package cmd

import "encoding/json"

const lambdaOutputCmdPolicy = `namespace com/lam
policy pol {
  let id = (x) => { yield x }
  rule nested = { yield { "f": id, "xs": [1, id] } }
  rule filtered = { yield filter([1, 2, 3], (x) => { yield x > 1 }) }
  rule whole = { yield id }
  export decision of nested
  export decision of filtered
  export decision of whole
}
`

// A lambda nested in the decision value encodes as the callable placeholder.
func (s *CmdTestSuite) TestExecJSONEncodesNestedLambda() {
	dir := s.writeTestPack(lambdaOutputCmdPolicy)
	var err error
	out := s.captureStdout(func() {
		err = runExecCLI(s.T().Context(), []string{"com/lam/pol/nested", "--pack-location", dir, "--output", "json"})
	})
	s.Require().NoError(err, out)
	var got []struct {
		Decision struct {
			Value json.RawMessage `json:"value"`
		} `json:"decision"`
	}
	s.Require().NoError(json.Unmarshal([]byte(out), &got), out)
	s.Require().Len(got, 1)
	s.JSONEq(`{"f":{"kind":"callable"},"xs":[1,{"kind":"callable"}]}`, string(got[0].Decision.Value))
}

// A lambda that only appears in the trace (a filter argument) no longer fails the encode.
func (s *CmdTestSuite) TestExecJSONEncodesLambdaInTrace() {
	dir := s.writeTestPack(lambdaOutputCmdPolicy)
	var err error
	out := s.captureStdout(func() {
		err = runExecCLI(s.T().Context(), []string{"com/lam/pol/filtered", "--pack-location", dir, "--output", "json"})
	})
	s.Require().NoError(err, out)
	s.True(json.Valid([]byte(out)), "exec output is not valid JSON")
	s.Contains(out, `"kind": "callable"`)
}

// A rule whose whole value is a lambda is that rule's error: exec fails.
func (s *CmdTestSuite) TestExecWholeValueLambdaIsRuleError() {
	dir := s.writeTestPack(lambdaOutputCmdPolicy)
	var err error
	s.captureStdout(func() {
		err = runExecCLI(s.T().Context(), []string{"com/lam/pol/whole", "--pack-location", dir, "--output", "json"})
	})
	s.Require().Error(err)
	s.Contains(err.Error(), "com/lam/pol/whole")
	s.Contains(err.Error(), "cannot be a callable")
}
