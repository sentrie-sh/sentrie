// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"strings"

	"github.com/sentrie-sh/sentrie/trinary"
)

// stringPolicyLF is a policy whose string literal spans lines, written with LF
// line endings; the rules compare it with the value a line break must have.
const stringPolicyLF = `namespace text/crlf

policy p {
  let s = "a
b"
  let t = "a\r
b"
  rule value = { yield s }
  rule equals_lf = { yield s == "a\nb" }
  rule has_no_cr = { yield not (s contains "\r") }
  rule escaped_cr_kept = { yield t == "a\r\nb" }
  export decision of value
  export decision of equals_lf
  export decision of has_no_cr
  export decision of escaped_cr_kept
}
`

// TestStringLiteralInACRLFFileEvaluatesLikeLF is #280 end to end: a pack file
// with CRLF line endings loads, and a string literal that spans lines is equal
// to the same text written with \n, as in an LF file. An escaped \r is kept.
func (s *RuntimeTestSuite) TestStringLiteralInACRLFFileEvaluatesLikeLF() {
	ctx := s.T().Context()
	for name, src := range map[string]string{
		"LF":   stringPolicyLF,
		"CRLF": strings.ReplaceAll(stringPolicyLF, "\n", "\r\n"),
	} {
		s.Run(name, func() {
			exec := s.newTempPackExecutor(map[string]string{"policy.sentrie": src})
			defer func() { s.NoError(exec.Close()) }()

			out, err := exec.ExecRule(ctx, "text/crlf", "p", "value", nil)
			s.Require().NoError(err)
			s.Equal("a\nb", out.Decision.Value.Any())

			for _, rule := range []string{"equals_lf", "has_no_cr", "escaped_cr_kept"} {
				out, err := exec.ExecRule(ctx, "text/crlf", "p", rule, nil)
				s.Require().NoError(err, rule)
				s.Equal(trinary.True, out.Decision.State, rule)
			}
		})
	}
}
