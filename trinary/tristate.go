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

package trinary

import (
	"encoding/json"
	"reflect"

	"github.com/sentrie-sh/sentrie/tokens"
)

type HasTrinary interface {
	ToTrinary() Value
}

type IsUndefined interface {
	IsUndefined() bool
}

// Value represents a trinary outcome: True, False, or Unknown.
type Value int

const (
	False Value = iota - 1
	Unknown
	True
)

func (r Value) String() string {
	switch r {
	case True:
		return "true"
	case False:
		return "false"
	case Unknown:
		return "unknown"
	default:
		return "unknown"
	}
}

func (r Value) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.String())
}

// Not implements logical NOT.
// True -> False, False -> True, Unknown -> Unknown
func (r Value) Not() Value {
	switch r {
	case True:
		return False
	case False:
		return True
	default:
		return Unknown
	}
}

// And implements tri-state AND using Kleene logic.
// | **AND**     | **True** | **False** | **Unknown** |
// | ----------- | -------- | --------- | ----------- |
// | **True**    | True     | False     | Unknown     |
// | **False**   | False    | False     | False       |
// | **Unknown** | Unknown  | False     | Unknown     |
func (r Value) And(other Value) Value {
	switch r {
	case True:
		// True ∧ x = x
		return other
	case False:
		// False ∧ _ = False
		return False
	case Unknown:
		// Unknown ∧ True = Unknown
		// Unknown ∧ False = False
		// Unknown ∧ Unknown = Unknown
		switch other {
		case True:
			return Unknown
		case False:
			return False
		default:
			return Unknown
		}
	default:
		return Unknown
	}
}

// Or implements tri-state OR as a Kleene logic.
// | **OR**      | **True** | **False** | **Unknown** |
// | ----------- | -------- | --------- | ----------- |
// | **True**    | True     | True      | True        |
// | **False**   | True     | False     | Unknown     |
// | **Unknown** | True     | Unknown   | Unknown     |
func (r Value) Or(other Value) Value {
	switch r {
	case True:
		// True ∨ _ = True
		return True
	case False:
		// False ∨ x = x
		return other
	case Unknown:
		// Unknown ∨ True = True
		// Unknown ∨ False = Unknown
		// Unknown ∨ Unknown = Unknown
		switch other {
		case True:
			return True
		case False:
			return Unknown
		default:
			return Unknown
		}
	default:
		return Unknown
	}
}

func (r Value) Equals(other Value) bool {
	return r == other
}

// Istrue returns true if the value is Pass
func (r Value) IsTrue() bool {
	return r == True
}

func FromToken(t tokens.Instance) Value {
	switch t.Kind {
	case tokens.KeywordTrue:
		return True
	case tokens.KeywordFalse:
		return False
	case tokens.KeywordUnknown:
		return Unknown
	default:
		return False
	}
}

// FromAny coerces any Go value into a Trinary Value.
// - nil → Unknown
// - Trinary Value → itself
// - bool → True/False
// - *bool → True/False/Unknown
// - anything else → truthy test via IsTruthy (true→True, false→False)
func From(v any) Value {
	if v == nil {
		return Unknown
	}

	switch t := v.(type) {
	case HasTrinary:
		return t.ToTrinary()
	case Value:
		return t
	case bool:
		if t {
			return True
		}
		return False
	case *bool:
		if t == nil {
			return Unknown
		}
		if *t {
			return True
		}
		return False
	}

	if isTruthy(v) {
		return True
	}

	return False
}

// isTruthy checks if a value is truthy.
// - nil → false
// - IsUndefined → false
// - bool → as-is
// - string → len > 0
// - int, uint, float → != 0
// - slice, array → > len > 0
// - map → > len > 0
// - ptr, interface → !IsNil()
// - struct → non-nil struct
// - default → true
func isTruthy(v any) bool {
	if v == nil {
		return false
	}

	if u, ok := v.(HasTrinary); ok {
		return u.ToTrinary().IsTrue()
	}

	if u, ok := v.(IsUndefined); ok && u.IsUndefined() {
		return false
	}

	rv := reflect.ValueOf(v)

	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return rv.Uint() != 0
	case reflect.Float32, reflect.Float64:
		return rv.Float() != 0
	case reflect.Bool:
		return rv.Bool()
	case reflect.Slice, reflect.Array, reflect.Map, reflect.String:
		return rv.Len() > 0
	case reflect.Ptr, reflect.Interface:
		if rv.IsNil() {
			return false
		}
		// Deref once and re-evaluate
		return isTruthy(rv.Elem().Interface())
	}

	// Default: non-nil values are truthy
	return true
}
