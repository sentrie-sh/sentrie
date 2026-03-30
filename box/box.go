// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2025 Binaek Sarkar
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
	"fmt"
	"math"

	"github.com/sentrie-sh/sentrie/trinary"
)

type undefinedBoundaryToken struct{}

var boundaryUndefined = undefinedBoundaryToken{}

type ValueKind uint8

const (
	ValueInvalid ValueKind = iota
	ValueUndefined
	ValueNull
	ValueBool
	ValueNumber
	ValueString
	ValueTrinary
	ValueList
	ValueMap
	ValueDocument
	// ValueObject is a backward-compatible alias for ValueDocument.
	ValueObject = ValueDocument
)

func (k ValueKind) String() string {
	switch k {
	case ValueUndefined:
		return "undefined"
	case ValueNull:
		return "null"
	case ValueBool:
		return "bool"
	case ValueNumber:
		return "number"
	case ValueString:
		return "string"
	case ValueTrinary:
		return "trinary"
	case ValueList:
		return "list"
	case ValueMap:
		return "map"
	case ValueDocument:
		return "document"
	default:
		return "invalid"
	}
}

type Value struct {
	kind ValueKind
	u64  uint64
	ref  any
}

func Undefined() Value { return Value{kind: ValueUndefined} }
func Null() Value      { return Value{kind: ValueNull} }

func Bool[T ~bool](x T) Value {
	if x {
		return Value{kind: ValueBool, u64: 1}
	}
	return Value{kind: ValueBool}
}

func Number[T ~int | ~int8 | ~int16 | ~int32 | ~int64 |
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
	~float32 | ~float64](x T) Value {
	return Value{
		kind: ValueNumber,
		u64:  math.Float64bits(float64(x)),
	}
}

func String[T ~string](x T) Value {
	return Value{kind: ValueString, ref: string(x)}
}

func Trinary(x trinary.Value) Value {
	return Value{kind: ValueTrinary, u64: uint64(x)}
}

func List(xs []Value) Value {
	return Value{kind: ValueList, ref: xs}
}

func Map(m map[string]Value) Value {
	return Value{kind: ValueMap, ref: m}
}

func Document[T any](x T) Value {
	return Value{kind: ValueDocument, ref: x}
}

// Object is a backward-compatible alias for Document.
func Object[T any](x T) Value {
	return Document(x)
}

func (v Value) Kind() ValueKind   { return v.kind }
func (v Value) IsValid() bool     { return v.kind != ValueInvalid }
func (v Value) IsUndefined() bool { return v.kind == ValueUndefined }
func (v Value) IsNull() bool      { return v.kind == ValueNull }
func (v Value) SameDocumentRef(other Value) bool {
	if v.kind != ValueDocument || other.kind != ValueDocument {
		return false
	}
	return v.ref == other.ref
}

// SameObjectRef is a backward-compatible alias for SameDocumentRef.
func (v Value) SameObjectRef(other Value) bool {
	return v.SameDocumentRef(other)
}

func (v Value) BoolValue() (bool, bool) {
	if v.kind != ValueBool {
		return false, false
	}
	return v.u64 != 0, true
}

func (v Value) NumberValue() (float64, bool) {
	if v.kind != ValueNumber {
		return 0, false
	}
	return math.Float64frombits(v.u64), true
}

func (v Value) StringValue() (string, bool) {
	if v.kind != ValueString {
		return "", false
	}
	s, ok := v.ref.(string)
	return s, ok
}

func (v Value) TrinaryValue() (trinary.Value, bool) {
	if v.kind != ValueTrinary {
		return trinary.Unknown, false
	}
	return trinary.Value(v.u64), true
}

func (v Value) ListValue() ([]Value, bool) {
	if v.kind != ValueList {
		return nil, false
	}
	xs, ok := v.ref.([]Value)
	return xs, ok
}

func (v Value) MapValue() (map[string]Value, bool) {
	if v.kind != ValueMap {
		return nil, false
	}
	m, ok := v.ref.(map[string]Value)
	return m, ok
}

// DocumentRef returns the wrapped host document for ValueDocument. For other kinds it returns (nil, false).
func (v Value) DocumentRef() (any, bool) {
	if v.kind != ValueDocument {
		return nil, false
	}
	return v.ref, true
}

// ObjectRef is a backward-compatible alias for DocumentRef.
func (v Value) ObjectRef() (any, bool) {
	return v.DocumentRef()
}

// TrinaryFrom returns the Kleene trinary outcome for b, matching trinary.From(b.Any()) without
// materializing an intermediate any slice or map for Box values.
func TrinaryFrom(b Value) trinary.Value {
	switch b.kind {
	case ValueUndefined, ValueNull:
		return trinary.Unknown
	case ValueBool:
		return trinary.From(b.u64 != 0)
	case ValueNumber:
		return trinary.From(math.Float64frombits(b.u64))
	case ValueString:
		s, _ := b.ref.(string)
		return trinary.From(s)
	case ValueTrinary:
		return trinary.Value(b.u64)
	case ValueList:
		xs, _ := b.ref.([]Value)
		if len(xs) == 0 {
			return trinary.False
		}
		return trinary.True
	case ValueMap:
		m, _ := b.ref.(map[string]Value)
		if len(m) == 0 {
			return trinary.False
		}
		return trinary.True
	case ValueDocument:
		return trinary.From(b.ref)
	case ValueInvalid:
		return trinary.Unknown
	default:
		return trinary.Unknown
	}
}

func (v Value) Any() any {
	switch v.kind {
	case ValueUndefined:
		return nil
	case ValueNull:
		return nil
	case ValueBool:
		return v.u64 != 0
	case ValueNumber:
		return math.Float64frombits(v.u64)
	case ValueString:
		s, _ := v.ref.(string)
		return s
	case ValueTrinary:
		return trinary.Value(v.u64)
	case ValueList:
		xs, _ := v.ref.([]Value)
		out := make([]any, 0, len(xs))
		for _, x := range xs {
			out = append(out, x.Any())
		}
		return out
	case ValueMap:
		m, _ := v.ref.(map[string]Value)
		out := make(map[string]any, len(m))
		for k, x := range m {
			out[k] = x.Any()
		}
		return out
	case ValueDocument:
		return v.ref
	default:
		return nil
	}
}

func (v Value) String() string {
	switch v.kind {
	case ValueInvalid:
		return "invalid"
	case ValueUndefined:
		return "undefined"
	case ValueNull:
		return "null"
	case ValueBool:
		if b, _ := v.BoolValue(); b {
			return "true"
		}
		return "false"
	case ValueNumber:
		n, _ := v.NumberValue()
		return fmt.Sprintf("%v", n)
	case ValueString:
		s, _ := v.StringValue()
		return s
	case ValueTrinary:
		t, _ := v.TrinaryValue()
		return t.String()
	default:
		return fmt.Sprintf("%v", v.Any())
	}
}

func (v Value) MarshalJSON() ([]byte, error) {
	if v.IsUndefined() {
		return []byte("null"), nil
	}
	return json.Marshal(v.Any())
}

func FromAny(x any) Value {
	switch t := x.(type) {
	case nil:
		return Null()
	case Value:
		return t
	case bool:
		return Bool(t)
	case int:
		return Number(t)
	case int8:
		return Number(t)
	case int16:
		return Number(t)
	case int32:
		return Number(t)
	case int64:
		return Number(t)
	case uint:
		return Number(t)
	case uint8:
		return Number(t)
	case uint16:
		return Number(t)
	case uint32:
		return Number(t)
	case uint64:
		return Number(t)
	case float32:
		return Number(t)
	case float64:
		return Number(t)
	case string:
		return String(t)
	case trinary.Value:
		return Trinary(t)
	case []Value:
		return List(t)
	case map[string]Value:
		return Map(t)
	case []any:
		out := make([]Value, 0, len(t))
		for _, item := range t {
			out = append(out, FromAny(item))
		}
		return List(out)
	case map[string]any:
		out := make(map[string]Value, len(t))
		for k, v := range t {
			out[k] = FromAny(v)
		}
		return Map(out)
	default:
		return Document(x)
	}
}

// ToBoundaryAny converts a boxed Value into an unboxed representation suitable
// for runtime boundaries while preserving undefined/null distinction.
func ToBoundaryAny(v Value) any {
	switch v.Kind() {
	case ValueUndefined:
		return boundaryUndefined
	case ValueList:
		xs, _ := v.ListValue()
		out := make([]any, 0, len(xs))
		for _, item := range xs {
			out = append(out, ToBoundaryAny(item))
		}
		return out
	case ValueMap:
		m, _ := v.MapValue()
		out := make(map[string]any, len(m))
		for k, item := range m {
			out[k] = ToBoundaryAny(item)
		}
		return out
	default:
		return v.Any()
	}
}

// FromBoundaryAny converts runtime boundary values back into boxed Value while
// preserving undefined/null distinction.
func FromBoundaryAny(x any) Value {
	if _, ok := x.(undefinedBoundaryToken); ok {
		return Undefined()
	}
	switch t := x.(type) {
	case []any:
		out := make([]Value, 0, len(t))
		for _, item := range t {
			out = append(out, FromBoundaryAny(item))
		}
		return List(out)
	case map[string]any:
		out := make(map[string]Value, len(t))
		for k, item := range t {
			out[k] = FromBoundaryAny(item)
		}
		return Map(out)
	default:
		return FromAny(x)
	}
}

func IsBoundaryUndefined(x any) bool {
	_, ok := x.(undefinedBoundaryToken)
	return ok
}
