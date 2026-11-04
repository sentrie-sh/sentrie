// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"regexp"
	"strings"
)

// commentedPolicy has a comment in nearly every position the parser keeps
// one: before and after the namespace, shape, derive and policy, inside
// blocks, between statements, and trailing after code.
const commentedPolicy = `-- header
-- second header line

namespace cmt/eval -- after namespace

-- a shape
shape Point { -- open
  -- first field
  x: number -- after x
  y: number
}

derive twice = (n: number): number => { yield n * 2 }
-- after derive

-- the policy
policy p { -- open
  -- a fact
  fact n: number -- after fact
  let a = n + 1 -- trailing
  let b = [1, -- first
    2, 3] -- list
  let c = a -- left
    + -- operator
    b[0] -- right
  let d = c > 4
    ? "big" -- then
    : "small" -- else
  let e = n |> twice()
  let f = {
    -- inside a block expression
    let tmp = e + 1 -- tmp
    yield tmp
  }

  -- a rule
  rule big = { -- open
    -- inside the rule
    yield f > 6 and d == "big" -- yield
  } -- close
  export decision of big attach a as a -- export
  rule small = { yield f > 100 }
  export decision of small
  -- last
}
-- end
`

var lineComment = regexp.MustCompile(`[ \t]*--[^\n]*`)

// TestCommentsDoNotChangeEvaluation checks that a policy full of comments
// evaluates to the same decisions and attachments as the same policy with
// every comment removed.
func (s *RuntimeTestSuite) TestCommentsDoNotChangeEvaluation() {
	ctx := s.T().Context()
	stripped := lineComment.ReplaceAllString(commentedPolicy, "")
	s.NotContains(stripped, "--")
	s.NotEqual(commentedPolicy, stripped)

	_, withComments := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "commented.sentrie", src: commentedPolicy})
	_, without := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "plain.sentrie", src: stripped})

	for _, n := range []int{1, 3, 10} {
		for _, rule := range []string{"big", "small"} {
			facts := map[string]any{"n": n}
			a, errA := withComments.ExecRule(ctx, "cmt/eval", "p", rule, facts)
			b, errB := without.ExecRule(ctx, "cmt/eval", "p", rule, facts)
			s.Require().NoError(errA, rule)
			s.Require().NoError(errB, rule)
			s.Equal(b.Decision.State, a.Decision.State, "rule %s, n=%d", rule, n)
			s.Equal(b.Decision.Value.String(), a.Decision.Value.String(), "rule %s, n=%d", rule, n)
			s.Equal(len(b.Attachments), len(a.Attachments), "rule %s, n=%d", rule, n)
		}
	}
	// n=10: f = 10*2+1 = 21 > 6, c = 11+1 = 12 > 4 -> "big".
	out, err := withComments.ExecRule(ctx, "cmt/eval", "p", "big", map[string]any{"n": 10})
	s.Require().NoError(err)
	s.Equal("true", strings.ToLower(out.Decision.State.String()))
	s.Contains(out.Attachments, "a")
}
