// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"time"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/trinary"
)

const memoClockSrc = `namespace com/ex

policy pol {
  use { now } from @sentrie/time as time
  use { verify, getPayload } from @sentrie/jwt as jwt
  use { sha256 } from @sentrie/hash as hash

  rule memo_now = { yield now()!60 == now() }
  rule plain = { yield 1 }

  export decision of memo_now
  export decision of plain
}
`

// TestMemoizedClockCallsDoNotCrossExecutions is the regression for the
// PR #197 review: the memo cache is executor-wide, so a memoized clock read
// (or a JWT check that depends on it) must not hand one execution's answer to
// the next.
func (s *RuntimeTestSuite) TestMemoizedClockCallsDoNotCrossExecutions() {
	ctx := s.T().Context()
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "memo_clock.sentrie", src: memoClockSrc})

	// The memoized read must equal the plain one in every execution. Without
	// the fix the second execution gets the first one's clock; the sleep makes
	// that observable while now() is in milliseconds, and the assertion stays
	// true whatever unit now() uses.
	for i := range 2 {
		if i > 0 {
			time.Sleep(5 * time.Millisecond)
		}
		out, err := exec.ExecRule(ctx, "com/ex", "pol", "memo_now", nil)
		s.Require().NoError(err)
		s.Equalf(trinary.True, out.Decision.State, "execution %d: a memoized now() returned an earlier execution's clock", i+1)
	}
}

func (s *RuntimeTestSuite) TestIsClockDependentCall() {
	ctx := s.T().Context()
	idx, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "memo_clock.sentrie", src: memoClockSrc})
	policy, err := idx.ResolvePolicy("com/ex", "pol")
	s.Require().NoError(err)
	ec := NewExecutionContext(policy, exec)
	s.Require().NoError(exec.bindUses(ctx, ec, policy))

	call := func(callee ast.Expression) *ast.CallExpression {
		return ast.NewCallExpression(callee, nil, true, nil, stubRange())
	}
	field := func(alias, fn string) ast.Expression {
		return ast.NewFieldAccessExpression(ast.NewIdentifier(alias, stubRange()), fn, stubRange())
	}
	s.True(isClockDependentCall(ec, policy, call(ast.NewIdentifier("now", stubRange()))))
	s.True(isClockDependentCall(ec, policy, call(field("time", "now"))))
	s.True(isClockDependentCall(ec, policy, call(field("jwt", "verify"))))
	s.False(isClockDependentCall(ec, policy, call(field("jwt", "getPayload"))))
	s.False(isClockDependentCall(ec, policy, call(field("hash", "sha256"))))
	s.False(isClockDependentCall(ec, policy, call(ast.NewIdentifier("count", stubRange()))))
}
