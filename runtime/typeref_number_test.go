// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"math"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/tokens"
	"github.com/sentrie-sh/sentrie/trinary"
)

func (r *RuntimeTestSuite) TestValidateAgainstNumberTypeRef() {
	typeRef := ast.NewNumberTypeRef(tokens.Range{File: "test.sentrie", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})

	r.Run("should return an error if the value is a string", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentrie", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstNumberTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, box.FromAny("not a number"), typeRef, mockExpr.Span())

		r.Error(err)
		r.Equal("value not a number is not a number", err.Error())
	})

	r.Run("should not return an error if the value is an int64", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentrie", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstNumberTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, box.FromAny(int64(123)), typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should return an error if the value is a bool", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentrie", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstNumberTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, box.FromAny(true), typeRef, mockExpr.Span())

		r.Error(err)
		r.Equal("value true is not a number", err.Error())
	})

	r.Run("should return an error if the value is a string number", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentrie", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstNumberTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, box.FromAny("123.45"), typeRef, mockExpr.Span())

		r.Error(err)
		r.Equal("value 123.45 is not a number", err.Error())
	})

	r.Run("should not return an error if the value is a positive number", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentrie", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstNumberTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, box.FromAny(float64(123.45)), typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should not return an error if the value is a negative number", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentrie", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstNumberTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, box.FromAny(float64(-123.45)), typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should not return an error if the value is zero", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentrie", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstNumberTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, box.FromAny(float64(0)), typeRef, mockExpr.Span())

		r.NoError(err)
	})
}

func (r *RuntimeTestSuite) TestValidateAgainstNumberTypeRefWithConstraints() {
	// Test positive constraint (no arguments)
	typeRef := ast.NewNumberTypeRef(tokens.Range{File: "test.sentrie", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})

	// Add a positive constraint
	constraint := ast.NewTypeRefConstraint(
		"positive",
		[]ast.Expression{},
		tokens.Range{File: "test.sentrie", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
	)
	_ = typeRef.AddConstraint(constraint)

	r.Run("should pass when value is positive", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentrie", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstNumberTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, box.FromAny(float64(15.0)), typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should pass when value is a small positive number", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentrie", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstNumberTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, box.FromAny(float64(0.001)), typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should fail when value is zero", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentrie", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstNumberTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, box.FromAny(float64(0)), typeRef, mockExpr.Span())

		r.Error(err)
		r.Contains(err.Error(), "constraint failed")
	})

	r.Run("should fail when value is negative", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentrie", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstNumberTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, box.FromAny(float64(-5.0)), typeRef, mockExpr.Span())

		r.Error(err)
		r.Contains(err.Error(), "constraint failed")
	})
}

func (r *RuntimeTestSuite) TestValidateAgainstNumberTypeRefEdgeCases() {
	// Test finite constraint
	typeRef := ast.NewNumberTypeRef(tokens.Range{File: "test.sentrie", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})

	// Add a finite constraint
	constraint := ast.NewTypeRefConstraint(
		"finite",
		[]ast.Expression{},
		tokens.Range{File: "test.sentrie", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}},
	)
	_ = typeRef.AddConstraint(constraint)

	r.Run("should pass when value is a normal finite number", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentrie", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstNumberTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, box.FromAny(float64(123.45)), typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should pass when value is zero", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentrie", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstNumberTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, box.FromAny(float64(0)), typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should pass when value is a very small number", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentrie", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstNumberTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, box.FromAny(float64(1e-300)), typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should pass when value is a very large number", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentrie", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstNumberTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, box.FromAny(float64(1e300)), typeRef, mockExpr.Span())

		r.NoError(err)
	})

	r.Run("should fail when value is positive infinity", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentrie", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstNumberTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, box.FromAny(math.Inf(1)), typeRef, mockExpr.Span())

		r.Error(err)
		r.Contains(err.Error(), "constraint failed")
	})

	r.Run("should fail when value is negative infinity", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentrie", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstNumberTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, box.FromAny(math.Inf(-1)), typeRef, mockExpr.Span())

		r.Error(err)
		r.Contains(err.Error(), "constraint failed")
	})

	r.Run("should fail when value is NaN", func() {
		// Create a mock expression for the test
		mockExpr := ast.NewIdentifier("test", tokens.Range{File: "test.sentrie", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}})
		err := validateAgainstNumberTypeRef(r.T().Context(), &ExecutionContext{}, &executorImpl{}, &index.Policy{}, box.FromAny(math.NaN()), typeRef, mockExpr.Span())

		r.Error(err)
		r.Contains(err.Error(), "constraint failed")
	})
}

// TestNumberInConstraintFromSource checks the `@in` constraint written in
// source (#285). `in` is a keyword, and before the fix the constraint could
// not be parsed at all, so it was reachable only from Go.
func (r *RuntimeTestSuite) TestNumberInConstraintFromSource() {
	ctx := r.T().Context()
	src := `namespace com/app

policy pol {
  fact n: number @in(1, 2)
  let one: number @in(1, 2) = 1
  let three: number @in(1, 2) = 3
  rule accepts = { yield one == 1 }
  rule rejects = { yield three == 3 }
  rule from_fact = { yield n > 0 }
  rule is_check = { yield 1 is number @in(1, 2) and not (3 is number @in(1, 2)) }
  export decision of accepts
  export decision of rejects
  export decision of from_fact
  export decision of is_check
}
`
	_, exec := r.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "in.sentrie", src: src})
	facts := map[string]any{"n": 1}

	out, err := exec.ExecRule(ctx, "com/app", "pol", "accepts", facts)
	r.Require().NoError(err)
	r.Equal(trinary.True, out.Decision.State)

	out, err = exec.ExecRule(ctx, "com/app", "pol", "is_check", facts)
	r.Require().NoError(err)
	r.Equal(trinary.True, out.Decision.State)

	out, err = exec.ExecRule(ctx, "com/app", "pol", "from_fact", facts)
	r.Require().NoError(err)
	r.Equal(trinary.True, out.Decision.State)

	_, err = exec.ExecRule(ctx, "com/app", "pol", "rejects", facts)
	r.Require().Error(err)
	r.Contains(err.Error(), "value 3 is not in the set")

	_, err = exec.ExecRule(ctx, "com/app", "pol", "from_fact", map[string]any{"n": 3})
	r.Require().Error(err)
}

const numberInPositionsSource = `namespace com/app

shape Roll { value: number @in(1, 2, 3) }
shape Pick number @in(1, 2)

policy pol {
  fact xs: list[number @in(1, 2)]
  fact payload: document
  fact n: number
  shape Local { pick: number @in(7, 8) }
  rule list_ok = { yield count(xs) > 0 }
  rule field_conforms = { yield payload is Roll }
  rule field_malformed = { yield payload is not Roll }
  rule alias_is = { yield n is Pick }
  rule alias_is_not = { yield n is not Pick }
  rule inline_is = { yield n is number @in(1, 2) }
  rule inline_is_not = { yield n is not number @in(1, 2) }
  rule list_elem_is = { yield [n] is list[number @in(1, 2)] }
  rule policy_shape = { yield { "pick": 7 } is Local and { "pick": 9 } is not Local }
  rule derive_ok = { yield pickOne(1) == 1 }
  rule derive_bad = { yield pickOne(2) == 2 }
  export decision of list_ok
  export decision of field_conforms
  export decision of field_malformed
  export decision of alias_is
  export decision of alias_is_not
  export decision of inline_is
  export decision of inline_is_not
  export decision of list_elem_is
  export decision of policy_shape
  export decision of derive_ok
  export decision of derive_bad
}

derive pickOne = (v: number @in(1)): number => { yield v }
`

// TestNumberInConstraintInEveryTypePosition checks that `@in` is accepted and
// enforced wherever a type can carry constraints (#285): a list element, a
// shape field, a shape alias, a policy-local shape, a derive parameter, a
// fact, and the right-hand type of `is` and `is not`.
func (r *RuntimeTestSuite) TestNumberInConstraintInEveryTypePosition() {
	ctx := r.T().Context()
	_, exec := r.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "in_positions.sentrie", src: numberInPositionsSource})

	decide := func(rule string, facts map[string]any) trinary.Value {
		out, err := exec.ExecRule(ctx, "com/app", "pol", rule, facts)
		r.Require().NoError(err, rule)
		return out.Decision.State
	}

	for _, tc := range []struct {
		n       int
		inSet   bool
		payload map[string]any
		xs      []any
	}{
		{n: 1, inSet: true, payload: map[string]any{"value": 2}, xs: []any{1, 2}},
		{n: 3, inSet: false, payload: map[string]any{"value": 4}, xs: []any{1, 2}},
	} {
		facts := map[string]any{"n": tc.n, "payload": tc.payload, "xs": tc.xs}
		want := trinary.From(tc.inSet)
		unwant := trinary.From(!tc.inSet)
		r.Equal(want, decide("field_conforms", facts), "shape field, n=%d", tc.n)
		r.Equal(unwant, decide("field_malformed", facts), "shape field is not, n=%d", tc.n)
		r.Equal(want, decide("alias_is", facts), "shape alias, n=%d", tc.n)
		r.Equal(unwant, decide("alias_is_not", facts), "shape alias is not, n=%d", tc.n)
		r.Equal(want, decide("inline_is", facts), "inline is, n=%d", tc.n)
		r.Equal(unwant, decide("inline_is_not", facts), "inline is not, n=%d", tc.n)
		r.Equal(want, decide("list_elem_is", facts), "list element, n=%d", tc.n)
		r.Equal(trinary.True, decide("policy_shape", facts), "policy-local shape")
		r.Equal(trinary.True, decide("derive_ok", facts), "derive parameter in set")
	}

	// A fact typed list[number @in(...)] rejects an element outside the set.
	base := map[string]any{"n": 1, "payload": map[string]any{"value": 1}}
	r.Equal(trinary.True, decide("list_ok", withFact(base, "xs", []any{1, 2})))
	_, err := exec.ExecRule(ctx, "com/app", "pol", "list_ok", withFact(base, "xs", []any{1, 3}))
	r.Require().Error(err)
	r.Contains(err.Error(), "value 3 is not in the set")

	// A derive parameter typed number @in(1) rejects 2.
	_, err = exec.ExecRule(ctx, "com/app", "pol", "derive_bad", withFact(base, "xs", []any{1}))
	r.Require().Error(err)
	r.Contains(err.Error(), "value 2 is not in the set")
}

func withFact(base map[string]any, name string, value any) map[string]any {
	out := map[string]any{name: value}
	for k, v := range base {
		out[k] = v
	}
	return out
}
