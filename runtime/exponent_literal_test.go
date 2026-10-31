// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"github.com/sentrie-sh/sentrie/trinary"
)

// TestExecRuleEvaluatesExponentLiterals runs rules written with exponent
// number literals (#180) end to end: an exponent literal is the same number
// its decimal spelling is, behaves as one operand in arithmetic and unary
// minus, works inside list and map literals, and does not disturb a field
// whose name looks like an exponent (`x.e5`).
func (s *RuntimeTestSuite) TestExecRuleEvaluatesExponentLiterals() {
	ctx := s.T().Context()
	src := `namespace com/ex

policy pol {
  fact x: document

  rule int_equiv = default false { yield 1e5 == 100000 }
  rule upper_marker = default false { yield 1E3 == 1000 }
  rule signed_plus = default false { yield 1e+2 == 100 }
  rule negative_exp = default false { yield 2.5e-1 + 1.5e1 == 15.25 }
  rule sub_not_exp = default false { yield 1e-1 - 1 == -0.9 }
  rule unary_minus = default false { yield -1e3 == -1000 }
  rule precedence = default false { yield 2 + 3e1 * 2 == 62 }
  rule in_list = default false { yield [1e1, 2e0][0] + [1e1, 2e0][1] == 12 }
  rule in_map = default false { yield { "k": 5e-1 }.k == 0.5 }
  rule field_named_like_exp = default false { yield x.e5 == 7 }
  rule not_equal = default false { yield 1e2 == 101 }

  export decision of int_equiv
  export decision of upper_marker
  export decision of signed_plus
  export decision of negative_exp
  export decision of sub_not_exp
  export decision of unary_minus
  export decision of precedence
  export decision of in_list
  export decision of in_map
  export decision of field_named_like_exp
  export decision of not_equal
}
`
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "exponent.sentrie", src: src})
	facts := map[string]any{"x": map[string]any{"e5": 7}}

	for _, ruleName := range []string{
		"int_equiv", "upper_marker", "signed_plus", "negative_exp", "sub_not_exp",
		"unary_minus", "precedence", "in_list", "in_map", "field_named_like_exp",
	} {
		s.Run(ruleName, func() {
			out, err := exec.ExecRule(ctx, "com/ex", "pol", ruleName, facts)
			s.Require().NoError(err)
			s.Equal(trinary.True, out.Decision.State)
		})
	}

	// A false comparison stays false, so the passing rules above are not
	// passing vacuously.
	out, err := exec.ExecRule(ctx, "com/ex", "pol", "not_equal", facts)
	s.Require().NoError(err)
	s.Equal(trinary.False, out.Decision.State)
}
