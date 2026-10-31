// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"github.com/sentrie-sh/sentrie/trinary"
)

const unionPositionsSource = `namespace com/app
shape Nums list[number]
shape Opt { id: string  label: string | number }
shape Wrap { items: list[string | number]  extra: any }

derive badReturn = (v: number): string | list[string] => { yield v }
derive okReturn = (v: number): string | number => { yield v }
derive listParam = (xs: list[string | number]): number => { yield count(xs) }
derive optParam = (v?: string | number): trinary => { yield v is defined }

policy pol {
  fact items?: list[string | number]
  fact table?: dict[number | string]
  fact nums?: Nums | string
  fact opt?: Opt
  fact wrap?: Wrap
  fact tri?: trinary | string
  fact nested?: list[list[number] | string]
  fact required: string | number
  rule got_items = { yield items is defined }
  rule got_table = { yield table is defined }
  rule got_nums = { yield nums is defined }
  rule got_opt = { yield opt is defined }
  rule got_wrap = { yield wrap is defined }
  rule got_tri = { yield tri is defined }
  rule got_nested = { yield nested is defined }
  rule got_required = { yield required is defined }

  rule let_ok = {
    let a: string | number = "s"
    let b: list[string | number] = [1, "a"]
    let c: any = null
    yield a == "s" and count(b) == 2 and c == null
  }
  rule let_bad = {
    let b: list[string | number] = [1, true]
    yield b
  }
  rule cast_ok = { yield (5 as string | number) == 5 and ("x" as number | string) == "x" }
  rule cast_bad = { yield true as string | number }
  rule cast_any = { yield (null as any) == null }
  rule derive_return_ok = { yield okReturn(3) == 3 }
  rule derive_return_bad = { yield badReturn(3) }
  rule derive_list_ok = { yield listParam([1, "a"]) == 2 }
  rule derive_list_bad = { yield listParam([1, true]) }
  rule derive_optional_omitted = { yield optParam() is not defined or optParam() == false }
  rule derive_optional_given = { yield optParam(1) }
  rule is_nested = { yield [[1], "a"] is list[list[number] | string] and not ([[1, "x"]] is list[list[number] | string]) }
  rule is_dict_union = { yield { "a": 1, "b": "x" } is dict[number | string] and [1] is not dict[number | string] }
  rule is_alias_member = { yield [1, 2] is Nums | string and "s" is Nums | string and not ([1, "x"] is Nums | string) }
  rule is_any_in_containers = { yield [1, "a", null] is list[any] and { "a": [1] } is dict[any] }
  rule is_not_union_false = { yield not (5 is not string | number) }
  rule is_not_union_true = { yield true is not string | number }
  rule is_nullable_any = { yield null is any? and 1 is any? }
  rule is_union_of_any = { yield 1 is any | string }
  rule is_single_member_order = { yield 1 is number | string and 1 is string | number }
  rule is_trinary_member = { yield true is trinary | string and "s" is trinary | string }
  rule is_not_broken_member = { yield "ab" is not number | string@minlength("q") }
  export decision of got_items
  export decision of got_table
  export decision of got_nums
  export decision of got_opt
  export decision of got_wrap
  export decision of got_tri
  export decision of got_nested
  export decision of got_required
  export decision of let_ok
  export decision of let_bad
  export decision of cast_ok
  export decision of cast_bad
  export decision of cast_any
  export decision of derive_return_ok
  export decision of derive_return_bad
  export decision of derive_list_ok
  export decision of derive_list_bad
  export decision of derive_optional_omitted
  export decision of derive_optional_given
  export decision of is_nested
  export decision of is_dict_union
  export decision of is_alias_member
  export decision of is_any_in_containers
  export decision of is_not_union_false
  export decision of is_not_union_true
  export decision of is_nullable_any
  export decision of is_union_of_any
  export decision of is_single_member_order
  export decision of is_trinary_member
  export decision of is_not_broken_member
}
`

// TestUnionAndAnyInEveryTypePosition runs unions and `any` through let
// annotations, casts, derive returns and parameters, and nested containers.
func (s *RuntimeTestSuite) TestUnionAndAnyInEveryTypePosition() {
	ctx := s.T().Context()
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "positions.sentrie", src: unionPositionsSource})
	facts := map[string]any{"required": "r"}
	for _, rule := range []string{
		"let_ok", "cast_ok", "cast_any", "derive_return_ok", "derive_list_ok",
		"derive_optional_given", "is_nested", "is_dict_union", "is_alias_member",
		"is_any_in_containers", "is_not_union_false", "is_not_union_true", "is_nullable_any",
		"is_union_of_any", "is_single_member_order", "is_trinary_member",
	} {
		s.Run(rule, func() {
			out, err := exec.ExecRule(ctx, "com/app", "pol", rule, facts)
			s.Require().NoError(err)
			s.Equal(trinary.True, out.Decision.State)
		})
	}

	errCases := []struct{ rule, msg string }{
		{"let_bad", "is not string | number"},
		{"cast_bad", "is not string | number"},
		{"derive_return_bad", "is not string | list[string]"},
		{"derive_list_bad", "is not string | number"},
		// the string member's constraint arguments are invalid: an error, never "is not" true
		{"is_not_broken_member", "expected number, got string"},
	}
	for _, tc := range errCases {
		s.Run(tc.rule, func() {
			_, err := exec.ExecRule(ctx, "com/app", "pol", tc.rule, facts)
			s.Require().Error(err)
			s.Contains(err.Error(), tc.msg)
		})
	}
}

// TestUnionAndAnyFactsNestedPositions covers fact validation where the union
// sits inside a container, a shape field, or an alias.
func (s *RuntimeTestSuite) TestUnionAndAnyFactsNestedPositions() {
	ctx := s.T().Context()
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "positions.sentrie", src: unionPositionsSource})

	ok := []struct {
		rule  string
		facts map[string]any
	}{
		// JSON-decoded values: numbers arrive as float64
		{"got_items", map[string]any{"items": []any{float64(1), "a", 2.5}}},
		{"got_items", map[string]any{"items": []any{}}},
		{"got_table", map[string]any{"table": map[string]any{"a": float64(1), "b": "x"}}},
		{"got_nums", map[string]any{"nums": []any{float64(1)}}},
		{"got_nums", map[string]any{"nums": "s"}},
		{"got_opt", map[string]any{"opt": map[string]any{"id": "x", "label": "t"}}},
		{"got_opt", map[string]any{"opt": map[string]any{"id": "x", "label": float64(2)}}},
		{"got_wrap", map[string]any{"wrap": map[string]any{"items": []any{"a"}, "extra": nil}}},
		{"got_wrap", map[string]any{"wrap": map[string]any{"items": []any{1}, "extra": map[string]any{"k": []any{1}}}}},
		{"got_tri", map[string]any{"tri": true}},
		{"got_tri", map[string]any{"tri": "s"}},
		{"got_nested", map[string]any{"nested": []any{[]any{float64(1)}, "a"}}},
	}
	for _, tc := range ok {
		tc.facts["required"] = 1
		out, err := exec.ExecRule(ctx, "com/app", "pol", tc.rule, tc.facts)
		s.Require().NoError(err, "%s %v", tc.rule, tc.facts)
		s.Equal(trinary.True, out.Decision.State, "%s %v", tc.rule, tc.facts)
	}

	bad := []struct {
		rule  string
		facts map[string]any
		msgs  []string
	}{
		{"got_items", map[string]any{"items": []any{1, true}}, []string{"is not string | number"}},
		{"got_items", map[string]any{"items": "x"}, []string{"is not an array"}},
		{"got_table", map[string]any{"table": []any{1}}, []string{"is not a dict"}},
		// the alias member's kind (list) accepts it, so its element failure is reported
		{"got_nums", map[string]any{"nums": []any{1, "x"}}, []string{"closest member Nums"}},
		{"got_nums", map[string]any{"nums": true}, []string{"is not Nums | string"}},
		{"got_opt", map[string]any{"opt": map[string]any{"id": "x", "label": true}}, []string{"is not string | number"}},
		{"got_wrap", map[string]any{"wrap": map[string]any{"items": []any{true}, "extra": 1}}, []string{"is not string | number"}},
		{"got_tri", map[string]any{"tri": 5}, []string{"is not trinary | string"}},
		{"got_nested", map[string]any{"nested": []any{[]any{"x"}}}, []string{"closest member list[number]"}},
		{"got_nested", map[string]any{"nested": []any{5}}, []string{"is not list[number] | string"}},
	}
	for _, tc := range bad {
		tc.facts["required"] = 1
		_, err := exec.ExecRule(ctx, "com/app", "pol", tc.rule, tc.facts)
		s.Require().Error(err, "%s %v", tc.rule, tc.facts)
		for _, msg := range tc.msgs {
			s.Contains(err.Error(), msg, "%s %v", tc.rule, tc.facts)
		}
	}

	s.Run("required union fact missing", func() {
		_, err := exec.ExecRule(ctx, "com/app", "pol", "got_required", map[string]any{})
		s.Require().Error(err)
	})
	s.Run("required union fact null is rejected", func() {
		_, err := exec.ExecRule(ctx, "com/app", "pol", "got_required", map[string]any{"required": nil})
		s.Require().Error(err)
	})
	s.Run("optional union fact omitted", func() {
		_, err := exec.ExecRule(ctx, "com/app", "pol", "got_items", map[string]any{"required": "r"})
		s.Require().NoError(err)
	})
}

// TestUnionErrorMessageIsDeterministic: the closest-member choice must not
// depend on map iteration or evaluation order.
func (s *RuntimeTestSuite) TestUnionErrorMessageIsDeterministic() {
	ctx := s.T().Context()
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "positions.sentrie", src: unionPositionsSource})
	facts := map[string]any{"required": 1, "nums": []any{1, "x"}}
	_, first := exec.ExecRule(ctx, "com/app", "pol", "got_nums", facts)
	s.Require().Error(first)
	for range 25 {
		_, err := exec.ExecRule(ctx, "com/app", "pol", "got_nums", facts)
		s.Require().Error(err)
		s.Equal(first.Error(), err.Error())
	}
}

// TestUnionBrokenMemberDoesNotFailOpen checks a union whose member cannot be
// resolved at runtime (index validation skipped) reports an error rather than
// treating the value as a mismatch, so `is not` cannot pass by accident.
func (s *RuntimeTestSuite) TestUnionBrokenMemberDoesNotFailOpen() {
	ctx := s.T().Context()
	src := `namespace com/app
policy pol {
  rule negated = { yield 5 is not Missing | string }
  rule plain = { yield 5 is Missing | string }
  rule matched_first = { yield "s" is string | Missing }
  export decision of negated
  export decision of plain
  export decision of matched_first
}
`
	idx := s.mustBuildDeriveIndex(ctx, deriveTestProgram{name: "broken.sentrie", src: src})
	rawExec, err := NewExecutor(idx)
	s.Require().NoError(err)
	for _, rule := range []string{"negated", "plain"} {
		s.Run(rule, func() {
			out, err := rawExec.ExecRule(ctx, "com/app", "pol", rule, nil)
			if err == nil {
				s.Failf("broken member must not yield a decision", "state %v", out.Decision.State)
			}
		})
	}
	s.Run("an earlier matching member wins before the broken one is reached", func() {
		out, err := rawExec.ExecRule(ctx, "com/app", "pol", "matched_first", nil)
		s.Require().NoError(err)
		s.Equal(trinary.True, out.Decision.State)
	})
}
