// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"github.com/sentrie-sh/sentrie/trinary"
)

func (s *RuntimeTestSuite) TestEvalIsDefinedAndIsEmptyEndToEnd() {
	ctx := s.T().Context()
	src := `namespace com/ex

derive blank = (v?: string): trinary => { yield v is empty }

policy pol {
  fact user: document as user
  fact extra?: document as extra
  rule has_x = { yield user.x is defined }
  rule no_y = { yield user.y is not defined }
  rule null_defined = { yield user.n is defined }
  rule no_extra = { yield extra is not defined and extra.k is not defined }
  rule empty_list = { yield [] is empty }
  rule full_list = { yield [1] is not empty }
  rule null_empty = { yield user.n is empty }
  rule missing_empty = { yield user.y is empty }
  rule zero_not_empty = { yield 0 is not empty }
  rule derive_blank = { yield blank("") and blank() and not blank("a") }
  export decision of has_x
  export decision of no_y
  export decision of null_defined
  export decision of no_extra
  export decision of empty_list
  export decision of full_list
  export decision of null_empty
  export decision of missing_empty
  export decision of zero_not_empty
  export decision of derive_blank
}
`
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "is_ops.sentrie", src: src})
	facts := map[string]any{"user": map[string]any{"x": 1, "n": nil}}

	for _, rule := range []string{"has_x", "no_y", "null_defined", "no_extra", "empty_list", "full_list", "null_empty", "missing_empty", "zero_not_empty", "derive_blank"} {
		s.Run(rule, func() {
			out, err := exec.ExecRule(ctx, "com/ex", "pol", rule, facts)
			s.Require().NoError(err)
			s.Equal(trinary.True, out.Decision.State)
		})
	}
}
