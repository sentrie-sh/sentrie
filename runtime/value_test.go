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
	"testing"

	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/trinary"
	"github.com/stretchr/testify/require"
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

func TestValueKindString(t *testing.T) {
	require.Equal(t, "invalid", ValueKind(0).String())
	require.Equal(t, "undefined", ValueUndefined.String())
	require.Equal(t, "null", ValueNull.String())
	require.Equal(t, "bool", ValueBool.String())
	require.Equal(t, "number", ValueNumber.String())
	require.Equal(t, "string", ValueString.String())
	require.Equal(t, "trinary", ValueTrinary.String())
	require.Equal(t, "list", ValueList.String())
	require.Equal(t, "map", ValueMap.String())
	require.Equal(t, "document", ValueObject.String())
	require.Equal(t, "invalid", ValueKind(255).String())
}

func TestValueConstructorsAndAccessors(t *testing.T) {
	u := Undefined()
	require.True(t, u.IsUndefined())
	require.True(t, u.IsValid())
	require.Equal(t, "undefined", u.String())
	_, ok := u.BoolValue()
	require.False(t, ok)

	nl := Null()
	require.True(t, nl.IsNull())
	require.True(t, nl.IsValid())
	require.Equal(t, "null", nl.String())

	bt := box.Bool(true)
	bv, ok := bt.BoolValue()
	require.True(t, ok)
	require.True(t, bv)
	require.Equal(t, "true", bt.String())

	bf := box.Bool(false)
	bv, ok = bf.BoolValue()
	require.True(t, ok)
	require.False(t, bv)
	require.Equal(t, "false", bf.String())

	n := box.Number(42)
	nv, ok := n.NumberValue()
	require.True(t, ok)
	require.Equal(t, 42.0, nv)
	require.Equal(t, "42", n.String())

	s := box.String("hello")
	sv, ok := s.StringValue()
	require.True(t, ok)
	require.Equal(t, "hello", sv)
	require.Equal(t, "hello", s.String())

	tv := Trinary(trinary.False)
	tvv, ok := tv.TrinaryValue()
	require.True(t, ok)
	require.Equal(t, trinary.False, tvv)
	require.Equal(t, "false", tv.String())
}

func TestValueContainers(t *testing.T) {
	list := List([]Value{box.Number(1), box.String("x")})
	xs, ok := list.ListValue()
	require.True(t, ok)
	require.Len(t, xs, 2)
	require.Equal(t, 1.0, xs[0].Any())
	require.Equal(t, "x", xs[1].Any())

	m := Map(map[string]Value{"a": box.Number(1), "b": box.Bool(true)})
	mv, ok := m.MapValue()
	require.True(t, ok)
	require.Equal(t, 1.0, mv["a"].Any())
	require.Equal(t, true, mv["b"].Any())

	obj := box.Object(struct{ Name string }{Name: "demo"})
	require.Equal(t, ValueObject, obj.Kind())
	require.Equal(t, struct{ Name string }{Name: "demo"}, obj.Any())
	require.Equal(t, "{demo}", obj.String())
}

func TestValueAnyUndefinedAndNull(t *testing.T) {
	require.Nil(t, Undefined().Any())
	require.Nil(t, Null().Any())
	require.Equal(t, trinary.False, Trinary(trinary.False).Any())
	require.Equal(t, false, box.Bool(false).Any())
	require.Equal(t, 3.0, box.Number(3).Any())
	require.Equal(t, "s", box.String("s").Any())
}

func TestValueAnyAndFromAnyRoundTrip(t *testing.T) {
	input := map[string]any{
		"a": nil,
		"b": true,
		"c": 12,
		"d": "s",
		"e": []any{1, "x", map[string]any{"nested": false}},
	}

	v := FromAny(input)
	require.Equal(t, ValueMap, v.Kind())

	outAny := v.Any()
	outMap, ok := outAny.(map[string]any)
	require.True(t, ok)
	require.Contains(t, outMap, "a")
	require.Contains(t, outMap, "e")

	reboxed := FromAny(v)
	require.Equal(t, v.Kind(), reboxed.Kind())
}

func TestValueMarshalJSON(t *testing.T) {
	v := Map(map[string]Value{
		"ok":   box.Bool(true),
		"num":  box.Number(3.14),
		"null": Null(),
	})
	b, err := json.Marshal(v)
	require.NoError(t, err)
	require.JSONEq(t, `{"null":null,"num":3.14,"ok":true}`, string(b))

	u, err := json.Marshal(Undefined())
	require.NoError(t, err)
	require.Equal(t, "null", string(u))
}

func TestValueDefaultBranchesAndMismatches(t *testing.T) {
	var invalid Value
	require.Equal(t, ValueInvalid, invalid.Kind())
	require.False(t, invalid.IsValid())
	require.Equal(t, "invalid", invalid.String())
	require.Nil(t, invalid.Any())

	_, ok := box.String("x").NumberValue()
	require.False(t, ok)
	_, ok = box.Number(1).StringValue()
	require.False(t, ok)
	_, ok = box.Bool(true).ListValue()
	require.False(t, ok)
	_, ok = box.Bool(true).MapValue()
	require.False(t, ok)
	_, ok = box.Number(1).TrinaryValue()
	require.False(t, ok)

	type custom struct{ X int }
	obj := custom{X: 9}
	require.Equal(t, ValueObject, FromAny(obj).Kind())
}

func TestFromAnyNumericAndCollectionCases(t *testing.T) {
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
		require.Equal(t, tc.kind, v.Kind())
	}
}

func TestBoundaryAnyRoundTripNestedContainers(t *testing.T) {
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
	require.True(t, ok)
	require.True(t, outMap["a"].IsUndefined())
	bList, ok := outMap["b"].ListValue()
	require.True(t, ok)
	nestedMap, ok := bList[1].MapValue()
	require.True(t, ok)
	require.True(t, nestedMap["nested"].IsUndefined())
	require.Equal(t, "x", nestedMap["ok"].Any())
}

func TestToBoundaryAnyPassthroughScalars(t *testing.T) {
	require.Equal(t, 1.0, ToBoundaryAny(box.Number(1)))
	require.Equal(t, "x", ToBoundaryAny(box.String("x")))
	require.Equal(t, true, ToBoundaryAny(box.Bool(true)))
}

func TestFromBoundaryAnyHandlesUndefinedToken(t *testing.T) {
	v := FromBoundaryAny(ToBoundaryAny(Undefined()))
	require.True(t, v.IsUndefined())
}
