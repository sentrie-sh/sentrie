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
