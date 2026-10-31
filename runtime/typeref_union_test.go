// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"github.com/sentrie-sh/sentrie/trinary"
)

const unionSource = `namespace com/app
shape Item { sku: string }
shape Tagged { label: string }
shape Scalar string | number
shape Holder { v: number | list[number] }

derive pass = (v: string | number): string | number => { yield v }
derive anyOut = (v: any): any => { yield v }

policy pol {
  fact rhs?: number | list[number]
  fact top?: any
  fact maybe?: string | number?
  fact bound?: string@minlength(3) | number@min(10)
  fact held?: Holder
  fact scalar?: Scalar
  rule got_rhs = { yield rhs is defined }
  rule got_top = { yield top is defined }
  rule got_maybe = { yield maybe is defined }
  rule got_bound = { yield bound is defined }
  rule got_held = { yield held is defined }
  rule got_scalar = { yield scalar is defined }
  rule is_union = { yield 5 is string | number and "x" is string | number and not (true is string | number) }
  rule is_list_member = { yield [1, 2] is number | list[number] and not ([1, "a"] is number | list[number]) }
  rule is_not_union = { yield "x" is not number | list[number] }
  rule is_shape_union = { yield { "label": "a" } is Item | Tagged and not ({ "x": 1 } is Item | Tagged) }
  rule is_nullable_member = { yield null is string | number? and not (null is string | number) }
  rule is_constrained_member = { yield 12 is string@minlength(3) | number@min(10) and not (5 is string@minlength(3) | number@min(10)) }
  rule is_any = { yield 1 is any and "s" is any and null is any and [1] is any and { "a": 1 } is any }
  rule is_any_undefined = { yield not ({ "a": 1 }.b is any) }
  rule is_union_ternary = { yield (5 is string | number ? 1 : 2) == 1 }
  rule quantifier = { yield any([1, 2, 3], (n) => { yield n > 2 }) }
  rule derive_union = { yield pass(5) == 5 and pass("a") == "a" }
  rule derive_bad_param = { yield pass(true) }
  rule derive_any = { yield anyOut(null) == null and anyOut(3) == 3 }
  rule let_union_bad = {
    let x: string | number = true
    yield x
  }
  rule broken_member = { yield "ab" is number | string@minlength("q") }
  export decision of got_rhs
  export decision of got_top
  export decision of got_maybe
  export decision of got_bound
  export decision of got_held
  export decision of got_scalar
  export decision of is_union
  export decision of is_list_member
  export decision of is_not_union
  export decision of is_shape_union
  export decision of is_nullable_member
  export decision of is_constrained_member
  export decision of is_any
  export decision of is_any_undefined
  export decision of is_union_ternary
  export decision of quantifier
  export decision of derive_union
  export decision of derive_bad_param
  export decision of derive_any
  export decision of let_union_bad
  export decision of broken_member
}
`

// TestUnionAndAnyInExpressions covers `is` with unions and `any`, typed
// derive parameters and returns, and the `any` quantifier beside the type.
func (s *RuntimeTestSuite) TestUnionAndAnyInExpressions() {
	ctx := s.T().Context()
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "union.sentrie", src: unionSource})
	for _, rule := range []string{
		"is_union", "is_list_member", "is_not_union", "is_shape_union", "is_nullable_member",
		"is_constrained_member", "is_any", "is_any_undefined", "is_union_ternary", "quantifier",
		"derive_union", "derive_any",
	} {
		s.Run(rule, func() {
			out, err := exec.ExecRule(ctx, "com/app", "pol", rule, nil)
			s.Require().NoError(err)
			s.Equal(trinary.True, out.Decision.State)
		})
	}
	errCases := map[string]string{
		"derive_bad_param": "is not string | number",
		"let_union_bad":    "is not string | number",
		// a member whose check cannot be made is an error, not a mismatch
		"broken_member": "expected number, got string",
	}
	for rule, msg := range errCases {
		s.Run(rule, func() {
			_, err := exec.ExecRule(ctx, "com/app", "pol", rule, nil)
			s.Require().Error(err)
			s.Contains(err.Error(), msg)
		})
	}
}

// TestUnionAndAnyFacts covers fact validation against unions and `any`,
// including the closest-member error when no member matches.
func (s *RuntimeTestSuite) TestUnionAndAnyFacts() {
	ctx := s.T().Context()
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "union.sentrie", src: unionSource})
	ok := []struct {
		rule  string
		facts map[string]any
	}{
		{"got_rhs", map[string]any{"rhs": 5}},
		{"got_rhs", map[string]any{"rhs": []any{1, 2}}},
		{"got_top", map[string]any{"top": nil}},
		{"got_top", map[string]any{"top": "s"}},
		{"got_top", map[string]any{"top": 1.5}},
		{"got_top", map[string]any{"top": []any{1, "a"}}},
		{"got_top", map[string]any{"top": map[string]any{"a": 1}}},
		{"got_maybe", map[string]any{"maybe": nil}},
		{"got_maybe", map[string]any{"maybe": "s"}},
		{"got_bound", map[string]any{"bound": "abc"}},
		{"got_bound", map[string]any{"bound": 12}},
		{"got_held", map[string]any{"held": map[string]any{"v": []any{1}}}},
		{"got_scalar", map[string]any{"scalar": "s"}},
	}
	for _, tc := range ok {
		out, err := exec.ExecRule(ctx, "com/app", "pol", tc.rule, tc.facts)
		s.Require().NoError(err, "%v", tc.facts)
		s.Equal(trinary.True, out.Decision.State, "%v", tc.facts)
	}

	bad := []struct {
		rule  string
		facts map[string]any
		msgs  []string
	}{
		// no member's kind accepts a string: the union is reported
		{"got_rhs", map[string]any{"rhs": "x"}, []string{"is not number | list[number]"}},
		// the list member's kind accepts it: its item failure is reported
		{"got_rhs", map[string]any{"rhs": []any{1, "a"}}, []string{"closest member list[number]", "is not a number"}},
		{"got_maybe", map[string]any{"maybe": true}, []string{"is not string | number?"}},
		// constraints bind to their member, and the closest member's fails
		{"got_bound", map[string]any{"bound": "ab"}, []string{"closest member string", "minlength"}},
		{"got_bound", map[string]any{"bound": 5}, []string{"closest member number", "'min'"}},
		{"got_held", map[string]any{"held": map[string]any{"v": "x"}}, []string{"field 'v'", "is not number | list[number]"}},
		{"got_scalar", map[string]any{"scalar": true}, []string{"is not string | number"}},
		{"got_bound", map[string]any{"bound": nil}, []string{"cannot be null"}},
	}
	for _, tc := range bad {
		_, err := exec.ExecRule(ctx, "com/app", "pol", tc.rule, tc.facts)
		s.Require().Error(err, "%v", tc.facts)
		for _, msg := range tc.msgs {
			s.Contains(err.Error(), msg, "%v", tc.facts)
		}
	}
}
