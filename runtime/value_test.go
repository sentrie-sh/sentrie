// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2026 Binaek Sarkar
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package runtime

import (
	"encoding/json"

	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/trinary"
)

type Value = box.Value
type ValueKind = box.ValueKind

const (
	ValueInvalid   = box.ValueInvalid
	ValueUndefined = box.ValueUndefined
	ValueNull      = box.ValueNull
	ValueBool      = box.ValueBool
	ValueNumber    = box.ValueNumber
	ValueString    = box.ValueString
	ValueTrinary   = box.ValueTrinary
	ValueList      = box.ValueList
	ValueMap       = box.ValueMap
	ValueObject    = box.ValueObject
)

var (
	Undefined       = box.Undefined
	Null            = box.Null
	Trinary         = box.Trinary
	List            = box.List
	Map             = box.Map
	FromAny         = box.FromAny
	ToBoundaryAny   = box.ToBoundaryAny
	FromBoundaryAny = box.FromBoundaryAny
)

func (s *RuntimeTestSuite) TestValueKindString() {
	s.Require().Equal("invalid", ValueKind(0).String())
	s.Require().Equal("undefined", ValueUndefined.String())
	s.Require().Equal("null", ValueNull.String())
	s.Require().Equal("bool", ValueBool.String())
	s.Require().Equal("number", ValueNumber.String())
	s.Require().Equal("string", ValueString.String())
	s.Require().Equal("trinary", ValueTrinary.String())
	s.Require().Equal("list", ValueList.String())
	s.Require().Equal("map", ValueMap.String())
	s.Require().Equal("document", ValueObject.String())
	s.Require().Equal("invalid", ValueKind(255).String())
}

func (s *RuntimeTestSuite) TestValueConstructorsAndAccessors() {
	u := Undefined()
	s.Require().True(u.IsUndefined())
	s.Require().True(u.IsValid())
	s.Require().Equal("undefined", u.String())
	_, ok := u.BoolValue()
	s.Require().False(ok)

	nl := Null()
	s.Require().True(nl.IsNull())
	s.Require().True(nl.IsValid())
	s.Require().Equal("null", nl.String())

	bt := box.Bool(true)
	bv, ok := bt.BoolValue()
	s.Require().True(ok)
	s.Require().True(bv)
	s.Require().Equal("true", bt.String())

	bf := box.Bool(false)
	bv, ok = bf.BoolValue()
	s.Require().True(ok)
	s.Require().False(bv)
	s.Require().Equal("false", bf.String())

	n := box.Number(42)
	nv, ok := n.NumberValue()
	s.Require().True(ok)
	s.Require().Equal(42.0, nv)
	s.Require().Equal("42", n.String())

	strVal := box.String("hello")
	sv, ok := strVal.StringValue()
	s.Require().True(ok)
	s.Require().Equal("hello", sv)
	s.Require().Equal("hello", strVal.String())

	tv := Trinary(trinary.False)
	tvv, ok := tv.TrinaryValue()
	s.Require().True(ok)
	s.Require().Equal(trinary.False, tvv)
	s.Require().Equal("false", tv.String())
}

func (s *RuntimeTestSuite) TestValueContainers() {
	list := List([]Value{box.Number(1), box.String("x")})
	xs, ok := list.ListValue()
	s.Require().True(ok)
	s.Require().Len(xs, 2)
	s.Require().Equal(1.0, xs[0].Any())
	s.Require().Equal("x", xs[1].Any())

	m := Map(map[string]Value{"a": box.Number(1), "b": box.Bool(true)})
	mv, ok := m.MapValue()
	s.Require().True(ok)
	s.Require().Equal(1.0, mv["a"].Any())
	s.Require().Equal(true, mv["b"].Any())

	obj := box.Object(struct{ Name string }{Name: "demo"})
	s.Require().Equal(ValueObject, obj.Kind())
	s.Require().Equal(struct{ Name string }{Name: "demo"}, obj.Any())
	s.Require().Equal("{demo}", obj.String())
}

func (s *RuntimeTestSuite) TestValueAnyUndefinedAndNull() {
	s.Require().Nil(Undefined().Any())
	s.Require().Nil(Null().Any())
	s.Require().Equal(trinary.False, Trinary(trinary.False).Any())
	s.Require().Equal(false, box.Bool(false).Any())
	s.Require().Equal(3.0, box.Number(3).Any())
	s.Require().Equal("s", box.String("s").Any())
}

func (s *RuntimeTestSuite) TestValueAnyAndFromAnyRoundTrip() {
	input := map[string]any{
		"a": nil,
		"b": true,
		"c": 12,
		"d": "s",
		"e": []any{1, "x", map[string]any{"nested": false}},
	}

	v := FromAny(input)
	s.Require().Equal(ValueMap, v.Kind())

	outAny := v.Any()
	outMap, ok := outAny.(map[string]any)
	s.Require().True(ok)
	s.Require().Contains(outMap, "a")
	s.Require().Contains(outMap, "e")

	reboxed := FromAny(v)
	s.Require().Equal(v.Kind(), reboxed.Kind())
}

func (s *RuntimeTestSuite) TestValueMarshalJSON() {
	v := Map(map[string]Value{
		"ok":   box.Bool(true),
		"num":  box.Number(3.14),
		"null": Null(),
	})
	b, err := json.Marshal(v)
	s.Require().NoError(err)
	s.Require().JSONEq(`{"null":null,"num":3.14,"ok":true}`, string(b))

	u, err := json.Marshal(Undefined())
	s.Require().NoError(err)
	s.Require().Equal("null", string(u))
}

func (s *RuntimeTestSuite) TestValueDefaultBranchesAndMismatches() {
	var invalid Value
	s.Require().Equal(ValueInvalid, invalid.Kind())
	s.Require().False(invalid.IsValid())
	s.Require().Equal("invalid", invalid.String())
	s.Require().Nil(invalid.Any())

	_, ok := box.String("x").NumberValue()
	s.Require().False(ok)
	_, ok = box.Number(1).StringValue()
	s.Require().False(ok)
	_, ok = box.Bool(true).ListValue()
	s.Require().False(ok)
	_, ok = box.Bool(true).MapValue()
	s.Require().False(ok)
	_, ok = box.Number(1).TrinaryValue()
	s.Require().False(ok)

	type custom struct{ X int }
	obj := custom{X: 9}
	s.Require().Equal(ValueObject, FromAny(obj).Kind())
}

func (s *RuntimeTestSuite) TestFromAnyNumericAndCollectionCases() {
	cases := []struct {
		in   any
		kind ValueKind
	}{
		{int(1), ValueNumber},
		{int8(1), ValueNumber},
		{int16(1), ValueNumber},
		{int32(1), ValueNumber},
		{int64(1), ValueNumber},
		{uint(1), ValueNumber},
		{uint8(1), ValueNumber},
		{uint16(1), ValueNumber},
		{uint32(1), ValueNumber},
		{uint64(1), ValueNumber},
		{float32(1.25), ValueNumber},
		{float64(2.5), ValueNumber},
		{true, ValueBool},
		{"x", ValueString},
		{trinary.True, ValueTrinary},
		{[]Value{box.Number(1)}, ValueList},
		{map[string]Value{"a": box.Number(1)}, ValueMap},
		{[]any{1, 2, 3}, ValueList},
		{map[string]any{"a": 1}, ValueMap},
	}

	for _, tc := range cases {
		v := FromAny(tc.in)
		s.Require().Equal(tc.kind, v.Kind())
	}
}

func (s *RuntimeTestSuite) TestBoundaryAnyRoundTripNestedContainers() {
	in := Map(map[string]Value{
		"a": Undefined(),
		"b": List([]Value{
			box.Number(1),
			Map(map[string]Value{
				"nested": Undefined(),
				"ok":     box.String("x"),
			}),
		}),
	})

	boundary := ToBoundaryAny(in)
	out := FromBoundaryAny(boundary)
	outMap, ok := out.MapValue()
	s.Require().True(ok)
	s.Require().True(outMap["a"].IsUndefined())
	bList, ok := outMap["b"].ListValue()
	s.Require().True(ok)
	nestedMap, ok := bList[1].MapValue()
	s.Require().True(ok)
	s.Require().True(nestedMap["nested"].IsUndefined())
	s.Require().Equal("x", nestedMap["ok"].Any())
}

func (s *RuntimeTestSuite) TestToBoundaryAnyPassthroughScalars() {
	s.Require().Equal(1.0, ToBoundaryAny(box.Number(1)))
	s.Require().Equal("x", ToBoundaryAny(box.String("x")))
	s.Require().Equal(true, ToBoundaryAny(box.Bool(true)))
}

func (s *RuntimeTestSuite) TestFromBoundaryAnyHandlesUndefinedToken() {
	v := FromBoundaryAny(ToBoundaryAny(Undefined()))
	s.Require().True(v.IsUndefined())
}
