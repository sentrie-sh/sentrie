// SPDX-License-Identifier: Apache-2.0

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
	"strings"

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

// From coerces any Go value into a trinary Value:
//   - nil or IsUndefined → Unknown
//   - HasTrinary / Value → delegates directly
//   - bool → True/False
//   - numeric primitives (int/uint/float families) → True when != 0
//   - string → textual keywords first, otherwise True when len > 0
//   - pointers and interfaces → dereference once and retry
//   - reflected string/slice/array/map kinds → reuse the string/len rules
//   - everything else defaults to True (non-nil structs, etc.)
func From(v any) Value {
	// Check for actual nil first (returns Unknown)
	if v == nil {
		return Unknown
	}

	if u, ok := v.(IsUndefined); ok && u.IsUndefined() {
		return Unknown
	}

	switch t := v.(type) {
	case HasTrinary:
		return t.ToTrinary()
	case Value:
		return t
	case bool:
		return boolToValue(t)
	case *bool:
		if t == nil {
			return Unknown
		}
		return boolToValue(*t)
	case int, int8, int16, int32, int64:
		return boolToValue(reflect.ValueOf(t).Int() != 0)
	case uint, uint8, uint16, uint32, uint64, uintptr:
		return boolToValue(reflect.ValueOf(t).Uint() != 0)
	case float32, float64:
		return boolToValue(reflect.ValueOf(t).Float() != 0)
	case string:
		return stringToValue(t)
	default:
		rv := reflect.ValueOf(v)
		switch rv.Kind() {
		case reflect.Ptr:
			if rv.IsNil() {
				if rv.Type().Elem().Kind() == reflect.Bool {
					return Unknown
				}
				return False
			}
			// Deref once and re-evaluate
			return From(rv.Elem().Interface())
		case reflect.Interface:
			if rv.IsNil() {
				return Unknown
			}
			return From(rv.Elem().Interface())
		case reflect.String:
			return stringToValue(rv.String())
		case reflect.Slice, reflect.Array, reflect.Map:
			return boolToValue(rv.Len() > 0)
		case reflect.Struct:
			return True
		}
	}

	// Default to True - this is a non-zero value
	return True
}

func boolToValue(condition bool) Value {
	if condition {
		return True
	}
	return False
}

func stringToValue(s string) Value {
	x := strings.ToLower(s)
	switch x {
	case "true", "1", "t":
		return True
	case "false", "0", "f":
		return False
	case "unknown", "-1", "n", "nil", "null", "undefined":
		return Unknown
	default:
		return boolToValue(len(s) > 0)
	}
}
