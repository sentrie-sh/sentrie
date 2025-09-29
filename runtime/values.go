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

package runtime

import "github.com/sentrie-sh/sentrie/trinary"

// Runtime values are plain Go values:
//  - bool, int64, float64, string
//  - []any (lists)
//  - map[string]any (maps/objects)
// Attachments & imports use map[string]any too.

// Helpers:
func AsBool(v any) bool { b, _ := v.(bool); return b }
func AsInt(v any) int64 {
	switch t := v.(type) {
	case int:
		return int64(t)
	case int64:
		return t
	case float64:
		return int64(t)
	default:
		return 0
	}
}

func AsFloat(v any) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case int:
		return float64(t)
	case int64:
		return float64(t)
	default:
		return 0
	}
}
func AsString(v any) string { s, _ := v.(string); return s }

func IsTruthy(v any) bool {
	return trinary.IsTruthy(v)
}

func IsUndefined(v any) bool {
	return v == Undefined
}

type undefined struct{}

var Undefined = &undefined{}

func (u *undefined) String() string {
	return "undefined"
}

func (u *undefined) Value() any {
	return nil
}
