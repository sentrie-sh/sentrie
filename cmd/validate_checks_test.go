// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package cmd

func (s *CmdTestSuite) validatePolicy(body string) error {
	dir := s.writeTestPack("namespace com/ex\npolicy pol {\n" + body + "}\n")
	return runValidateCLI(s.T().Context(), []string{"--pack-location", dir})
}

// TestValidateCmdRejectsSelfReference is #155: `rule r = r` used to exit 0.
func (s *CmdTestSuite) TestValidateCmdRejectsSelfReference() {
	err := s.validatePolicy("  rule r = r\n  export decision of r\n")
	s.Require().Error(err)
	s.Contains(err.Error(), "rule 'com/ex/pol/r' references itself")
	s.Contains(err.Error(), "policy.sentrie:3:")
}

func (s *CmdTestSuite) TestValidateCmdRejectsUndefinedIdentifier() {
	err := s.validatePolicy("  rule r = { yield undefined_thing }\n  export decision of r\n")
	s.Require().Error(err)
	s.Contains(err.Error(), "identifier 'undefined_thing' not found at ")
}

func (s *CmdTestSuite) TestValidateCmdHintsBuiltinUsedAsValue() {
	err := s.validatePolicy("  rule r = { yield count }\n  export decision of r\n")
	s.Require().Error(err)
	s.Contains(err.Error(), "'count' is a builtin, not a value; call it as count(...)")
}

func (s *CmdTestSuite) TestValidateCmdAcceptsShadowedNames() {
	s.NoError(s.validatePolicy("  rule x = { let x = 1\n yield x == 1 }\n  export decision of x\n"))
}

// TestValidateCmdAcceptsLetShadowingCachedOuterBinding is the PR review case:
// the outer x is read and cached first, so the inner let's `x` reads the
// cached 1 at runtime instead of recursing. validate must not reject a pack
// that exec evaluates to true.
// A fact of the same name is read before any let, so the inner let is never
// evaluated there either. A same-policy rule runs in its caller's context, so
// r2's let reads the caller's cached block let or lambda parameter.
func (s *CmdTestSuite) TestValidateCmdAcceptsLetShadowingCachedOuterBinding() {
	cases := map[string]string{
		"policy let": "  let x = 1\n  rule r = (x + { let x = x + 1 yield x }) == 2\n",
		"rule":       "  rule x = 1\n  rule r = (x + { let x = x + 1 yield x }) == 2\n",
		"fact alias": "  fact x?: number default 1\n  rule r = { let x = x + 1 yield x } == 1\n",
		// G and H: r2 runs in r's context and reads r's cached x.
		"calling rule's block let":        "  rule r2 = { let x = x + 1 yield x }\n  rule r = { let x = 1 yield x + r2 } == 2\n",
		"calling rule's lambda parameter": "  rule r2 = { let x = x + 1 yield x }\n  rule r = count(filter([1], (x) => { yield x + r2 == 2 })) == 1\n",
	}
	for name, body := range cases {
		s.Run(name, func() {
			dir := s.writeTestPack("namespace com/ex\npolicy pol {\n" + body + "  export decision of r\n}\n")
			s.Require().NoError(runValidateCLI(s.T().Context(), []string{"--pack-location", dir}))
			// Table output: `--output json` fails on a pack that evaluates a
			// lambda ("cannot marshal callable value"), on the base too.
			var runErr error
			out := s.captureStdout(func() {
				runErr = runExecCLI(s.T().Context(), []string{"com/ex/pol/r", "--pack-location", dir})
			})
			s.Require().NoError(runErr)
			s.Contains(out, "r: true", out)
		})
	}
}

// TestValidateCmdStillRejectsUnshadowedLetSelfReference: with no outer x,
// `let x = x + 1` fails at runtime and is still a validation error.
func (s *CmdTestSuite) TestValidateCmdStillRejectsUnshadowedLetSelfReference() {
	err := s.validatePolicy("  rule r = { let x = x + 1 yield x } == 2\n  export decision of r\n")
	s.Require().Error(err)
	s.Contains(err.Error(), "let 'x' references itself")
}

// TestValidateCmdStillAcceptsStaticLiteralTypeError pins the deferred half of
// #155: `1 + true` is not rejected by validate yet.
func (s *CmdTestSuite) TestValidateCmdStillAcceptsStaticLiteralTypeError() {
	s.NoError(s.validatePolicy("  rule r = { yield 1 + true }\n  export decision of r\n"))
}

// TestValidateCmdRejectsStrayPositionalArgument covers the PROVISIONAL
// decision: the removed rule argument is now an error, not silently ignored.
func (s *CmdTestSuite) TestValidateCmdRejectsStrayPositionalArgument() {
	dir := s.writeTestPack("namespace com/ex\npolicy pol {\n  rule r = true\n  export decision of r\n}\n")
	err := runValidateCLI(s.T().Context(), []string{"--pack-location", dir, "com/ex/pol/r"})
	s.Require().Error(err)
	s.Contains(err.Error(), `validate takes no arguments, got "com/ex/pol/r"`)
	s.Contains(err.Error(), "--pack-location")
	// Before the flags too.
	err = runValidateCLI(s.T().Context(), []string{"com/ex/pol/r", "--pack-location", dir})
	s.Require().Error(err)
	s.Contains(err.Error(), "validate takes no arguments")
}

// TestValidateCmdRejectsSingleDashWord: cling treats -x as a positional, so
// validate must reject it itself rather than pass silently.
func (s *CmdTestSuite) TestValidateCmdRejectsSingleDashWord() {
	dir := s.writeTestPack("namespace com/ex\npolicy pol {\n  rule r = true\n  export decision of r\n}\n")
	err := runValidateCLI(s.T().Context(), []string{"--pack-location", dir, "-x"})
	s.Require().Error(err)
	s.Contains(err.Error(), `validate: unknown flag "-x"`)
	s.Contains(err.Error(), "two dashes")
}

// TestValidateCmdRejectsRemovedFactsFlag covers the removed --facts flag.
func (s *CmdTestSuite) TestValidateCmdRejectsRemovedFactsFlag() {
	dir := s.writeTestPack("namespace com/ex\npolicy pol {\n  rule r = true\n  export decision of r\n}\n")
	err := runValidateCLI(s.T().Context(), []string{"--pack-location", dir, "--facts", "{}"})
	s.Require().Error(err)
}
