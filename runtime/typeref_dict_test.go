// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/trinary"
)

// TestValidateAgainstDictTypeRefValues checks that every value of a dict[T]
// is validated against T, and that the first failing key in sorted order is
// the one reported (#299).
func (r *RuntimeTestSuite) TestValidateAgainstDictTypeRefValues() {
	numbers := ast.NewDictTypeRef(ast.NewNumberTypeRef(stubRange()), stubRange())
	validate := func(v any, typeRef *ast.DictTypeRef) error {
		return validateAgainstDictTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, box.FromAny(v), typeRef, stubRange())
	}

	r.Run("accepts an empty dict", func() {
		r.NoError(validate(map[string]any{}, numbers))
	})

	r.Run("accepts values of the element type", func() {
		r.NoError(validate(map[string]any{"x": 1.0, "y": 2.0}, numbers))
	})

	r.Run("names the first failing key in sorted order", func() {
		for range 20 {
			err := validate(map[string]any{"d": "no", "b": "no", "c": 3.0, "a": 1.0}, numbers)
			r.Require().Error(err)
			r.Contains(err.Error(), "value for key \"b\" is not valid")
		}
	})

	r.Run("checks nested element types", func() {
		nested := ast.NewDictTypeRef(ast.NewListTypeRef(ast.NewNumberTypeRef(stubRange()), stubRange()), stubRange())
		r.NoError(validate(map[string]any{"k": []any{1.0, 2.0}}, nested))
		err := validate(map[string]any{"k": []any{1.0, "two"}}, nested)
		r.Require().Error(err)
		r.Contains(err.Error(), "value for key \"k\" is not valid")
		r.Contains(err.Error(), "item is not valid")
	})

	r.Run("names the key at each level of a nested dict", func() {
		nested := ast.NewDictTypeRef(numbers, stubRange())
		err := validate(map[string]any{"outer": map[string]any{"inner": "no"}}, nested)
		r.Require().Error(err)
		r.Contains(err.Error(), "value for key \"outer\" is not valid")
		r.Contains(err.Error(), "value for key \"inner\" is not valid")
	})

	r.Run("accepts a null value for a nullable element type", func() {
		nullable := ast.NewDictTypeRef(ast.NewNullableTypeRef(ast.NewNumberTypeRef(stubRange()), stubRange()), stubRange())
		r.NoError(validate(map[string]any{"k": nil, "n": 1.0}, nullable))
		r.Error(validate(map[string]any{"k": nil}, numbers))
	})

	r.Run("skips an entry holding undefined", func() {
		// Undefined is not a value, so the entry counts as absent (#148);
		// a real value of the wrong type beside it still fails.
		r.NoError(validate(map[string]box.Value{"u": box.Undefined(), "n": box.FromAny(1.0)}, numbers))
		r.Error(validate(map[string]box.Value{"u": box.Undefined(), "s": box.FromAny("no")}, numbers))
		nested := ast.NewDictTypeRef(numbers, stubRange())
		r.NoError(validate(map[string]box.Value{"o": box.Dict(map[string]box.Value{"u": box.Undefined()})}, nested))
	})
}

// TestDictElementTypeInEveryPosition checks that dict[T] validates its values
// wherever a type is written: facts, typed lets, casts, `is`, shape fields,
// and derive parameters and returns (#299).
func (s *RuntimeTestSuite) TestDictElementTypeInEveryPosition() {
	ctx := s.T().Context()
	src := `namespace com/app
shape Holder { scores: dict[number] }

derive total = (d: dict[number]): number => { yield 1 }
derive bad_return = (): dict[number] => { yield { "k": "s" } }

policy pol {
  fact scores: dict[number] as scores
  rule is_ok = { yield { "a": 1 } is dict[number] }
  rule is_bad = { yield { "a": 1, "b": "s" } is dict[number] }
  rule is_nested = { yield { "a": { "b": [1] } } is dict[dict[list[number]]] and not ({ "a": { "b": ["s"] } } is dict[dict[list[number]]]) }
  rule is_nullable = { yield { "a": null } is dict[number?] and not ({ "a": null } is dict[number]) }
  rule is_constrained = { yield { "a": 3 } is dict[number@min(2)] and not ({ "a": 1 } is dict[number@min(2)]) }
  rule is_shape = { yield { "scores": { "a": 1 } } is Holder and not ({ "scores": { "a": "s" } } is Holder) }
  rule fact_value = { yield scores }
  rule let_bad = {
    let d: dict[number] = { "a": "s" }
    yield d
  }
  rule cast_bad = { yield { "a": "s" } as dict[number] }
  rule param_bad = { yield total({ "a": "s" }) }
  rule return_bad = { yield bad_return() }
  export decision of is_ok
  export decision of is_bad
  export decision of is_nested
  export decision of is_nullable
  export decision of is_constrained
  export decision of is_shape
  export decision of fact_value
  export decision of let_bad
  export decision of cast_bad
  export decision of param_bad
  export decision of return_bad
}
`
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "dict.sentrie", src: src})

	checks := map[string]trinary.Value{
		"is_ok":          trinary.True,
		"is_bad":         trinary.False,
		"is_nested":      trinary.True,
		"is_nullable":    trinary.True,
		"is_constrained": trinary.True,
		"is_shape":       trinary.True,
	}
	for rule, want := range checks {
		s.Run(rule, func() {
			out, err := exec.ExecRule(ctx, "com/app", "pol", rule, map[string]any{"scores": map[string]any{"a": 1.0}})
			s.Require().NoError(err)
			s.Equal(want, out.Decision.State)
		})
	}

	s.Run("fact accepts a valid dict", func() {
		_, err := exec.ExecRule(ctx, "com/app", "pol", "fact_value", map[string]any{"scores": map[string]any{"a": 1.0}})
		s.Require().NoError(err)
	})

	s.Run("fact rejects a value of the wrong type", func() {
		_, err := exec.ExecRule(ctx, "com/app", "pol", "fact_value", map[string]any{"scores": map[string]any{"a": 1.0, "b": "s"}})
		s.Require().Error(err)
		s.Contains(err.Error(), "value for key \"b\" is not valid")
	})

	for _, rule := range []string{"let_bad", "cast_bad", "param_bad", "return_bad"} {
		s.Run(rule, func() {
			_, err := exec.ExecRule(ctx, "com/app", "pol", rule, map[string]any{"scores": map[string]any{"a": 1.0}})
			s.Require().Error(err)
			s.Contains(err.Error(), "value for key")
		})
	}
}

// TestDictElementTypeStructuredValues checks dict[T] where T is a shape or a
// record, a lambda returning a dict, and that a bad value fails the same way
// on every run (#299).
func (s *RuntimeTestSuite) TestDictElementTypeStructuredValues() {
	ctx := s.T().Context()
	src := `namespace com/app
shape Item {
  name: string
  qty: number
}

policy pol {
  fact items: dict[Item] as items
  fact pairs: dict[record[string, number]] as pairs
  rule shape_ok = { yield { "a": { "name": "x", "qty": 1 } } is dict[Item] }
  rule shape_missing_field = { yield { "a": { "name": "x", "qty": 1 }, "z": { "name": "y" } } is dict[Item] }
  rule shape_wrong_field = { yield { "a": { "name": "x", "qty": "many" } } is dict[Item] }
  rule record_ok = { yield { "a": ["x", 1] } is dict[record[string, number]] }
  rule record_bad = { yield { "a": ["x", "y"] } is dict[record[string, number]] }
  rule empty_ok = { yield {} is dict[Item] and {} is dict[dict[number]] }
  rule lambda_bad = { yield collect([1], (n: number): dict[number] => { yield { "k": "s" } }) }
  rule lambda_ok = { yield count(collect([1], (n: number): dict[number] => { yield { "k": n } })) == 1 }
  rule fact_items = { yield items }
  rule fact_pairs = { yield pairs }
  export decision of shape_ok
  export decision of shape_missing_field
  export decision of shape_wrong_field
  export decision of record_ok
  export decision of record_bad
  export decision of empty_ok
  export decision of lambda_bad
  export decision of lambda_ok
  export decision of fact_items
  export decision of fact_pairs
}
`
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "dictstruct.sentrie", src: src})
	good := map[string]any{
		"items": map[string]any{"a": map[string]any{"name": "x", "qty": 1.0}},
		"pairs": map[string]any{"a": []any{"x", 1.0}},
	}

	checks := map[string]trinary.Value{
		"shape_ok":            trinary.True,
		"shape_missing_field": trinary.False,
		"shape_wrong_field":   trinary.False,
		"record_ok":           trinary.True,
		"record_bad":          trinary.False,
		"empty_ok":            trinary.True,
		"lambda_ok":           trinary.True,
	}
	for rule, want := range checks {
		s.Run(rule, func() {
			out, err := exec.ExecRule(ctx, "com/app", "pol", rule, good)
			s.Require().NoError(err)
			s.Equal(want, out.Decision.State)
		})
	}

	s.Run("lambda return of the wrong element type is an error", func() {
		_, err := exec.ExecRule(ctx, "com/app", "pol", "lambda_bad", good)
		s.Require().Error(err)
		s.Contains(err.Error(), "value for key \"k\" is not valid")
	})

	s.Run("fact dict of shapes rejects a value missing a field and names the key", func() {
		facts := map[string]any{
			"items": map[string]any{
				"a": map[string]any{"name": "x", "qty": 1.0},
				"z": map[string]any{"name": "y"},
			},
			"pairs": map[string]any{},
		}
		for range 10 {
			_, err := exec.ExecRule(ctx, "com/app", "pol", "fact_items", facts)
			s.Require().Error(err)
			s.Contains(err.Error(), "value for key \"z\" is not valid")
		}
	})

	s.Run("fact dict of records rejects a value of the wrong element type", func() {
		facts := map[string]any{
			"items": map[string]any{},
			"pairs": map[string]any{"a": []any{"x", 1.0}, "b": []any{"x", "y"}},
		}
		_, err := exec.ExecRule(ctx, "com/app", "pol", "fact_pairs", facts)
		s.Require().Error(err)
		s.Contains(err.Error(), "value for key \"b\" is not valid")
	})

	s.Run("empty facts conform", func() {
		_, err := exec.ExecRule(ctx, "com/app", "pol", "fact_items", map[string]any{"items": map[string]any{}, "pairs": map[string]any{}})
		s.Require().NoError(err)
	})
}

// TestDictElementTypeFromModule checks that an object returned by a TypeScript
// module is checked value by value when bound to dict[T] (#299).
func (s *RuntimeTestSuite) TestDictElementTypeFromModule() {
	ctx := s.T().Context()
	exec := s.newTempPackExecutor(map[string]string{
		"objs.ts": `
export function good(): Record<string, number> { return { a: 1, b: 2 }; }
export function bad(): Record<string, unknown> { return { a: 1, b: "two" }; }
`,
		"policy.sentrie": `namespace dictmod/test

policy p {
  use { good, bad } from "./objs" as m
  rule ok = {
    let d: dict[number] = m.good()
    yield d
  }
  rule wrong = {
    let d: dict[number] = m.bad()
    yield d
  }
  export decision of ok
  export decision of wrong
}
`,
	})

	_, err := exec.ExecRule(ctx, "dictmod/test", "p", "ok", nil)
	s.Require().NoError(err)

	_, err = exec.ExecRule(ctx, "dictmod/test", "p", "wrong", nil)
	s.Require().Error(err)
	s.Contains(err.Error(), "value for key \"b\" is not valid")
}
