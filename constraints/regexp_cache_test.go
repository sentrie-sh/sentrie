// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package constraints_test

import (
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/constraints"
)

// TestRegexpPatternIsCompiledOnce checks that a regexp constraint's pattern
// is compiled once and reused by the argument check and every match, rather
// than recompiled each time a value is checked.
func (s *ConstraintsTestSuite) TestRegexpPatternIsCompiledOnce() {
	const pattern = `^cache-[a-z]+$`
	def := constraints.StringContraintCheckers["regexp"]
	args := []box.Value{box.String(pattern)}

	s.Require().NoError(def.CheckArgs(args))
	first, err := constraints.CompilePattern(pattern)
	s.Require().NoError(err)

	s.NoError(def.Checker(s.T().Context(), box.String("cache-hit"), args))
	s.Error(def.Checker(s.T().Context(), box.String("miss"), args))

	again, err := constraints.CompilePattern(pattern)
	s.Require().NoError(err)
	s.Same(first, again, "the pattern was recompiled")
}

// TestRegexpInvalidPatternIsNotCached checks that a pattern that fails to
// compile reports the same error every time.
func (s *ConstraintsTestSuite) TestRegexpInvalidPatternIsNotCached() {
	def := constraints.StringContraintCheckers["regexp"]
	for range 2 {
		err := def.CheckArgs([]box.Value{box.String("[")})
		s.Require().Error(err)
		s.Contains(err.Error(), "invalid regexp pattern")
	}
}
