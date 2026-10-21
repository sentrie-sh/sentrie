// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import "github.com/sentrie-sh/sentrie/trinary"

// TestNowUnitsAgree is the regression for #153: the builtin now() and
// @sentrie/time.now() both report the execution start in Unix seconds, and
// now_ms() is the same instant in milliseconds.
func (s *RuntimeTestSuite) TestNowUnitsAgree() {
	ctx := s.T().Context()
	src := `namespace com/ex

policy pol {
  use { now } from @sentrie/time as time

  rule same_instant = { yield now() == time.now() }
  rule ms_matches = { yield now_ms() >= now() * 1000 and now_ms() < (now() + 1) * 1000 }
  rule stable = { yield now() == now() and now_ms() == now_ms() }

  export decision of same_instant
  export decision of ms_matches
  export decision of stable
}
`
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "now_units.sentrie", src: src})
	for _, rule := range []string{"same_instant", "ms_matches", "stable"} {
		s.Run(rule, func() {
			out, err := exec.ExecRule(ctx, "com/ex", "pol", rule, nil)
			s.Require().NoError(err)
			s.Equal(trinary.True, out.Decision.State)
		})
	}
}
