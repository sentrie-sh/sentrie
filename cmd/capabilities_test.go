// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/sentrie-sh/sentrie/xerr"
)

func (s *CmdTestSuite) writeCapabilities(src string) string {
	path := filepath.Join(s.T().TempDir(), "caps.json")
	s.Require().NoError(os.WriteFile(path, []byte(src), 0o600))
	return path
}

func runExecCLI(ctx context.Context, args []string) error {
	cli := Setup(ctx, "test")
	return Execute(ctx, cli, append([]string{"sentrie", "exec"}, args...))
}

const capabilitiesCmdPolicy = `namespace com/ex
policy pol {
  use { sha256 } from @sentrie/hash as hash
  rule ok = { yield count([1]) > 0 }
  export decision of ok
}
`

func (s *CmdTestSuite) TestValidateCmdFailsOutsideTheGrantWithSpan() {
	dir := s.writeTestPack(capabilitiesCmdPolicy)
	caps := s.writeCapabilities(`{"version": 1, "builtins": [], "modules": []}`)

	err := runValidateCLI(s.T().Context(), []string{"com/ex/pol", "--pack-location", dir, "--capabilities", caps})
	s.Require().Error(err)
	s.True(errors.Is(err, xerr.ErrNotGranted))
	s.Contains(err.Error(), `policy.sentrie:3:`)
	s.Contains(err.Error(), `module "@sentrie/hash"`)
	s.Contains(err.Error(), `policy.sentrie:4:`)
	s.Contains(err.Error(), `builtin "count"`)
}

func (s *CmdTestSuite) TestValidateCmdPassesInsideTheGrant() {
	dir := s.writeTestPack(capabilitiesCmdPolicy)
	caps := s.writeCapabilities(`{"version": 1, "builtins": ["count"], "modules": ["@sentrie/hash"], "local_modules": false}`)
	s.Require().NoError(runValidateCLI(s.T().Context(), []string{"com/ex/pol", "--pack-location", dir, "--capabilities", caps}))
	// Without a file everything is granted.
	s.Require().NoError(runValidateCLI(s.T().Context(), []string{"com/ex/pol", "--pack-location", dir}))
}

func (s *CmdTestSuite) TestExecCmdFailsOutsideTheGrant() {
	dir := s.writeTestPack(capabilitiesCmdPolicy)
	caps := s.writeCapabilities(`{"version": 1, "builtins": ["count"], "modules": []}`)
	err := runExecCLI(s.T().Context(), []string{"com/ex/pol/ok", "--pack-location", dir, "--capabilities", caps})
	s.Require().Error(err)
	s.True(errors.Is(err, xerr.ErrNotGranted))

	caps = s.writeCapabilities(`{"version": 1, "builtins": ["count"], "modules": ["@sentrie/hash"]}`)
	s.captureStdout(func() {
		s.Require().NoError(runExecCLI(s.T().Context(), []string{"com/ex/pol/ok", "--pack-location", dir, "--capabilities", caps}))
	})
}

func (s *CmdTestSuite) TestCommandsRejectABadCapabilitiesFile() {
	dir := s.writeTestPack(capabilitiesCmdPolicy)
	bad := s.writeCapabilities(`{"version": 1, "unknown": true}`)
	ctx := s.T().Context()

	err := runValidateCLI(ctx, []string{"com/ex/pol", "--pack-location", dir, "--capabilities", bad})
	s.Require().Error(err)
	s.Contains(err.Error(), `unknown field "unknown"`)

	err = runExecCLI(ctx, []string{"com/ex/pol/ok", "--pack-location", dir, "--capabilities", filepath.Join(dir, "missing.json")})
	s.Require().Error(err)

	err = runServeCLI(ctx, []string{"--pack-location", dir, "--capabilities", bad})
	s.Require().Error(err)
	s.Contains(err.Error(), `unknown field "unknown"`)
}

func (s *CmdTestSuite) TestCommandsDocumentTheCapabilitiesFlag() {
	for _, run := range []func(context.Context, []string) error{runExecCLI, runServeCLI, runValidateCLI} {
		out := s.captureStdout(func() {
			s.Require().NoError(run(s.T().Context(), []string{"--help"}))
		})
		s.Contains(out, "--capabilities")
	}
}
