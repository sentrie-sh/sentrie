// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"strings"
)

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

const ruleErrorCmdPolicy = `namespace com/ex
policy pol {
  rule ok = { yield true }
  rule bad = { yield 1 / 0 }
  export decision of ok
  export decision of bad
}
policy allbad {
  rule bad1 = { yield 1 / 0 }
  rule bad2 = { yield "x" - 1 }
  export decision of bad1
  export decision of bad2
}
`

// Every rule failing prints every rule and reports "2 of 2 rules failed".
func (s *CmdTestSuite) TestExecPolicyAllRulesFailing() {
	dir := s.writeTestPack(ruleErrorCmdPolicy)
	var err error
	out := s.captureStdout(func() {
		err = runExecCLI(s.T().Context(), []string{"com/ex/allbad", "--pack-location", dir, "--output", "json"})
	})
	s.Require().Error(err)
	s.Equal("2 of 2 rules failed", err.Error())
	var got []struct {
		Rule  string `json:"rule"`
		Error *struct {
			Detail string `json:"detail"`
		} `json:"error"`
	}
	s.Require().NoError(json.Unmarshal([]byte(out), &got), out)
	s.Require().Len(got, 2)
	for _, d := range got {
		s.NotNil(d.Error, d.Rule)
	}
}

// A named rule that fails is the plain error: nothing printed on stdout.
func (s *CmdTestSuite) TestExecNamedRuleFailureIsAPlainError() {
	dir := s.writeTestPack(ruleErrorCmdPolicy)
	var err error
	out := s.captureStdout(func() {
		err = runExecCLI(s.T().Context(), []string{"com/ex/pol/bad", "--pack-location", dir, "--output", "json"})
	})
	s.Require().Error(err)
	s.Contains(err.Error(), "divide by zero")
	s.NotContains(err.Error(), "rules failed")
	s.Empty(out)
}

// The binary exits 1 with "Error: ..." on stderr and a clean, parseable JSON
// document on stdout when a rule failed.
func (s *CmdTestSuite) TestExecBinaryReportsRuleErrorsOnStderr() {
	dir := s.writeTestPack(ruleErrorCmdPolicy)
	bin := s.signalTestBinary()
	var stdout, stderr bytes.Buffer
	c := exec.CommandContext(s.T().Context(), bin, "exec", "com/ex/pol", "--pack-location", dir, "--output", "json")
	c.Stdout, c.Stderr = &stdout, &stderr
	err := c.Run()

	var exitErr *exec.ExitError
	s.Require().ErrorAs(err, &exitErr)
	s.Equal(1, exitErr.ExitCode())
	s.Contains(stderr.String(), "Error: 1 of 2 rules failed")
	s.NotContains(stdout.String(), "Error:")
	var got []map[string]any
	s.Require().NoError(json.Unmarshal(stdout.Bytes(), &got), strings.TrimSpace(stdout.String()))
	s.Len(got, 2)
}
