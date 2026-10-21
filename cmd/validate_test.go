// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
)

const testPackToml = `[schema]
version = 1

[pack]
name = "cmdtest"
version = "0.0.1"
`

// writeTestPack writes a pack manifest and one policy file into a temp dir.
func (s *CmdTestSuite) writeTestPack(policy string) string {
	dir := s.T().TempDir()
	s.Require().NoError(os.WriteFile(filepath.Join(dir, "sentrie.pack.toml"), []byte(testPackToml), 0o600))
	s.Require().NoError(os.WriteFile(filepath.Join(dir, "policy.sentrie"), []byte(policy), 0o600))
	return dir
}

func (s *CmdTestSuite) captureStderr(fn func()) string {
	s.T().Helper()
	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	s.Require().NoError(err)
	defer func() { s.Require().NoError(r.Close()) }()
	os.Stderr = w
	defer func() { os.Stderr = oldStderr }()
	fn()
	s.Require().NoError(w.Close())
	var buf bytes.Buffer
	_, err = buf.ReadFrom(r)
	s.Require().NoError(err)
	return buf.String()
}

func runValidateCLI(ctx context.Context, args []string) error {
	cli := Setup(ctx, "test")
	return Execute(ctx, cli, append([]string{"sentrie", "validate"}, args...))
}

const nondeterministicPolicy = `namespace com/ex
policy pol {
  use { v4 } from @sentrie/uuid as uuid
  rule ok = { yield now() > 0 }
  export decision of ok
}
`

func (s *CmdTestSuite) TestValidateCmdWarnsOnNondeterministicSources() {
	dir := s.writeTestPack(nondeterministicPolicy)
	var runErr error
	out := s.captureStderr(func() {
		runErr = runValidateCLI(s.T().Context(), []string{"com/ex/pol", "--pack-location", dir})
	})
	s.Require().NoError(runErr, "warnings must not fail validation")
	s.Contains(out, "warning: ")
	s.Contains(out, "@sentrie/uuid.v4 is nondeterministic")
	s.Contains(out, `builtin "now" is nondeterministic`)
	s.Contains(out, "policy.sentrie:3:")
	s.Contains(out, "policy.sentrie:4:")
}

func (s *CmdTestSuite) TestValidateCmdQuietForDeterministicPack() {
	dir := s.writeTestPack(`namespace com/ex
policy pol {
  rule ok = { yield count([1]) > 0 }
  export decision of ok
}
`)
	var runErr error
	out := s.captureStderr(func() {
		runErr = runValidateCLI(s.T().Context(), []string{"com/ex/pol", "--pack-location", dir})
	})
	s.Require().NoError(runErr)
	s.NotContains(out, "warning:")
}
