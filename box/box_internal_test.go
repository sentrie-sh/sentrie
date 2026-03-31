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

package box

import (
	"encoding/json"
	"testing"

	"github.com/sentrie-sh/sentrie/trinary"
	"github.com/stretchr/testify/require"
)

func TestValueKindString_AllBranches(t *testing.T) {
	require.Equal(t, "invalid", ValueInvalid.String())
	require.Equal(t, "undefined", ValueUndefined.String())
	require.Equal(t, "null", ValueNull.String())
	require.Equal(t, "bool", ValueBool.String())
	require.Equal(t, "number", ValueNumber.String())
	require.Equal(t, "string", ValueString.String())
	require.Equal(t, "trinary", ValueTrinary.String())
	require.Equal(t, "list", ValueList.String())
	require.Equal(t, "map", ValueMap.String())
	require.Equal(t, "document", ValueDocument.String())
	require.Equal(t, "invalid", ValueKind(255).String())
}

func TestValuePredicatesAndAliases(t *testing.T) {
	type doc struct{ x int }
	shared := &doc{x: 1}
	other := &doc{x: 1}

	require.False(t, Value{}.IsValid())
	require.True(t, Number(1).IsValid())
	require.True(t, Undefined().IsUndefined())
	require.False(t, Null().IsUndefined())
	require.True(t, Null().IsNull())
	require.False(t, Undefined().IsNull())

	require.True(t, Document(shared).SameDocumentRef(Document(shared)))
	require.False(t, Document(shared).SameDocumentRef(Document(other)))
	require.False(t, Document(shared).SameDocumentRef(Number(1)))

	require.True(t, Object(shared).SameObjectRef(Object(shared)))
	ref, ok := Object(shared).ObjectRef()
	require.True(t, ok)
	require.Same(t, shared, ref)
}

func TestAccessorsWithWrongKindOrMalformedPayload(t *testing.T) {
	_, ok := Number(1).BoolValue()
	require.False(t, ok)
	_, ok = Bool(true).TrinaryValue()
	require.False(t, ok)
	_, ok = Number(1).DocumentRef()
	require.False(t, ok)
	_, ok = Number(1).ListValue()
	require.False(t, ok)
	validList, ok := List([]Value{Number(1)}).ListValue()
	require.True(t, ok)
	require.Len(t, validList, 1)

	list := Value{kind: ValueList, ref: "not-a-list"}
	gotList, ok := list.ListValue()
	require.False(t, ok)
	require.Nil(t, gotList)

	m := Value{kind: ValueMap, ref: "not-a-map"}
	gotMap, ok := m.MapValue()
	require.False(t, ok)
	require.Nil(t, gotMap)
}

func TestAnyAndString_WithMalformedAndInvalidKinds(t *testing.T) {
	require.Equal(t, []any{}, Value{kind: ValueList, ref: "bad"}.Any())
	require.Equal(t, map[string]any{}, Value{kind: ValueMap, ref: "bad"}.Any())
	require.Nil(t, Value{kind: ValueInvalid}.Any())
	require.Nil(t, Value{kind: ValueKind(254)}.Any())

	require.Equal(t, "invalid", Value{kind: ValueInvalid}.String())
	require.Equal(t, "false", Bool(false).String())
	require.Equal(t, "true", Bool(true).String())
}

func TestTrinaryFrom_CoversInvalidAndDefaultKinds(t *testing.T) {
	require.Equal(t, trinary.Unknown, TrinaryFrom(Value{kind: ValueInvalid}))
	require.Equal(t, trinary.Unknown, TrinaryFrom(Value{kind: ValueKind(253)}))
}

func TestMarshalJSON_FromAny_AndBoundaries(t *testing.T) {
	b, err := Undefined().MarshalJSON()
	require.NoError(t, err)
	require.Equal(t, "null", string(b))

	b, err = List([]Value{String("x"), Number(1)}).MarshalJSON()
	require.NoError(t, err)
	var decoded []any
	require.NoError(t, json.Unmarshal(b, &decoded))
	require.Len(t, decoded, 2)

	require.Equal(t, Null(), FromAny(nil))
	require.Equal(t, Bool(true), FromAny(true))
	require.Equal(t, Number(int(1)), FromAny(int(1)))
	require.Equal(t, Number(int8(1)), FromAny(int8(1)))
	require.Equal(t, Number(int16(1)), FromAny(int16(1)))
	require.Equal(t, Number(int32(1)), FromAny(int32(1)))
	require.Equal(t, Number(int64(1)), FromAny(int64(1)))
	require.Equal(t, Number(uint(1)), FromAny(uint(1)))
	require.Equal(t, Number(uint8(1)), FromAny(uint8(1)))
	require.Equal(t, Number(uint16(1)), FromAny(uint16(1)))
	require.Equal(t, Number(uint32(1)), FromAny(uint32(1)))
	require.Equal(t, Number(uint64(1)), FromAny(uint64(1)))
	require.Equal(t, Number(float32(1.5)), FromAny(float32(1.5)))
	require.Equal(t, Number(float64(2.5)), FromAny(float64(2.5)))
	require.Equal(t, String("x"), FromAny("x"))
	require.Equal(t, Trinary(trinary.True), FromAny(trinary.True))

	valueInput := Number(9)
	require.Equal(t, valueInput, FromAny(valueInput))

	listInput := []Value{Number(1)}
	require.Equal(t, List(listInput), FromAny(listInput))
	mapInput := map[string]Value{"a": Number(1)}
	require.Equal(t, Map(mapInput), FromAny(mapInput))

	require.Equal(t, List([]Value{Number(1), Bool(true)}), FromAny([]any{1, true}))
	require.Equal(t, Map(map[string]Value{"a": Number(1)}), FromAny(map[string]any{"a": 1}))

	type hostDoc struct{ ID int }
	doc := hostDoc{ID: 7}
	require.Equal(t, Document(doc), FromAny(doc))

	undefinedBoundary := ToBoundaryAny(Undefined())
	require.True(t, IsBoundaryUndefined(undefinedBoundary))
	require.Equal(t, Undefined(), FromBoundaryAny(undefinedBoundary))

	roundTrip := FromBoundaryAny(ToBoundaryAny(Map(map[string]Value{
		"a": Number(1),
		"b": List([]Value{Undefined(), String("x")}),
	})))
	require.True(t, EqualValues(Map(map[string]Value{
		"a": Number(1),
		"b": List([]Value{Undefined(), String("x")}),
	}), roundTrip))

	require.False(t, IsBoundaryUndefined(nil))
}

func TestEqualValues_InvalidKindFallsBackToFalse(t *testing.T) {
	require.False(t, EqualValues(Value{kind: ValueInvalid}, Value{kind: ValueInvalid}))
	require.False(t, EqualValues(List([]Value{Number(1)}), List([]Value{Number(1), Number(2)})))
	require.True(t, EqualValues(Bool(true), Bool(true)))
	require.False(t, EqualValues(Bool(true), Bool(false)))
	require.True(t, EqualValues(Number(10), Number(10.0)))
	require.False(t, EqualValues(Number(10), Number(11)))
	require.True(t, EqualValues(String("x"), String("x")))
	require.False(t, EqualValues(String("x"), String("y")))
	require.True(t, EqualValues(Trinary(trinary.True), Trinary(trinary.True)))
	require.False(t, EqualValues(Trinary(trinary.True), Trinary(trinary.False)))
}
