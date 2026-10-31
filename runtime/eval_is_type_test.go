// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"encoding/json"
	"strconv"

	"github.com/sentrie-sh/sentrie/trinary"
)

func (s *RuntimeTestSuite) TestEvalIsTypeCheck() {
	ctx := s.T().Context()
	lib := `namespace com/lib
shape Account { id: string }
export shape Account
`
	src := `namespace com/app
shape U { name: string  age: number }

policy pol {
  shape Local { ok: trinary }
  let good = { "name": "a", "age": 1 }
  let bad = { "name": "a" }
  let num = 5
  rule shape_ok = { yield good is U }
  rule shape_missing_field = { yield bad is U }
  rule shape_wrong_kind = { yield num is U }
  rule shape_negated = { yield bad is not U }
  rule policy_shape = { yield { "ok": true } is Local }
  rule qualified = { yield { "id": "x" } is com/lib/Account }
  rule primitive = { yield "s" is string and 1 is number and not ("s" is number) }
  rule list_type = { yield [1, 2] is list[number] and not (["a"] is list[number]) }
  rule constrained = { yield "abc" is string@minlength(2) and not ("a" is string@minlength(2)) }
  rule nullable = { yield null is string? and not (null is string) }
  rule ternary_after_type = { yield ("s" is string ? 1 : 2) == 1 }
  rule undefined_operand = { yield not (good.missing is string) and not (good.missing is string?) }
  rule legacy_equality = { yield num is num }
  rule in_derive = { yield isU(good) and not isU(bad) }
  export decision of shape_ok
  export decision of shape_missing_field
  export decision of shape_wrong_kind
  export decision of shape_negated
  export decision of policy_shape
  export decision of qualified
  export decision of primitive
  export decision of list_type
  export decision of constrained
  export decision of nullable
  export decision of ternary_after_type
  export decision of undefined_operand
  export decision of legacy_equality
  export decision of in_derive
}

derive isU = (v: document): trinary => { yield v is U }
`
	_, exec := s.mustBuildDeriveExecutor(ctx,
		deriveTestProgram{name: "lib.sentrie", src: lib},
		deriveTestProgram{name: "app.sentrie", src: src},
	)

	cases := map[string]trinary.Value{
		"shape_ok":            trinary.True,
		"shape_missing_field": trinary.False,
		"shape_wrong_kind":    trinary.False,
		"shape_negated":       trinary.True,
		"policy_shape":        trinary.True,
		"qualified":           trinary.True,
		"primitive":           trinary.True,
		"list_type":           trinary.True,
		"constrained":         trinary.True,
		"nullable":            trinary.True,
		"ternary_after_type":  trinary.True,
		"undefined_operand":   trinary.True,
		"legacy_equality":     trinary.True,
		"in_derive":           trinary.True,
	}
	for rule, want := range cases {
		s.Run(rule, func() {
			out, err := exec.ExecRule(ctx, "com/app", "pol", rule, nil)
			s.Require().NoError(err)
			s.Equal(want, out.Decision.State)
		})
	}
}

func (s *RuntimeTestSuite) TestEvalIsNameResolutionErrors() {
	ctx := s.T().Context()
	cases := map[string]struct {
		body string
		msg  string
	}{
		"misspelt shape": {body: `rule r = { yield 1 is Usr }`, msg: `"Usr" is neither a shape nor a value in scope`},
		"shape and let":  {body: "let U = 1\n  rule r = { yield 1 is U }", msg: `"U" names both a shape and a value in scope`},
	}
	for name, tc := range cases {
		s.Run(name, func() {
			src := "namespace com/app\nshape U { name: string }\npolicy pol {\n  " + tc.body + "\n  export decision of r\n}\n"
			idx := s.mustBuildDeriveIndex(ctx, deriveTestProgram{name: "err.sentrie", src: src})
			err := idx.Validate(ctx)
			s.Require().Error(err)
			s.Contains(err.Error(), tc.msg)
		})
	}
}

// TestEvalIsTypeCheckErrorsPropagate checks that a failure to make the check
// is an error, never a false that `is not` would turn into true.
func (s *RuntimeTestSuite) TestEvalIsTypeCheckErrorsPropagate() {
	ctx := s.T().Context()
	src := `namespace com/app
shape Box { inner: Hidden }

policy pol {
  rule bad_arg_kind = { yield "ab" is string@minlength("q") }
  rule bad_regexp = { yield "ab" is not string@regexp("[") }
  rule hidden_field_shape = { yield { "inner": {} } is Box }
  export decision of bad_arg_kind
  export decision of bad_regexp
  export decision of hidden_field_shape
}
`
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "errs.sentrie", src: src})
	cases := map[string]string{
		"bad_arg_kind":       "expected number, got string",
		"bad_regexp":         "invalid regexp pattern",
		"hidden_field_shape": "Hidden",
	}
	for rule, msg := range cases {
		s.Run(rule, func() {
			_, err := exec.ExecRule(ctx, "com/app", "pol", rule, nil)
			s.Require().Error(err)
			s.Contains(err.Error(), msg)
		})
	}
}

// TestEvalIsTypeUnvalidatedMissingShapeErrors covers the runtime path for a
// type the index never checked: a missing shape is an error, not false.
func (s *RuntimeTestSuite) TestEvalIsTypeUnvalidatedMissingShapeErrors() {
	ctx := s.T().Context()
	src := `namespace com/app
policy pol {
  rule r = { yield [{ "a": 1 }] is list[Missing] }
  export decision of r
}
`
	idx := s.mustBuildDeriveIndex(ctx, deriveTestProgram{name: "unvalidated.sentrie", src: src})
	rawExec, err := NewExecutor(idx)
	s.Require().NoError(err)
	exec, ok := rawExec.(*executorImpl)
	s.Require().True(ok)
	pol, err := idx.ResolvePolicy("com/app", "pol")
	s.Require().NoError(err)
	ec := NewExecutionContext(pol, exec)
	defer ec.Dispose()
	_, _, err = eval(ctx, ec, exec, pol, pol.Rules["r"].Body)
	s.Require().Error(err)
	s.Contains(err.Error(), "Missing")
}

// TestEvalIsTypeBrokenConstraintArgumentErrorsForEveryValue is the #273
// reproduction on the runtime path, for an index that was never validated: a
// broken constraint argument is an error whatever the value's kind, so
// `x is not T` never reads it as a mismatch and returns true.
func (s *RuntimeTestSuite) TestEvalIsTypeBrokenConstraintArgumentErrorsForEveryValue() {
	ctx := s.T().Context()
	src := `namespace com/app
policy pol {
  rule num_value = { yield 1 is not string@regexp("[") }
  rule str_value = { yield "a" is not string@regexp("[") }
  rule null_value = { yield null is not string@regexp("[") }
  rule null_nullable = { yield null is not string?@regexp("[") }
  rule list_item = { yield ["a"] is not list[number@multiple_of(0)] }
  rule trinary_number = { yield 1 is not trinary@eq(1) }
  rule range_string = { yield "a" is not number@range(5, 1) }
  rule valid = { yield 1 is not string@regexp("^a") }
  export decision of num_value
}
`
	idx := s.mustBuildDeriveIndex(ctx, deriveTestProgram{name: "broken_arg.sentrie", src: src})
	rawExec, err := NewExecutor(idx)
	s.Require().NoError(err)
	exec, ok := rawExec.(*executorImpl)
	s.Require().True(ok)
	pol, err := idx.ResolvePolicy("com/app", "pol")
	s.Require().NoError(err)

	cases := map[string]string{
		"num_value":      "invalid regexp pattern",
		"str_value":      "invalid regexp pattern",
		"null_value":     "invalid regexp pattern",
		"null_nullable":  "invalid regexp pattern",
		"trinary_number": "eq constraint expects a boolean argument",
		"range_string":   "range minimum 5 is greater than maximum 1",
	}
	for rule, msg := range cases {
		s.Run(rule, func() {
			ec := NewExecutionContext(pol, exec)
			defer ec.Dispose()
			_, _, err := eval(ctx, ec, exec, pol, pol.Rules[rule].Body)
			s.Require().Error(err)
			s.Contains(err.Error(), msg)
			s.False(isTypeMismatch(err), err.Error())
		})
	}

	s.Run("a list item type is checked per item, so a mismatched item reports it", func() {
		ec := NewExecutionContext(pol, exec)
		defer ec.Dispose()
		_, _, err := eval(ctx, ec, exec, pol, pol.Rules["list_item"].Body)
		s.Require().Error(err)
		s.Contains(err.Error(), "divisor cannot be zero")
	})

	s.Run("a valid argument still reads a wrong kind as a mismatch", func() {
		ec := NewExecutionContext(pol, exec)
		defer ec.Dispose()
		got, _, err := eval(ctx, ec, exec, pol, pol.Rules["valid"].Body)
		s.Require().NoError(err)
		tv, ok := got.TrinaryValue()
		s.Require().True(ok)
		s.Equal(trinary.True, tv)
	})
}

// TestEvalIsTypeInDeriveUsesDeriveScope checks that a derive body resolves
// shapes in its own namespace, as the index did, whoever calls it.
func (s *RuntimeTestSuite) TestEvalIsTypeInDeriveUsesDeriveScope() {
	ctx := s.T().Context()
	lib := `namespace t/d
shape Foo { a: string }
derive isFoo = (v: document): trinary => { yield v is Foo }
export derive isFoo
`
	for name, callerShapes := range map[string]string{
		"caller without the shape":              "",
		"caller with its own shape of the name": "shape Foo { b: number }\n",
	} {
		s.Run(name, func() {
			caller := "namespace t/e\n" + callerShapes + `policy p {
  let u = { "a": "x" }
  rule e1 = { yield t/d/isFoo(u) }
  export decision of e1
}
`
			_, exec := s.mustBuildDeriveExecutor(ctx,
				deriveTestProgram{name: "d.sentrie", src: lib},
				deriveTestProgram{name: "e.sentrie", src: caller},
			)
			out, err := exec.ExecRule(ctx, "t/e", "p", "e1", nil)
			s.Require().NoError(err)
			s.Equal(trinary.True, out.Decision.State)
		})
	}
}

const isEndToEndSource = `namespace com/app
shape Item { sku: string@minlength(3)  qty: number@min(1) }
shape Order { id: string  items: list[Item]  note?: string }

policy pol {
  fact payload: document
  rule conforms = { yield payload is Order }
  rule malformed = { yield payload is not Order }
  rule skus = { yield payload.tags is list[string@minlength(2)] }
  rule nullable_elems = { yield payload.tags is list[string?] }
  export decision of conforms
  export decision of malformed
  export decision of skus
  export decision of nullable_elems
}
`

// TestEvalIsTypeEndToEndThroughExec runs `is` and `is not` through ExecRule
// with facts, checking the decision state and the JSON value for input that
// conforms and input that does not in different ways. The deny-style rule
// (`is not`) must be true for every malformed input, never fail open.
func (s *RuntimeTestSuite) TestEvalIsTypeEndToEndThroughExec() {
	ctx := s.T().Context()
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "e2e.sentrie", src: isEndToEndSource})

	item := func(sku string, qty any) map[string]any { return map[string]any{"sku": sku, "qty": qty} }
	cases := map[string]struct {
		payload   map[string]any
		conforms  bool
		malformed bool
	}{
		"conforming order": {
			payload:  map[string]any{"id": "o1", "items": []any{item("abc", 2), item("abcd", 1)}},
			conforms: true,
		},
		"conforming order with a note": {
			payload:  map[string]any{"id": "o1", "items": []any{}, "note": "hi"},
			conforms: true,
		},
		"missing items":            {payload: map[string]any{"id": "o1"}, malformed: true},
		"wrong id kind":            {payload: map[string]any{"id": 7, "items": []any{}}, malformed: true},
		"element constraint fails": {payload: map[string]any{"id": "o1", "items": []any{item("abc", 2), item("ab", 2)}}, malformed: true},
		"nested kind fails":        {payload: map[string]any{"id": "o1", "items": []any{item("abc", "2")}}, malformed: true},
		"nested number below min":  {payload: map[string]any{"id": "o1", "items": []any{item("abc", 0)}}, malformed: true},
		"element is not a document": {
			payload:   map[string]any{"id": "o1", "items": []any{"abc"}},
			malformed: true,
		},
		"items is not a list": {payload: map[string]any{"id": "o1", "items": item("abc", 1)}, malformed: true},
		"optional note wrong kind": {
			payload:   map[string]any{"id": "o1", "items": []any{}, "note": 1},
			malformed: true,
		},
		"empty document": {payload: map[string]any{}, malformed: true},
	}
	for name, tc := range cases {
		s.Run(name, func() {
			facts := map[string]any{"payload": tc.payload}
			out, err := exec.ExecRule(ctx, "com/app", "pol", "conforms", facts)
			s.Require().NoError(err)
			s.Equal(trinary.From(tc.conforms), out.Decision.State)
			raw, err := json.Marshal(out.Decision.Value)
			s.Require().NoError(err)
			s.JSONEq(strconv.FormatBool(tc.conforms), string(raw))

			out, err = exec.ExecRule(ctx, "com/app", "pol", "malformed", facts)
			s.Require().NoError(err)
			s.Equal(trinary.From(tc.malformed), out.Decision.State, "deny-style rule must be true for malformed input")
		})
	}
}

// TestEvalIsTypeConstraintsOnListElements covers a constraint on the element
// type of a list, and null elements against a nullable and a non-nullable
// element type.
func (s *RuntimeTestSuite) TestEvalIsTypeConstraintsOnListElements() {
	ctx := s.T().Context()
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "e2e.sentrie", src: isEndToEndSource})
	cases := map[string]struct {
		tags     any
		skus     trinary.Value
		nullable trinary.Value
	}{
		"every element long enough": {tags: []any{"ab", "abc"}, skus: trinary.True, nullable: trinary.True},
		"one element too short":     {tags: []any{"ab", "c"}, skus: trinary.False, nullable: trinary.True},
		"null element":              {tags: []any{"ab", nil}, skus: trinary.False, nullable: trinary.True},
		"number element":            {tags: []any{"ab", 3}, skus: trinary.False, nullable: trinary.False},
		"not a list":                {tags: "ab", skus: trinary.False, nullable: trinary.False},
		"field absent":              {tags: nil, skus: trinary.False, nullable: trinary.True},
	}
	for name, tc := range cases {
		s.Run(name, func() {
			payload := map[string]any{"id": "x", "items": []any{}}
			if name != "field absent" {
				payload["tags"] = tc.tags
			}
			facts := map[string]any{"payload": payload}
			out, err := exec.ExecRule(ctx, "com/app", "pol", "skus", facts)
			s.Require().NoError(err)
			s.Equal(tc.skus, out.Decision.State)
			out, err = exec.ExecRule(ctx, "com/app", "pol", "nullable_elems", facts)
			s.Require().NoError(err)
			// an absent field is undefined, so `is` is false even for list[string?]
			want := tc.nullable
			if name == "field absent" {
				want = trinary.False
			}
			s.Equal(want, out.Decision.State)
		})
	}
}

// TestEvalIsTypeInLambdaForBuiltins runs `x is T` inside lambdas handed to
// any, all and filter, where the shape resolves in the enclosing policy.
func (s *RuntimeTestSuite) TestEvalIsTypeInLambdaForBuiltins() {
	ctx := s.T().Context()
	src := `namespace com/app
shape Item { sku: string@minlength(3) }

policy pol {
  fact docs: list[document]
  rule any_item = { yield any(docs, (d: document): trinary => { yield d is Item }) }
  rule all_items = { yield all(docs, (d: document): trinary => { yield d is Item }) }
  rule no_bad = { yield all(docs, (d: document): trinary => { yield not (d is not Item) }) }
  rule item_count = { yield count(filter(docs, (d: document): trinary => { yield d is Item })) == 2 }
  rule shape_in_inner = { yield any(docs, (d: document): trinary => { yield any([d], (e: document): trinary => { yield e is Item }) }) }
  export decision of any_item
  export decision of all_items
  export decision of no_bad
  export decision of item_count
  export decision of shape_in_inner
}
`
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "lambda.sentrie", src: src})
	good := func(sku string) any { return map[string]any{"sku": sku} }
	cases := map[string]struct {
		docs []any
		want map[string]trinary.Value
	}{
		"mixed": {
			docs: []any{good("abc"), map[string]any{"sku": "a"}, good("abcd"), map[string]any{}},
			want: map[string]trinary.Value{"any_item": trinary.True, "all_items": trinary.False, "no_bad": trinary.False, "item_count": trinary.True, "shape_in_inner": trinary.True},
		},
		"all conforming": {
			docs: []any{good("abc"), good("abcd")},
			want: map[string]trinary.Value{"any_item": trinary.True, "all_items": trinary.True, "no_bad": trinary.True, "item_count": trinary.True, "shape_in_inner": trinary.True},
		},
		"none conforming": {
			docs: []any{map[string]any{"sku": 1}, map[string]any{}},
			want: map[string]trinary.Value{"any_item": trinary.False, "all_items": trinary.False, "no_bad": trinary.False, "item_count": trinary.False, "shape_in_inner": trinary.False},
		},
	}
	for name, tc := range cases {
		for rule, want := range tc.want {
			s.Run(name+"/"+rule, func() {
				out, err := exec.ExecRule(ctx, "com/app", "pol", rule, map[string]any{"docs": tc.docs})
				s.Require().NoError(err)
				s.Equal(want, out.Decision.State)
			})
		}
	}
}

// TestEvalIsTypeDenyRuleDoesNotFailOpen checks `deny = x is not T` where x is
// empty, undefined, null or of the wrong kind (deny is true), and where the check itself
// is broken (an error, never a quiet true or false).
func (s *RuntimeTestSuite) TestEvalIsTypeDenyRuleDoesNotFailOpen() {
	ctx := s.T().Context()
	src := `namespace com/app
shape Allowed { role: string@one_of("admin", "ops") }
shape BrokenField { name: string@regexp("[") }

policy pol {
  fact x: document
  rule deny = { yield x is not Allowed }
  rule deny_nullable = { yield x.s is not string? }
  rule deny_str = { yield x.s is not string@minlength(2) }
  rule deny_broken_field = { yield x is not BrokenField }
  rule deny_broken_arg = { yield x.s is not string@minlength("q") }
  export decision of deny
  export decision of deny_nullable
  export decision of deny_str
  export decision of deny_broken_field
  export decision of deny_broken_arg
}
`
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "deny.sentrie", src: src})
	run := func(rule string, facts map[string]any) (trinary.Value, error) {
		out, err := exec.ExecRule(ctx, "com/app", "pol", rule, facts)
		if err != nil {
			return trinary.Unknown, err
		}
		return out.Decision.State, nil
	}

	s.Run("deny is true for an empty document", func() {
		got, err := run("deny", map[string]any{"x": map[string]any{}})
		s.Require().NoError(err)
		s.Equal(trinary.True, got)
	})
	s.Run("deny is true for a wrong kind inside the document", func() {
		got, err := run("deny", map[string]any{"x": map[string]any{"role": 5}})
		s.Require().NoError(err)
		s.Equal(trinary.True, got)
	})
	s.Run("deny is true for a value the constraint rejects", func() {
		got, err := run("deny", map[string]any{"x": map[string]any{"role": "guest"}})
		s.Require().NoError(err)
		s.Equal(trinary.True, got)
	})
	s.Run("deny is false for an allowed value", func() {
		got, err := run("deny", map[string]any{"x": map[string]any{"role": "ops"}})
		s.Require().NoError(err)
		s.Equal(trinary.False, got)
	})
	s.Run("absent field is undefined, not string?", func() {
		got, err := run("deny_nullable", map[string]any{"x": map[string]any{}})
		s.Require().NoError(err)
		s.Equal(trinary.True, got)
	})
	s.Run("null field is string?", func() {
		got, err := run("deny_nullable", map[string]any{"x": map[string]any{"s": nil}})
		s.Require().NoError(err)
		s.Equal(trinary.False, got)
	})
	s.Run("deny on a too short string", func() {
		got, err := run("deny_str", map[string]any{"x": map[string]any{"s": "a"}})
		s.Require().NoError(err)
		s.Equal(trinary.True, got)
	})
	s.Run("a broken regexp inside a field's type errors for a conforming document", func() {
		_, err := run("deny_broken_field", map[string]any{"x": map[string]any{"name": "a"}})
		s.Require().Error(err)
		s.Contains(err.Error(), "invalid regexp pattern")
	})
	s.Run("a broken constraint argument errors even when the value is too short", func() {
		_, err := run("deny_broken_arg", map[string]any{"x": map[string]any{"s": "a"}})
		s.Require().Error(err)
	})
}

// TestEvalIsTypeMemoizedDerive checks that a memoized derive call whose body
// is an `is` check is cached per argument value: a conforming and a
// non-conforming document never share an answer, and repeats are stable.
func (s *RuntimeTestSuite) TestEvalIsTypeMemoizedDerive() {
	ctx := s.T().Context()
	src := `namespace com/app
shape Item { sku: string@minlength(3) }
derive isItem = (v: document): trinary => { yield v is Item }

policy pol {
  fact doc: document
  rule memo = { yield isItem(doc)!60 }
  rule plain = { yield isItem(doc) }
  export decision of memo
  export decision of plain
}
`
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "memo.sentrie", src: src})
	conforming := map[string]any{"doc": map[string]any{"sku": "abc"}}
	bad := map[string]any{"doc": map[string]any{"sku": "a"}}
	for i, step := range []struct {
		facts map[string]any
		want  trinary.Value
	}{
		{conforming, trinary.True},
		{bad, trinary.False},
		{conforming, trinary.True},
		{bad, trinary.False},
	} {
		for _, rule := range []string{"memo", "plain"} {
			out, err := exec.ExecRule(ctx, "com/app", "pol", rule, step.facts)
			s.Require().NoError(err, "step %d %s", i, rule)
			s.Equal(step.want, out.Decision.State, "step %d %s", i, rule)
		}
	}
}
