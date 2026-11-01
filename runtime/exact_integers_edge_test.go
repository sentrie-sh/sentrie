// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"bytes"
	"encoding/json"

	"github.com/sentrie-sh/sentrie/box"
)

// exactIntegersEdgePolicy covers #166 edge cases the end-to-end policy does
// not: relational operators, int and float equality in collections, overflow
// at both ends, JSON rendering of floats, number constraints and memoization.
// Facts: a = 2^53+1, b = 2^53.
const exactIntegersEdgePolicy = `namespace exact/edge

shape Id number@min(9007199254740993)@max(9007199254740995)

derive ident = (x: number): number => { yield x }

policy p {
  fact a: number
  fact b: number
  fact constrained: Id

  rule lt_ab = { yield a < b }
  rule le_ab = { yield a <= b }
  rule gt_ab = { yield a > b }
  rule ge_ba = { yield b >= a }
  rule ge_ab = { yield a >= b }
  rule int_gt_float = { yield 9007199254740993 > 9007199254740992.0 }
  rule int_lt_float = { yield 9007199254740993 < 9007199254740994.0 }
  rule int_eq_float_exact = { yield 9007199254740992 == 9007199254740992.0 }
  rule int_eq_float_rounded = { yield 9007199254740993 == 9007199254740992.0 }
  rule int_ne_float_rounded = { yield 9007199254740993 != 9007199254740992.0 }
  rule max_gt_float = { yield 9223372036854775807 < 9223372036854775808.0 }
  rule zero_eq_negative_zero = { yield 0 == -0.0 }

  rule in_int_float = { yield 5 in [5.0, 6] }
  rule in_neighbour = { yield a in [b] }
  rule in_exact = { yield a in [b, a] }
  rule contains_neighbour = { yield [a] contains b }
  rule contains_float = { yield [b, 2.0] contains 2 }
  rule dict_int_float = { yield { "k": 1 } == { "k": 1.0 } }
  rule dict_neighbour = { yield { "k": a } == { "k": b } }
  rule list_neighbour = { yield [a] == [b] }
  rule distinct_mixed = { yield distinct([a, b, 2, 2.0, a]) }

  rule add_max = { yield 9223372036854775806 + 1 }
  rule sub_min = { yield -9223372036854775807 - 1 }
  rule mul_min = { yield -4611686018427387904 * 2 }
  rule add_overflow = { yield 9223372036854775807 + 1 }
  rule sub_overflow = { yield -9223372036854775807 - 2 }
  rule mul_overflow = { yield 4611686018427387904 * 2 }
  rule mul_negative_overflow = { yield -4611686018427387905 * 2 }
  rule self_difference = { yield a - a }
  rule exact_half = { yield 9007199254740992 / 2 }
  rule modulo = { yield 9007199254740993 % 10 }
  rule negative_modulo = { yield -7 % 3 }
  rule promoted = { yield 9007199254740993 + 0.0 }

  rule float_integral = { yield 2.0 }
  rule float_fraction = { yield 1.5 }
  rule negative_zero = { yield -0.0 }

  rule constrained_echo = { yield constrained }

  rule memo_distinct_ids = { yield ident(a)!60 == a and ident(b)!60 == b }
  rule memo_large_ttl = { yield ident(1)!9223372036854775807 }

  export decision of lt_ab
  export decision of le_ab
  export decision of gt_ab
  export decision of ge_ba
  export decision of ge_ab
  export decision of int_gt_float
  export decision of int_lt_float
  export decision of int_eq_float_exact
  export decision of int_eq_float_rounded
  export decision of int_ne_float_rounded
  export decision of max_gt_float
  export decision of zero_eq_negative_zero
  export decision of in_int_float
  export decision of in_neighbour
  export decision of in_exact
  export decision of contains_neighbour
  export decision of contains_float
  export decision of dict_int_float
  export decision of dict_neighbour
  export decision of list_neighbour
  export decision of distinct_mixed
  export decision of add_max
  export decision of sub_min
  export decision of mul_min
  export decision of add_overflow
  export decision of sub_overflow
  export decision of mul_overflow
  export decision of mul_negative_overflow
  export decision of self_difference
  export decision of exact_half
  export decision of modulo
  export decision of negative_modulo
  export decision of promoted
  export decision of float_integral
  export decision of float_fraction
  export decision of negative_zero
  export decision of constrained_echo
  export decision of memo_distinct_ids
  export decision of memo_large_ttl
}
`

func (s *RuntimeTestSuite) decodeEdgeFacts(src string) map[string]any {
	var facts map[string]any
	dec := json.NewDecoder(bytes.NewReader([]byte(src)))
	dec.UseNumber()
	s.Require().NoError(dec.Decode(&facts))
	return facts
}

func (s *RuntimeTestSuite) TestExactIntegerEdgeCases() {
	ctx := s.T().Context()
	exec := s.newTempPackExecutor(map[string]string{"policy.sentrie": exactIntegersEdgePolicy})
	defer func() { s.NoError(exec.Close()) }()
	facts := s.decodeEdgeFacts(`{"a": 9007199254740993, "b": 9007199254740992, "constrained": 9007199254740994}`)

	render := func(rule string) string {
		out, err := exec.ExecRule(ctx, "exact/edge", "p", rule, facts)
		s.Require().NoError(err, rule)
		b, err := json.Marshal(out.Decision.Value)
		s.Require().NoError(err, rule)
		return string(b)
	}

	want := map[string]string{
		// relational operators order neighbours above 2^53
		"lt_ab": "false", "le_ab": "false", "gt_ab": "true", "ge_ba": "false", "ge_ab": "true",
		// an integer is compared with a float without rounding the integer
		"int_gt_float": "true", "int_lt_float": "true", "int_eq_float_exact": "true",
		"int_eq_float_rounded": "false", "int_ne_float_rounded": "true", "max_gt_float": "true",
		"zero_eq_negative_zero": "true",
		// integer and float equality inside collections
		"in_int_float": "true", "in_neighbour": "false", "in_exact": "true",
		"contains_neighbour": "false", "contains_float": "true",
		"dict_int_float": "true", "dict_neighbour": "false", "list_neighbour": "false",
		"distinct_mixed": "[9007199254740993,9007199254740992,2]",
		// arithmetic reaches both ends of int64 without overflowing
		"add_max": "9223372036854775807", "sub_min": "-9223372036854775808", "mul_min": "-9223372036854775808",
		"self_difference": "0", "exact_half": "4503599627370496", "modulo": "3", "negative_modulo": "-1",
		// mixed arithmetic promotes to float
		"promoted": "9007199254740992",
		// JSON rendering of floats is stable
		"float_integral": "2", "float_fraction": "1.5", "negative_zero": "-0",
		// number constraints on a large fact
		"constrained_echo": "9007199254740994",
		// memoization keys distinct large integers apart
		"memo_distinct_ids": "true", "memo_large_ttl": "1",
	}
	for rule, expected := range want {
		s.Equal(expected, render(rule), rule)
	}

	for _, rule := range []string{"add_overflow", "sub_overflow", "mul_overflow", "mul_negative_overflow"} {
		_, err := exec.ExecRule(ctx, "exact/edge", "p", rule, facts)
		s.Require().Error(err, rule)
		s.ErrorIs(err, box.ErrIntegerOverflow, rule)
	}
}

func (s *RuntimeTestSuite) TestExactIntegerShapeConstraintBounds() {
	ctx := s.T().Context()
	exec := s.newTempPackExecutor(map[string]string{"policy.sentrie": exactIntegersEdgePolicy})
	defer func() { s.NoError(exec.Close()) }()

	run := func(id string) error {
		facts := s.decodeEdgeFacts(`{"a": 1, "b": 2, "constrained": ` + id + `}`)
		_, err := exec.ExecRule(ctx, "exact/edge", "p", "constrained_echo", facts)
		return err
	}
	s.Require().NoError(run("9007199254740993"), "lower bound is inclusive")
	s.Require().NoError(run("9007199254740995"), "upper bound is inclusive")
	s.Require().Error(run("9007199254740992"), "one below the lower bound must fail, float64 would round it up")
	s.Require().Error(run("9007199254740996"), "one above the upper bound must fail")
}

// TestExactIntegerCrossingIntoJSEdges covers the recorded #166 decision that an
// integer outside the JavaScript safe range raises an error at the module
// boundary and is never rounded.
func (s *RuntimeTestSuite) TestExactIntegerCrossingIntoJSEdges() {
	ctx := s.T().Context()
	exec := s.newTempPackExecutor(map[string]string{
		"echo.ts": `export function echo(x: number): number { return x; }
export function first(xs: number[]): number { return xs[0]; }
export function pick(d: any): number { return d.inner.id; }
`,
		"policy.sentrie": `namespace exact/jsedge

policy p {
  fact id: number

  use { echo, first, pick } from "./echo" as m

  rule safe_negative = { yield m.echo(-9007199254740991) }
  rule unsafe_negative = { yield m.echo(-9007199254740993) }
  rule unsafe_literal_arith = { yield m.echo(9007199254740991 + 2) }
  rule unsafe_fact = { yield m.echo(id) }
  rule safe_in_list = { yield m.first([7, 9007199254740991]) }
  rule unsafe_in_list = { yield m.first([7, 9007199254740993]) }
  rule unsafe_deep_dict = { yield m.pick({ "inner": { "id": 9007199254740993 } }) }
  rule float_stays_float = { yield m.echo(9007199254740994.0 * 4096.0) }

  export decision of safe_negative
  export decision of unsafe_negative
  export decision of unsafe_literal_arith
  export decision of unsafe_fact
  export decision of safe_in_list
  export decision of unsafe_in_list
  export decision of unsafe_deep_dict
  export decision of float_stays_float
}
`,
	})
	defer func() { s.NoError(exec.Close()) }()
	bigFact := s.decodeEdgeFacts(`{"id": 9007199254740993}`)
	smallFact := s.decodeEdgeFacts(`{"id": 12}`)

	out, err := exec.ExecRule(ctx, "exact/jsedge", "p", "safe_negative", smallFact)
	s.Require().NoError(err)
	i, ok := out.Decision.Value.IntValue()
	s.True(ok)
	s.Equal(int64(-9007199254740991), i)

	out, err = exec.ExecRule(ctx, "exact/jsedge", "p", "safe_in_list", smallFact)
	s.Require().NoError(err)
	i, ok = out.Decision.Value.IntValue()
	s.True(ok)
	s.Equal(int64(7), i)

	for _, rule := range []string{"unsafe_negative", "unsafe_literal_arith", "unsafe_in_list", "unsafe_deep_dict"} {
		_, err = exec.ExecRule(ctx, "exact/jsedge", "p", rule, smallFact)
		s.Require().Error(err, rule)
		s.Contains(err.Error(), "outside the JavaScript safe integer range", rule)
	}

	_, err = exec.ExecRule(ctx, "exact/jsedge", "p", "unsafe_fact", bigFact)
	s.Require().Error(err, "a large fact passed straight to a module must not be rounded")
	s.Contains(err.Error(), "outside the JavaScript safe integer range")

	out, err = exec.ExecRule(ctx, "exact/jsedge", "p", "unsafe_fact", smallFact)
	s.Require().NoError(err)
	i, ok = out.Decision.Value.IntValue()
	s.True(ok)
	s.Equal(int64(12), i)

	// A float is a double already, so it crosses unchanged even above 2^53.
	_, err = exec.ExecRule(ctx, "exact/jsedge", "p", "float_stays_float", smallFact)
	s.Require().NoError(err)
}
