// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"github.com/sentrie-sh/sentrie/trinary"
)

// TestPolicyDeriveAsFirstStatementIsCallable is the #300 reproduction through
// the runtime: a policy whose first body statement is a derive validates, and
// a rule that calls the derive evaluates, by ExecRule and by ExecPolicy.
func (s *RuntimeTestSuite) TestPolicyDeriveAsFirstStatementIsCallable() {
	ctx := s.T().Context()
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{
		name: "derive_first.sentrie",
		src: `namespace t/a

policy p {
  derive n = (t: string): number => {
    yield count(t)
  }
  rule ok = { yield n("x") == 1 }
  rule no = { yield n("xy") == 1 }
  export decision of ok
  export decision of no
}
`,
	})
	defer func() { s.NoError(exec.Close()) }()

	out, err := exec.ExecRule(ctx, "t/a", "p", "ok", nil)
	s.Require().NoError(err)
	s.Equal(trinary.True, out.Decision.State)

	out, err = exec.ExecRule(ctx, "t/a", "p", "no", nil)
	s.Require().NoError(err)
	s.Equal(trinary.False, out.Decision.State)

	all, err := exec.ExecPolicy(ctx, "t/a", "p", nil)
	s.Require().NoError(err)
	s.Len(all, 2)
}

// TestPolicyDeriveAfterFactReadsFact covers a derive that follows the header
// statements and is called with a fact value, so the derive body is resolved
// in the policy scope after the header phases.
func (s *RuntimeTestSuite) TestPolicyDeriveAfterFactReadsFact() {
	ctx := s.T().Context()
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{
		name: "derive_after_fact.sentrie",
		src: `namespace t/a

policy p {
  fact name: string
  derive long = (v: string): trinary => { yield count(v) > 3 }
  rule ok = { yield long(name) }
  export decision of ok
}
`,
	})
	defer func() { s.NoError(exec.Close()) }()

	out, err := exec.ExecRule(ctx, "t/a", "p", "ok", map[string]any{"name": "abcd"})
	s.Require().NoError(err)
	s.Equal(trinary.True, out.Decision.State)

	out, err = exec.ExecRule(ctx, "t/a", "p", "ok", map[string]any{"name": "ab"})
	s.Require().NoError(err)
	s.Equal(trinary.False, out.Decision.State)
}
