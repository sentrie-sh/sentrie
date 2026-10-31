// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/sentrie-sh/sentrie/trinary"
)

// ruleJSON runs a rule and returns its decision value as JSON.
func (s *RuntimeTestSuite) ruleJSON(exec *executorImpl, ns, policy, rule string, facts map[string]any) string {
	out, err := exec.ExecRule(s.T().Context(), ns, policy, rule, facts)
	s.Require().NoError(err, rule)
	raw, err := json.Marshal(out.Decision.Value)
	s.Require().NoError(err, rule)
	return string(raw)
}

// TestSetOutputIsStableAcrossInputOrders: the same elements, however the input
// orders them and however many times the rule runs, serialise to one sorted
// array, with mixed kinds in a fixed kind order.
func (s *RuntimeTestSuite) TestSetOutputIsStableAcrossInputOrders() {
	elems := []string{`"b"`, `2`, `true`, `null`, `[1]`, `{"k": 1}`, `"a"`, `-1`}
	orders := [][]int{{0, 1, 2, 3, 4, 5, 6, 7}, {7, 6, 5, 4, 3, 2, 1, 0}, {3, 5, 0, 7, 1, 6, 2, 4}}
	var src strings.Builder
	src.WriteString("namespace set/order\n\npolicy p {\n  fact words: set[string]\n  rule out_words = { yield words }\n  export decision of out_words\n")
	for i, order := range orders {
		parts := make([]string, len(order))
		for k, j := range order {
			parts[k] = elems[j]
		}
		fmt.Fprintf(&src, "  rule mixed%d = { yield set([%s]) }\n  export decision of mixed%d\n", i, strings.Join(parts, ", "), i)
	}
	src.WriteString("}\n")
	exec := s.newTempPackExecutor(map[string]string{"policy.sentrie": src.String()})
	defer func() { s.NoError(exec.Close()) }()

	const want = `[null,true,-1,2,"a","b",[1],{"k":1}]`
	for i := range orders {
		for repeat := 0; repeat < 3; repeat++ {
			s.Equal(want, s.ruleJSON(exec, "set/order", "p", fmt.Sprintf("mixed%d", i), map[string]any{"words": []any{}}), "order %d run %d", i, repeat)
		}
	}

	for _, in := range [][]any{{"c", "a", "b"}, {"b", "c", "a"}, {"a", "b", "c"}} {
		s.Equal(`["a","b","c"]`, s.ruleJSON(exec, "set/order", "p", "out_words", map[string]any{"words": in}))
	}
	s.Equal(`[]`, s.ruleJSON(exec, "set/order", "p", "out_words", map[string]any{"words": []any{}}))
}

// TestSetEqualNumbersAreOneElement: -0 and 0, 1 and 1.0, are the same element,
// so set() collapses them and a fact holding both is a duplicate error.
func (s *RuntimeTestSuite) TestSetEqualNumbersAreOneElement() {
	exec := s.newTempPackExecutor(map[string]string{"policy.sentrie": `namespace set/num

policy p {
  fact nums: set[number]
  rule collapsed = { yield set([1, 1.0, 0, -0, 2]) }
  rule fact_out = { yield nums }
  export decision of collapsed
  export decision of fact_out
}
`})
	defer func() { s.NoError(exec.Close()) }()

	s.Equal(`[0,1,2]`, s.ruleJSON(exec, "set/num", "p", "collapsed", map[string]any{"nums": []any{}}))
	s.Equal(`[0,1]`, s.ruleJSON(exec, "set/num", "p", "fact_out", map[string]any{"nums": []any{float64(1), float64(0)}}))

	for _, nums := range [][]any{{float64(1), float64(1)}, {float64(0), math.Copysign(0, -1)}} {
		_, err := exec.ExecRule(s.T().Context(), "set/num", "p", "fact_out", map[string]any{"nums": nums})
		s.Require().Error(err, fmt.Sprint(nums))
		s.Contains(err.Error(), "duplicate set element", fmt.Sprint(nums))
	}
}

// TestSetFactShapes: a fact declared set[T] rejects a non-array, an optional
// set[T]? may be omitted (null) or given, and a set[T] nested in a list keeps
// being validated for type and uniqueness.
func (s *RuntimeTestSuite) TestSetFactShapes() {
	ctx := s.T().Context()
	exec := s.newTempPackExecutor(map[string]string{"policy.sentrie": `namespace set/facts

policy p {
  fact req: set[string]
  fact opt?: set[number]?
  fact nested?: list[set[number]] default []
  rule has_opt = { yield opt is set[number] }
  rule opt_out = { yield opt }
  rule nested_n = { yield count(nested) }
  rule req_n = { yield count(req) }
  export decision of has_opt
  export decision of opt_out
  export decision of nested_n
  export decision of req_n
}
`})
	defer func() { s.NoError(exec.Close()) }()

	base := func(extra map[string]any) map[string]any {
		f := map[string]any{"req": []any{"x"}}
		for k, v := range extra {
			f[k] = v
		}
		return f
	}

	// omitted optional: null, not a set
	out, err := exec.ExecRule(ctx, "set/facts", "p", "has_opt", base(nil))
	s.Require().NoError(err)
	s.Equal(trinary.False, out.Decision.State)
	s.Equal("null", s.ruleJSON(exec, "set/facts", "p", "opt_out", base(nil)))
	s.Equal("null", s.ruleJSON(exec, "set/facts", "p", "opt_out", base(map[string]any{"opt": nil})))

	// provided optional: a sorted set
	s.Equal("[1,2,3]", s.ruleJSON(exec, "set/facts", "p", "opt_out", base(map[string]any{"opt": []any{float64(3), float64(1), float64(2)}})))
	out, err = exec.ExecRule(ctx, "set/facts", "p", "has_opt", base(map[string]any{"opt": []any{float64(1)}}))
	s.Require().NoError(err)
	s.Equal(trinary.True, out.Decision.State)

	// duplicate and wrong element type on the optional
	_, err = exec.ExecRule(ctx, "set/facts", "p", "opt_out", base(map[string]any{"opt": []any{float64(1), float64(1)}}))
	s.Require().Error(err)
	s.Contains(err.Error(), "duplicate set element")
	_, err = exec.ExecRule(ctx, "set/facts", "p", "opt_out", base(map[string]any{"opt": []any{"a"}}))
	s.Require().Error(err)

	// not an array at all
	for _, bad := range []any{"read", float64(1), map[string]any{"a": 1}} {
		_, err = exec.ExecRule(ctx, "set/facts", "p", "req_n", map[string]any{"req": bad})
		s.Require().Error(err, fmt.Sprint(bad))
	}

	// nested set[T]: elements validated, duplicates rejected
	s.Equal("2", s.ruleJSON(exec, "set/facts", "p", "nested_n", base(map[string]any{"nested": []any{[]any{float64(1)}, []any{float64(2), float64(3)}}})))
	_, err = exec.ExecRule(ctx, "set/facts", "p", "nested_n", base(map[string]any{"nested": []any{[]any{float64(1), float64(1)}}}))
	s.Require().Error(err)
	_, err = exec.ExecRule(ctx, "set/facts", "p", "nested_n", base(map[string]any{"nested": []any{[]any{"a"}}}))
	s.Require().Error(err)
}

// TestSetBuiltinsInDeriveAndOnListFacts: the set builtins are pure, so a
// derive body may call them, and they take plain list facts directly. A derive
// parameter or return typed set[T] converts a list into a set.
func (s *RuntimeTestSuite) TestSetBuiltinsInDeriveAndOnListFacts() {
	ctx := s.T().Context()
	exec := s.newTempPackExecutor(map[string]string{"policy.sentrie": `namespace set/dv

derive missing = (need: list[string], have: list[string]): list[string] => { yield difference(need, have) }
derive covers = (need: list[string], have: list[string]): boolean => { yield is_subset(need, have) }
derive both = (a: list[string], b: list[string]): list[string] => { yield union(intersection(a, b), set(a)) }
derive as_set = (xs: set[number]): set[number] => { yield xs }
derive to_set = (xs: list[number]): set[number] => { yield [3, 1, 2] }
derive cnt = (xs: set[number]): number => { yield count(xs) }

policy p {
  fact need: list[string]
  fact have: list[string]
  rule d_missing = { yield missing(need, have) }
  rule d_covers = { yield covers(need, have) }
  rule d_both = { yield both(need, have) }
  rule d_param = { yield as_set([2, 1]) }
  rule d_return = { yield to_set([]) }
  rule d_dup_param = { yield as_set([1, 1]) }
  rule d_count = { yield cnt([1, 2, 3]) }
  rule list_union = { yield union(need, have) }
  rule list_sub = { yield is_subset(need, have) and not is_superset(need, have) }
  rule inline = { yield is_subset(["a"], ["a", "b"]) }
  rule let_out = {
    let x: set[number] = [3, 1, 2]
    yield x
  }
  rule cast_out = { yield ["b", "a", "b"] as set[string] }
  export decision of d_missing
  export decision of d_covers
  export decision of d_both
  export decision of d_param
  export decision of d_return
  export decision of d_dup_param
  export decision of d_count
  export decision of list_union
  export decision of list_sub
  export decision of inline
  export decision of let_out
  export decision of cast_out
}
`})
	defer func() { s.NoError(exec.Close()) }()

	facts := map[string]any{"need": []any{"write", "read", "read"}, "have": []any{"read", "list"}}
	s.Equal(`["write"]`, s.ruleJSON(exec, "set/dv", "p", "d_missing", facts))
	s.Equal(`["list","read","write"]`, s.ruleJSON(exec, "set/dv", "p", "list_union", facts), "plain list facts with duplicates are accepted by the builtins")
	s.Equal(`["read"]`, s.ruleJSON(exec, "set/dv", "p", "d_both", map[string]any{"need": []any{"read"}, "have": []any{"read"}}))
	s.Equal(`[1,2]`, s.ruleJSON(exec, "set/dv", "p", "d_param", facts), "list argument bound to a set parameter is sorted")
	s.Equal(`[1,2,3]`, s.ruleJSON(exec, "set/dv", "p", "d_return", facts), "list returned from a set[number] derive is sorted")
	s.Equal(`3`, s.ruleJSON(exec, "set/dv", "p", "d_count", facts))
	s.Equal(`[1,2,3]`, s.ruleJSON(exec, "set/dv", "p", "let_out", facts), "a typed let binds a list as a sorted set")
	s.Equal(`["a","b"]`, s.ruleJSON(exec, "set/dv", "p", "cast_out", facts), "a cast to set[T] dedupes and sorts")

	out, err := exec.ExecRule(ctx, "set/dv", "p", "d_covers", facts)
	s.Require().NoError(err)
	s.Equal(trinary.False, out.Decision.State)
	out, err = exec.ExecRule(ctx, "set/dv", "p", "d_covers", map[string]any{"need": []any{"a"}, "have": []any{"a", "b"}})
	s.Require().NoError(err)
	s.Equal(trinary.True, out.Decision.State)
	out, err = exec.ExecRule(ctx, "set/dv", "p", "list_sub", map[string]any{"need": []any{"a"}, "have": []any{"a", "b"}})
	s.Require().NoError(err)
	s.Equal(trinary.True, out.Decision.State)
	out, err = exec.ExecRule(ctx, "set/dv", "p", "inline", facts)
	s.Require().NoError(err)
	s.Equal(trinary.True, out.Decision.State)

	// a duplicate list reaching a set[T] parameter is a validation error
	_, err = exec.ExecRule(ctx, "set/dv", "p", "d_dup_param", facts)
	s.Require().Error(err)
	s.Contains(err.Error(), "duplicate set element")
}

// TestSetBuiltinStrictness: undefined and null operands are errors, not empty sets.
func (s *RuntimeTestSuite) TestSetBuiltinStrictness() {
	ctx := s.T().Context()
	exec := s.newTempPackExecutor(map[string]string{"policy.sentrie": `namespace set/strict

policy p {
  fact maybe?: list[string]? default null
  fact undef?: list[string]
  rule u_null = { yield union(maybe, ["a"]) }
  rule i_null = { yield intersection(["a"], maybe) }
  rule s_null = { yield set(maybe) }
  rule sub_undef = { yield is_subset(["a"], undef) }
  rule d_undef = { yield difference(undef, ["a"]) }
  export decision of u_null
  export decision of i_null
  export decision of s_null
  export decision of sub_undef
  export decision of d_undef
}
`})
	defer func() { s.NoError(exec.Close()) }()

	for _, rule := range []string{"u_null", "i_null", "s_null", "sub_undef", "d_undef"} {
		_, err := exec.ExecRule(ctx, "set/strict", "p", rule, map[string]any{})
		s.Require().Error(err, rule)
	}
}

// TestSetAcrossTheJSBoundary: a set handed to a TypeScript module arrives as a
// sorted array; what the module returns is a plain list, so its order is the
// module's, not re-sorted.
func (s *RuntimeTestSuite) TestSetAcrossTheJSBoundary() {
	ctx := s.T().Context()
	exec := s.newTempPackExecutor(map[string]string{
		"sets.ts": `
export function describe(xs: string[]): string {
  return (Array.isArray(xs) ? "array:" : "other:") + xs.join(",");
}
export function reversed(xs: string[]): string[] { return [...xs].reverse(); }
export function size(xs: string[]): number { return xs.length; }
`,
		"policy.sentrie": `namespace set/js

policy p {
  fact scopes: set[string]
  use { describe, reversed, size } from "./sets" as m
  rule seen = { yield m.describe(scopes) == "array:list,read,write" }
  rule rev = { yield m.reversed(scopes) }
  rule rev_back = { yield set(m.reversed(scopes)) }
  rule sz = { yield m.size(scopes) == 3 }
  rule via_set_op = { yield m.describe(union(scopes, ["a"])) == "array:a,list,read,write" }
  export decision of seen
  export decision of rev
  export decision of rev_back
  export decision of sz
  export decision of via_set_op
}
`,
	})
	defer func() { s.NoError(exec.Close()) }()

	facts := map[string]any{"scopes": []any{"write", "list", "read"}}
	for _, rule := range []string{"seen", "sz", "via_set_op"} {
		out, err := exec.ExecRule(ctx, "set/js", "p", rule, facts)
		s.Require().NoError(err, rule)
		s.Equal(trinary.True, out.Decision.State, rule)
	}
	s.Equal(`["write","read","list"]`, s.ruleJSON(exec, "set/js", "p", "rev", facts), "a module's return value is a plain list in the module's order")
	s.Equal(`["list","read","write"]`, s.ruleJSON(exec, "set/js", "p", "rev_back", facts), "set() restores the canonical order")
}

// TestTypedLambdaReturnBindsTruthValueAndSet covers the lambda-return call
// site of bindTypedValue: an imported decision returned from a lambda declared
// trinary holds its state (#136), and a list returned from a lambda declared
// set[number] is sorted (#162).
func (s *RuntimeTestSuite) TestTypedLambdaReturnBindsTruthValueAndSet() {
	ctx := s.T().Context()
	src := importEnvelopeBase + `policy p {
  rule impF = import decision f from t/q
  rule impU = import decision u from t/q
  rule lamF = { yield collect([1], (x: number): trinary => { yield impF })[0] == false }
  rule lamU = { yield collect([1], (x: number): trinary => { yield impU })[0] or false }
  rule lamSet = { yield collect([1], (x: number): set[number] => { yield [3, 1, 2] })[0] }
  export decision of lamF
  export decision of lamU
  export decision of lamSet
}
`
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "lambda_bind.sentrie", src: src})

	out, err := exec.ExecRule(ctx, "t", "p", "lamF", nil)
	s.Require().NoError(err)
	s.Equal(trinary.True, out.Decision.State)
	out, err = exec.ExecRule(ctx, "t", "p", "lamU", nil)
	s.Require().NoError(err)
	s.Equal(trinary.Unknown, out.Decision.State)
	s.Equal(`[1,2,3]`, s.ruleJSON(exec, "t", "p", "lamSet", nil))
}
