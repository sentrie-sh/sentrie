// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"github.com/sentrie-sh/sentrie/trinary"
)

// TestExecRuleUsesStringAndMathModules evaluates rules that import
// @sentrie/string and @sentrie/math from policy code (#158), including facts
// that arrive as lists, and checks that a module error fails the rule rather
// than flowing on as a value.
func (s *RuntimeTestSuite) TestExecRuleUsesStringAndMathModules() {
	ctx := s.T().Context()
	src := `namespace com/ex

policy pol {
  fact path: string
  fact role: string
  fact roles: list[string]
  fact scores: list[number]

  use { lower, startsWith, equalFold, join, split, format } from @sentrie/string as str
  use { max, clamp, round } from @sentrie/math

  rule prefix_ok = default false {
    yield str.startsWith(str.lower(path), "/api/")
  }
  rule fold_ok = default false {
    yield str.equalFold(role, "ADMIN")
  }
  rule join_ok = default false {
    yield str.join(roles, ",") == "reader,writer"
  }
  rule split_ok = default false {
    yield count(str.split(path, "/")) == 4
  }
  rule format_ok = default false {
    yield str.format("%s scored %d", role, math.max(scores)) == "Admin scored 9"
  }
  rule math_ok = default false {
    yield math.clamp(math.round(math.max(scores) * 1.5), 0, 10) == 10
  }
  rule throws = default false {
    yield str.lower(math.max(scores)) == "9"
  }

  export decision of prefix_ok
  export decision of fold_ok
  export decision of join_ok
  export decision of split_ok
  export decision of format_ok
  export decision of math_ok
  export decision of throws
}
`
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "string_math.sentrie", src: src})
	facts := map[string]any{
		"path":   "/API/v1/users",
		"role":   "Admin",
		"roles":  []any{"reader", "writer"},
		"scores": []any{4, 9, 2.5},
	}

	for _, ruleName := range []string{"prefix_ok", "fold_ok", "join_ok", "split_ok", "format_ok", "math_ok"} {
		s.Run(ruleName, func() {
			out, err := exec.ExecRule(ctx, "com/ex", "pol", ruleName, facts)
			s.Require().NoError(err)
			s.Equal(trinary.True, out.Decision.State)
		})
	}

	_, err := exec.ExecRule(ctx, "com/ex", "pol", "throws", facts)
	s.Require().Error(err)
	s.Contains(err.Error(), "lower: argument 1 must be a string, got number")
}
