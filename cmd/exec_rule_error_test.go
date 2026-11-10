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
