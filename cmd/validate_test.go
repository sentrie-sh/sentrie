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

const unknownWithFactPolicy = `namespace t
policy q {
  fact user?: document as user default {"role": "guest"}
  rule ok = user.role == "admin"
  export decision of ok
}
policy p {
  rule typo  = import decision ok from t/q with usr  as {"role": "admin"}
  rule right = import decision ok from t/q with user as {"role": "admin"}
  export decision of typo
  export decision of right
}
`

// TestValidateCmdRejectsUnknownImportWithFact is the reproduction from #151:
// validate used to exit 0 on a typo'd with name.
func (s *CmdTestSuite) TestValidateCmdRejectsUnknownImportWithFact() {
	dir := s.writeTestPack(unknownWithFactPolicy)
	err := runValidateCLI(s.T().Context(), []string{"t/p", "--pack-location", dir})
	s.Require().Error(err)
	s.Contains(err.Error(), `import binds unknown fact "usr"`)
	s.Contains(err.Error(), "policy t/q accepts fact aliases: user")
}

const deriveFirstPolicy = `namespace t/a
policy p {
  derive n = (t: string): number => {
    yield count(t)
  }
  rule ok = { yield n("x") == 1 }
  export decision of ok
}
`

// TestValidateCmdAcceptsPolicyDeriveFirst is the reproduction from #300:
// validate used to fail with a late-header error on a derive that started the
// policy body.
func (s *CmdTestSuite) TestValidateCmdAcceptsPolicyDeriveFirst() {
	dir := s.writeTestPack(deriveFirstPolicy)
	err := runValidateCLI(s.T().Context(), []string{"t/a/p", "--pack-location", dir})
	s.Require().NoError(err)
}

// TestValidateCmdRejectsFactAfterPolicyDerive pins the other side: a derive
// starts the body, so a fact after it is still a late header, and the message
// names derives among the statements that start the body.
func (s *CmdTestSuite) TestValidateCmdRejectsFactAfterPolicyDerive() {
	dir := s.writeTestPack(`namespace t/a
policy p {
  derive n = (t: string): number => { yield count(t) }
  fact f: string
  rule ok = { yield n(f) == 1 }
  export decision of ok
}
`)
	err := runValidateCLI(s.T().Context(), []string{"t/a/p", "--pack-location", dir})
	s.Require().Error(err)
	s.Contains(err.Error(), "'fact' must appear before rules, exports, lets, derives, and shapes")
}
