// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import "github.com/sentrie-sh/sentrie/trinary"

// TestJSBuiltinFailuresFailClosed is the policy-level regression for #111.
// A builtin used to return its failure as an Error object, which crossed back
// into Sentrie as a truthy document: a malformed CIDR allowed and a malformed
// token decoded. Both must now fail the rule instead.
func (s *RuntimeTestSuite) TestJSBuiltinFailuresFailClosed() {
	ctx := s.T().Context()
	src := `namespace com/ex

policy pol {
  use { cidrContains } from @sentrie/net as net
  use { decode } from @sentrie/jwt as jwt

  rule n_ok = { yield net.cidrContains("10.0.0.0/8", "10.1.2.3") ? "allowed" : "denied" }
  rule n1 = { yield net.cidrContains("not-a-cidr", "1.2.3.4") ? "allowed" : "denied" }
  rule j1 = { yield jwt.decode("x.y.z") }
  rule j2 = { yield jwt.decode("x.y.z", "secret") }
  let cfg = {"issuer": "example"}
  rule j_missing_key = { yield jwt.decode("eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiJhZG1pbiJ9.AAAA", cfg.secret) }

  export decision of n_ok
  export decision of n1
  export decision of j1
  export decision of j2
  export decision of j_missing_key
}
`
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "js_failures.sentrie", src: src})

	out, err := exec.ExecRule(ctx, "com/ex", "pol", "n_ok", nil)
	s.Require().NoError(err)
	s.Equal(trinary.True, out.Decision.State)

	for rule, msg := range map[string]string{
		"n1": "invalid CIDR address",
		"j1": "decode requires 2 or 3 arguments",
		"j2": "illegal base64 data",
		// An unset secret reaches JavaScript as undefined; decode must not
		// treat that as "skip verification" (PR #197 review).
		"j_missing_key": "decode requires a key",
	} {
		s.Run(rule, func() {
			out, err := exec.ExecRule(ctx, "com/ex", "pol", rule, nil)
			s.Require().Errorf(err, "rule %s must fail, got %+v", rule, out)
			s.Contains(err.Error(), msg)
		})
	}
}
