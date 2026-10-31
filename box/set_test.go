// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package box

import (
	"encoding/json"
	"math"

	"github.com/sentrie-sh/sentrie/trinary"
)

func (s *BoxTestSuite) TestNewSetCanonicalOrderAndDedupe() {
	set, err := NewSet([]Value{
		Dict(map[string]Value{"b": Number(1)}),
		List([]Value{Number(2)}),
		String("x"),
		Number(1),
		Bool(true),
		Null(),
		Number(1),
		Bool(false),
		Dict(map[string]Value{"a": Number(1)}),
		List([]Value{Number(1), Number(2)}),
	})
	s.Require().NoError(err)
	s.True(set.IsSet())
	raw, err := json.Marshal(set)
	s.Require().NoError(err)
	s.Equal(`[null,false,true,1,"x",[1,2],[2],{"a":1},{"b":1}]`, string(raw))

	plain := List([]Value{Number(1)})
	s.False(plain.IsSet())
	s.False(Dict(nil).IsSet())
}

func (s *BoxTestSuite) TestNewSetUniqueRejectsDuplicates() {
	_, err := NewSetUnique([]Value{String("a"), String("b"), String("a")})
	s.EqualError(err, "duplicate set element a")

	set, err := NewSetUnique([]Value{Number(0.0), Number(1)})
	s.Require().NoError(err)
	_, err = NewSetUnique([]Value{Number(math.Copysign(0, -1)), Number(0)})
	s.Error(err, "-0 and 0 are equal, so they are duplicates")
	s.Equal(2, len(set.ref.([]Value)))
}

func (s *BoxTestSuite) TestNewSetRejectsUnorderedElements() {
	cases := map[string]Value{
		"a set cannot hold a value of kind undefined": Undefined(),
		"a set cannot hold an unknown value":          Trinary(trinary.Unknown),
		"a set cannot hold NaN":                       Number(math.NaN()),
		"a set cannot hold a value of kind document":  Document(struct{}{}),
		"a set cannot hold a value of kind callable":  Callable(nil),
	}
	for want, v := range cases {
		_, err := NewSet([]Value{v})
		s.EqualError(err, want)
		_, err = NewSet([]Value{List([]Value{v})})
		s.EqualError(err, want, "nested in a list")
		_, err = NewSet([]Value{Dict(map[string]Value{"k": v})})
		s.EqualError(err, want, "nested in a dict")
	}
}

// TestCompareSetElementsAgreesWithEqualValues guards the invariant that makes
// membership and duplicates well defined: the canonical order finds two
// elements equal exactly when EqualValues does.
func (s *BoxTestSuite) TestCompareSetElementsAgreesWithEqualValues() {
	vals := []Value{
		Null(), Bool(false), Bool(true), Number(0), Number(1), Number(-1),
		String(""), String("a"), List(nil), List([]Value{Number(1)}),
		List([]Value{Number(1), Number(2)}), Dict(nil),
		Dict(map[string]Value{"a": Number(1)}), Dict(map[string]Value{"a": Number(2)}),
		Dict(map[string]Value{"a": Number(1), "b": Null()}),
	}
	// A DecisionEnvelope equals a bool by its state under EqualValues, so it
	// is left out here: NewSet strips the tag from every element it holds.
	for _, a := range vals {
		for _, b := range vals {
			s.Equal(EqualValues(a, b), compareSetElements(a, b) == 0, "%v vs %v", a, b)
			s.Equal(compareSetElements(a, b), -compareSetElements(b, a), "%v vs %v antisymmetric", a, b)
		}
	}
}

func (s *BoxTestSuite) TestSetContainsMatchesLinearScan() {
	set, err := NewSet([]Value{String("b"), Number(2), Bool(true), DecisionEnvelope(map[string]Value{}, trinary.True)})
	s.Require().NoError(err)
	for _, needle := range []Value{String("b"), String("c"), Number(2), Number(3), Bool(true), Bool(false), Null(), Dict(map[string]Value{})} {
		want := false
		xs, _ := set.ListValue()
		for _, x := range xs {
			if EqualValues(x, needle) {
				want = true
			}
		}
		s.Equal(want, ContainsValue(set, needle), "%v", needle)
	}
}

func (s *BoxTestSuite) TestSetOperations() {
	mustSet := func(xs ...Value) Value {
		v, err := NewSet(xs)
		s.Require().NoError(err)
		return v
	}
	a := mustSet(Number(3), Number(1), Number(2), Number(1))
	b := mustSet(Number(2), Number(4))
	empty := mustSet()

	u, err := SetUnion(a, b)
	s.Require().NoError(err)
	s.Equal([]any{1.0, 2.0, 3.0, 4.0}, u.Any())
	s.True(u.IsSet())

	i, err := SetIntersection(a, b)
	s.Require().NoError(err)
	s.Equal([]any{2.0}, i.Any())

	d, err := SetDifference(a, b)
	s.Require().NoError(err)
	s.Equal([]any{1.0, 3.0}, d.Any())

	d, err = SetDifference(b, empty)
	s.Require().NoError(err)
	s.Equal([]any{2.0, 4.0}, d.Any())

	ok, err := SetIsSubset(mustSet(Number(2)), a)
	s.Require().NoError(err)
	s.True(ok)
	ok, err = SetIsSubset(b, a)
	s.Require().NoError(err)
	s.False(ok)
	ok, err = SetIsSubset(empty, empty)
	s.Require().NoError(err)
	s.True(ok)

	// the operations take sets only; ToSet is the conversion step
	_, err = SetUnion(List([]Value{Number(1)}), b)
	s.EqualError(err, "first operand is not a set")
	_, err = SetIsSubset(b, String("x"))
	s.EqualError(err, "second operand is not a set")
	_, err = ToSet(String("x"))
	s.EqualError(err, "value of kind string is not a set or list")
}

// TestSetCanonicalisesElementsAtEveryDepth covers the elements a set holds:
// an imported decision at any depth is its plain dict, so `[impT]` and
// `[true]` are two distinct elements whose held forms are not EqualValues;
// -0 at any depth is 0, so the output does not depend on input order.
func (s *BoxTestSuite) TestSetCanonicalisesElementsAtEveryDepth() {
	impT := DecisionEnvelope(map[string]Value{"state": String("true")}, trinary.True)
	set, err := NewSet([]Value{List([]Value{impT}), List([]Value{Bool(true)})})
	s.Require().NoError(err)
	xs, _ := set.ListValue()
	s.Require().Len(xs, 2)
	for i := range xs {
		for j := range xs {
			s.Equal(i == j, EqualValues(xs[i], xs[j]), "held elements %d and %d", i, j)
		}
	}
	inner, _ := xs[1].ListValue()
	_, tagged := inner[0].DecisionState()
	s.False(tagged, "a nested decision must be held as a plain dict")

	deep := Dict(map[string]Value{"k": List([]Value{impT})})
	set, err = NewSet([]Value{deep})
	s.Require().NoError(err)
	xs, _ = set.ListValue()
	m, _ := xs[0].DictValue()
	l, _ := m["k"].ListValue()
	_, tagged = l[0].DecisionState()
	s.False(tagged, "a decision inside a dict inside a list must be held as a plain dict")

	negZero := Number(math.Copysign(0, -1))
	for _, order := range [][]Value{
		{negZero, Number(0), List([]Value{negZero}), List([]Value{Number(0)})},
		{Number(0), negZero, List([]Value{Number(0)}), List([]Value{negZero})},
	} {
		set, err := NewSet(order)
		s.Require().NoError(err)
		raw, err := json.Marshal(set)
		s.Require().NoError(err)
		s.Equal(`[0,[0]]`, string(raw))
	}

	// a tagged list still holding -0 forces the rewrite of a nested set
	nested := Value{kind: ValueList, u64: setTag, ref: []Value{negZero}}
	set, err = NewSet([]Value{nested})
	s.Require().NoError(err)
	xs, _ = set.ListValue()
	s.True(xs[0].IsSet(), "a nested set keeps its tag when its elements are rewritten")
}

// TestSetTagDoesNotSurviveRebuild checks that a set is a list everywhere it
// crosses a boundary and that a rebuilt list is plain.
func (s *BoxTestSuite) TestSetTagDoesNotSurviveRebuild() {
	set, err := NewSet([]Value{String("b"), String("a")})
	s.Require().NoError(err)
	s.Equal(ValueList, set.Kind())
	b, err := TryToBoundaryAny(set)
	s.Require().NoError(err)
	s.Equal([]any{"a", "b"}, b)
	s.False(FromBoundaryAny(b).IsSet())
	xs, _ := set.ListValue()
	s.False(List(xs).IsSet())
}
