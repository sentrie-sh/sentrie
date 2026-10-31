// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"strings"

	"github.com/sentrie-sh/sentrie/trinary"
)

// heredocPolicyLF is a policy whose heredocs, written with LF line endings,
// have the values the rules below compare them with.
const heredocPolicyLF = `namespace heredoc/crlf

policy p {
  let s = <<<EOT
hello
  world

EOT
  let t = <<<TWO
` + "second  \n" + "TWO  \n" + `  rule value = { yield s }
  rule equals_lf_string = { yield s == "hello\n  world\n\n" }
  rule has_no_cr = { yield not (s contains "\r") }
  rule second_value = { yield t }
  rule second_keeps_trailing_spaces = { yield t == "second  \n" }
  export decision of value
  export decision of equals_lf_string
  export decision of has_no_cr
  export decision of second_value
  export decision of second_keeps_trailing_spaces
}
`

// TestHereDocInACRLFFileEvaluatesLikeLF is #270 end to end: a pack file with
// CRLF line endings loads, and its heredocs evaluate to the same LF-ended
// values, equal to the same text written as a string, as in an LF file.
func (s *RuntimeTestSuite) TestHereDocInACRLFFileEvaluatesLikeLF() {
	ctx := s.T().Context()
	for name, src := range map[string]string{
		"LF":   heredocPolicyLF,
		"CRLF": strings.ReplaceAll(heredocPolicyLF, "\n", "\r\n"),
	} {
		s.Run(name, func() {
			exec := s.newTempPackExecutor(map[string]string{"policy.sentrie": src})
			defer func() { s.NoError(exec.Close()) }()

			for rule, want := range map[string]string{
				"value":        "hello\n  world\n\n",
				"second_value": "second  \n",
			} {
				out, err := exec.ExecRule(ctx, "heredoc/crlf", "p", rule, nil)
				s.Require().NoError(err, rule)
				s.Equal(want, out.Decision.Value.Any(), rule)
			}
			for _, rule := range []string{"equals_lf_string", "has_no_cr", "second_keeps_trailing_spaces"} {
				out, err := exec.ExecRule(ctx, "heredoc/crlf", "p", rule, nil)
				s.Require().NoError(err, rule)
				s.Equal(trinary.True, out.Decision.State, rule)
			}
		})
	}
}
