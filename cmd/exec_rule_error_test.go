// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package cmd

import "encoding/json"

// A policy with a failing rule prints every rule's result, the failing one
// with its error, and exec still fails (#143).
func (s *CmdTestSuite) TestExecPolicyPrintsEveryRuleAndFailsOnARuleError() {
	dir := s.writeTestPack(lambdaOutputCmdPolicy)
	var err error
	out := s.captureStdout(func() {
		err = runExecCLI(s.T().Context(), []string{"com/lam/pol", "--pack-location", dir, "--output", "json"})
	})
	s.Require().Error(err)
	s.Equal("1 of 3 rules failed", err.Error())

	var got []struct {
		Rule  string `json:"rule"`
		Error *struct {
			Title  string `json:"title"`
			Status int    `json:"status"`
			Detail string `json:"detail"`
		} `json:"error"`
	}
	s.Require().NoError(json.Unmarshal([]byte(out), &got), out)
	s.Require().Len(got, 3)
	for _, d := range got {
		if d.Rule != "whole" {
			s.Nil(d.Error, d.Rule)
			continue
		}
		s.Require().NotNil(d.Error)
		s.Equal("Evaluation Failed", d.Error.Title)
		s.Equal(500, d.Error.Status)
		s.Contains(d.Error.Detail, "cannot be a callable")
	}
}

// The table output lists a failing rule's error under Errors.
func (s *CmdTestSuite) TestExecPolicyTableShowsRuleErrors() {
	dir := s.writeTestPack(lambdaOutputCmdPolicy)
	var err error
	out := s.captureStdout(func() {
		err = runExecCLI(s.T().Context(), []string{"com/lam/pol", "--pack-location", dir})
	})
	s.Require().Error(err)
	s.Contains(out, "Errors:")
	s.Contains(out, "! whole: rule com/lam/pol/whole: a rule's value cannot be a callable")
	s.Contains(out, "nested:")
}

// A missing required fact on a policy is a plain error: no per-rule output.
func (s *CmdTestSuite) TestExecPolicyMissingFactIsAPlainError() {
	dir := s.writeTestPack(`namespace com/ex
policy pol {
  fact user: any
  rule a = { yield true }
  rule b = { yield 1 / 0 }
  export decision of a
  export decision of b
}
`)
	var err error
	out := s.captureStdout(func() {
		err = runExecCLI(s.T().Context(), []string{"com/ex/pol", "--pack-location", dir})
	})
	s.Require().Error(err)
	s.Contains(err.Error(), "required fact not found: user")
	s.NotContains(err.Error(), "rules failed")
	s.Empty(out)
}
