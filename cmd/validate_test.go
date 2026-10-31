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

const exponentLiteralPolicy = `namespace com/ex
policy pol {
  rule ok = default false { yield 1e5 == 100000 and 2.5e-3 * 1000 == 2.5 }
  export decision of ok
}
`

// TestValidateCmdAcceptsExponentLiterals pins #180: validate accepts a policy
// that uses exponent number literals.
func (s *CmdTestSuite) TestValidateCmdAcceptsExponentLiterals() {
	dir := s.writeTestPack(exponentLiteralPolicy)
	err := runValidateCLI(s.T().Context(), []string{"com/ex/pol", "--pack-location", dir})
	s.Require().NoError(err)
}

// TestValidateCmdReportsMalformedExponentLiteral pins #180: a malformed
// exponent fails validate with the lexer's own message and the position of the
// literal, not a "no prefix parse function" error.
func (s *CmdTestSuite) TestValidateCmdReportsMalformedExponentLiteral() {
	policy := `namespace com/ex
policy pol {
  rule ok = default false { yield 1e+ == 1 }
  export decision of ok
}
`
	dir := s.writeTestPack(policy)
	err := runValidateCLI(s.T().Context(), []string{"com/ex/pol", "--pack-location", dir})
	s.Require().Error(err)
	s.Contains(err.Error(), `invalid number literal "1e+": the exponent needs at least one digit`)
	s.Contains(err.Error(), "policy.sentrie:3:35-38")
	s.NotContains(err.Error(), "no prefix parse function")
}
