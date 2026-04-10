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

	"github.com/sentrie-sh/sentrie/trinary"
)

func (s *BoxTestSuite) TestValueKindString_AllBranches() {
	s.Equal("invalid", ValueInvalid.String())
	s.Equal("undefined", ValueUndefined.String())
	s.Equal("null", ValueNull.String())
	s.Equal("bool", ValueBool.String())
	s.Equal("number", ValueNumber.String())
	s.Equal("string", ValueString.String())
	s.Equal("trinary", ValueTrinary.String())
	s.Equal("list", ValueList.String())
	s.Equal("dict", ValueDict.String())
	s.Equal("document", ValueDocument.String())
	s.Equal("callable", ValueCallable.String())
	s.Equal("invalid", ValueKind(255).String())
}

func (s *BoxTestSuite) TestTryToBoundaryAnyCallableErrors() {
	_, err := TryToBoundaryAny(Callable(struct{}{}))
	s.ErrorIs(err, ErrCallableBoundary)
}

func (s *BoxTestSuite) TestValuePredicatesAndAliases() {
	type doc struct{ x int }
	shared := &doc{x: 1}
	other := &doc{x: 1}
	s.False(Value{}.IsValid())
	s.True(Number(1).IsValid())
	s.True(Undefined().IsUndefined())
	s.False(Null().IsUndefined())
	s.True(Null().IsNull())
	s.False(Undefined().IsNull())
	s.True(Document(shared).SameDocumentRef(Document(shared)))
	s.False(Document(shared).SameDocumentRef(Document(other)))
	s.False(Document(shared).SameDocumentRef(Number(1)))
	s.True(Object(shared).SameObjectRef(Object(shared)))
	ref, ok := Object(shared).ObjectRef()
	s.True(ok)
	s.Same(shared, ref)
}

func (s *BoxTestSuite) TestAccessorsWithWrongKindOrMalformedPayload() {
	_, ok := Number(1).BoolValue()
	s.False(ok)
	_, ok = Bool(true).TrinaryValue()
	s.False(ok)
	_, ok = Number(1).DocumentRef()
	s.False(ok)
	_, ok = Number(1).ListValue()
	s.False(ok)
	validList, ok := List([]Value{Number(1)}).ListValue()
	s.True(ok)
	s.Len(validList, 1)
	list := Value{kind: ValueList, ref: "not-a-list"}
	gotList, ok := list.ListValue()
	s.False(ok)
	s.Nil(gotList)
	m := Value{kind: ValueDict, ref: "not-a-map"}
	gotMap, ok := m.DictValue()
	s.False(ok)
	s.Nil(gotMap)
}

func (s *BoxTestSuite) TestAnyAndString_WithMalformedAndInvalidKinds() {
	s.Equal([]any{}, Value{kind: ValueList, ref: "bad"}.Any())
	s.Equal(map[string]any{}, Value{kind: ValueDict, ref: "bad"}.Any())
	s.Nil(Value{kind: ValueInvalid}.Any())
	s.Nil(Value{kind: ValueKind(254)}.Any())
	s.Equal("invalid", Value{kind: ValueInvalid}.String())
	s.Equal("false", Bool(false).String())
	s.Equal("true", Bool(true).String())
}

func (s *BoxTestSuite) TestTrinaryFrom_CoversInvalidAndDefaultKinds() {
	s.Equal(trinary.Unknown, TrinaryFrom(Value{kind: ValueInvalid}))
	s.Equal(trinary.Unknown, TrinaryFrom(Value{kind: ValueKind(253)}))
}

func (s *BoxTestSuite) TestMarshalJSON_FromAny_AndBoundaries() {
	b, err := Undefined().MarshalJSON()
	s.Require().NoError(err)
	s.Equal("null", string(b))
	b, err = List([]Value{String("x"), Number(1)}).MarshalJSON()
	s.Require().NoError(err)
	var decoded []any
	s.Require().NoError(json.Unmarshal(b, &decoded))
	s.Len(decoded, 2)
	s.Equal(Null(), FromAny(nil))
	s.Equal(Bool(true), FromAny(true))
	s.Equal(Number(int(1)), FromAny(int(1)))
	s.Equal(Number(int8(1)), FromAny(int8(1)))
	s.Equal(Number(int16(1)), FromAny(int16(1)))
	s.Equal(Number(int32(1)), FromAny(int32(1)))
	s.Equal(Number(int64(1)), FromAny(int64(1)))
	s.Equal(Number(uint(1)), FromAny(uint(1)))
	s.Equal(Number(uint8(1)), FromAny(uint8(1)))
	s.Equal(Number(uint16(1)), FromAny(uint16(1)))
	s.Equal(Number(uint32(1)), FromAny(uint32(1)))
	s.Equal(Number(uint64(1)), FromAny(uint64(1)))
	s.Equal(Number(float32(1.5)), FromAny(float32(1.5)))
	s.Equal(Number(float64(2.5)), FromAny(float64(2.5)))
	s.Equal(String("x"), FromAny("x"))
	s.Equal(Trinary(trinary.True), FromAny(trinary.True))
	valueInput := Number(9)
	s.Equal(valueInput, FromAny(valueInput))
	listInput := []Value{Number(1)}
	s.Equal(List(listInput), FromAny(listInput))
	mapInput := map[string]Value{"a": Number(1)}
	s.Equal(Dict(mapInput), FromAny(mapInput))
	s.Equal(List([]Value{Number(1), Bool(true)}), FromAny([]any{1, true}))
	s.Equal(Dict(map[string]Value{"a": Number(1)}), FromAny(map[string]any{"a": 1}))
	type hostDoc struct{ ID int }
	doc := hostDoc{ID: 7}
	s.Equal(Document(doc), FromAny(doc))
	undefinedBoundary := ToBoundaryAny(Undefined())
	s.True(IsBoundaryUndefined(undefinedBoundary))
	s.Equal(Undefined(), FromBoundaryAny(undefinedBoundary))
	roundTrip := FromBoundaryAny(ToBoundaryAny(Dict(map[string]Value{
		"a": Number(1),
		"b": List([]Value{Undefined(), String("x")}),
	})))
	s.True(EqualValues(Dict(map[string]Value{
		"a": Number(1),
		"b": List([]Value{Undefined(), String("x")}),
	}), roundTrip))
	s.False(IsBoundaryUndefined(nil))
}

func (s *BoxTestSuite) TestToBoundaryAnyCallableDoesNotPanic() {
	nested := Dict(map[string]Value{
		"items": List([]Value{
			Number(1),
			Callable(struct{}{}),
		}),
	})

	s.NotPanics(func() {
		out := ToBoundaryAny(nested)
		m, ok := out.(map[string]any)
		s.True(ok)
		items, ok := m["items"].([]any)
		s.True(ok)
		s.Equal(1.0, items[0])
		s.Equal("<callable>", items[1])
	})
}

func (s *BoxTestSuite) TestCallableRenderingAndMarshalBehavior() {
	v := Callable(struct{}{})
	s.Equal("<callable>", v.Any())
	s.Equal("<callable>", v.String())

	_, err := v.MarshalJSON()
	s.Require().Error(err)
	s.Require().ErrorContains(err, "cannot marshal callable")
}

func (s *BoxTestSuite) TestEqualValues_InvalidKindFallsBackToFalse() {
	s.False(EqualValues(Value{kind: ValueInvalid}, Value{kind: ValueInvalid}))
	s.False(EqualValues(List([]Value{Number(1)}), List([]Value{Number(1), Number(2)})))
	s.True(EqualValues(Bool(true), Bool(true)))
	s.False(EqualValues(Bool(true), Bool(false)))
	s.True(EqualValues(Number(10), Number(10.0)))
	s.False(EqualValues(Number(10), Number(11)))
	s.True(EqualValues(String("x"), String("x")))
	s.False(EqualValues(String("x"), String("y")))
	s.True(EqualValues(Trinary(trinary.True), Trinary(trinary.True)))
	s.False(EqualValues(Trinary(trinary.True), Trinary(trinary.False)))
}
