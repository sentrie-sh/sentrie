// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"github.com/sentrie-sh/sentrie/trinary"
)

const stringOrderSource = `namespace str/order

policy p {
  fact at: string
  fact names: list[string]

  rule literal_before = "2024-01-01T00:00:00Z" < "2024-06-01T00:00:00Z"
  rule literal_after = "2024-01-01T00:00:00Z" > "2024-06-01T00:00:00Z"
  rule literal_equal_lte = "x" <= "x"
  rule literal_equal_gte = "x" >= "x"
  rule fact_in_window = at >= "2024-01-01T00:00:00Z" and at < "2025-01-01T00:00:00Z"
  rule fact_vs_fact = at < names[0]
  rule gate = when (at > "m") { yield true }
  rule pick = { yield at > "m" ? "late" : "early" }
  rule filtered = { yield count(filter(names, (n: string): trinary => { yield n >= "b" and n < "d" })) == 2 }
  rule mixed = { yield at < 1 }
  rule mixed_in_gate = when (1 < at) { yield true }

  export decision of literal_before
  export decision of literal_after
  export decision of literal_equal_lte
  export decision of literal_equal_gte
  export decision of fact_in_window
  export decision of fact_vs_fact
  export decision of gate
  export decision of pick
  export decision of filtered
  export decision of mixed
  export decision of mixed_in_gate
}
`

// TestStringOrderingEndToEndThroughExec runs the relational operators on strings
// through the full load, index and execute pipeline: as a rule body, against
// fact values, as a `when` gate, as a ternary condition and inside a lambda
// passed to filter. Mixed string and number operands must fail the rule.
func (s *RuntimeTestSuite) TestStringOrderingEndToEndThroughExec() {
	ctx := s.T().Context()
	exec := s.newTempPackExecutor(map[string]string{"order.sentrie": stringOrderSource})
	defer func() { s.NoError(exec.Close()) }()

	facts := func(at string) map[string]any {
		return map[string]any{"at": at, "names": []any{"a", "b", "c", "d"}}
	}

	s.Run("literals and facts", func() {
		cases := []struct {
			rule string
			at   string
			want trinary.Value
		}{
			{"literal_before", "z", trinary.True},
			{"literal_after", "z", trinary.False},
			{"literal_equal_lte", "z", trinary.True},
			{"literal_equal_gte", "z", trinary.True},
			{"fact_in_window", "2024-06-01T12:00:00Z", trinary.True},
			{"fact_in_window", "2023-12-31T23:59:59Z", trinary.False},
			{"fact_in_window", "2025-01-01T00:00:00Z", trinary.False},
			{"fact_vs_fact", "0", trinary.True},
			{"fact_vs_fact", "a", trinary.False},
			{"fact_vs_fact", "B", trinary.True},
			{"gate", "z", trinary.True},
			{"gate", "a", trinary.Unknown},
			{"filtered", "a", trinary.True},
		}
		for _, tc := range cases {
			out, err := exec.ExecRule(ctx, "str/order", "p", tc.rule, facts(tc.at))
			s.Require().NoError(err, "%s at=%q", tc.rule, tc.at)
			s.Equal(tc.want, out.Decision.State, "%s at=%q", tc.rule, tc.at)
		}
	})

	s.Run("ternary picks the branch by string order", func() {
		for at, want := range map[string]string{"z": "late", "a": "early", "m": "early", "n": "late"} {
			out, err := exec.ExecRule(ctx, "str/order", "p", "pick", facts(at))
			s.Require().NoError(err)
			s.Equal(want, out.Decision.Value.Any(), "at=%q", at)
		}
	})

	s.Run("mixed string and number fails the rule", func() {
		for _, rule := range []string{"mixed", "mixed_in_gate"} {
			_, err := exec.ExecRule(ctx, "str/order", "p", rule, facts("a"))
			s.Require().Error(err, rule)
			s.Contains(err.Error(), "not a number", rule)
		}
	})
}
