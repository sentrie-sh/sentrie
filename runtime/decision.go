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

import "github.com/binaek/sentra/trinary"

type Decision struct {
	State trinary.Value `json:"state"`
	Value any           `json:"value"`
}

func (d Decision) ToTrinary() trinary.Value {
	return d.State
}

type DecisionAttachments map[string]any

// Behaviour:
// - nil           → Unknown
// - *Decision     → as-is
// - trinary.Value  → Decision with the same state
// - anything else → trinary.From(val)
func DecisionOf(val any) *Decision {
	if val == nil {
		return &Decision{State: trinary.Unknown, Value: nil}
	}

	if d, ok := val.(*Decision); ok {
		return d
	}

	if d, ok := val.(Decision); ok {
		return &Decision{State: d.State, Value: d.Value}
	}

	if d, ok := val.(trinary.HasTrinary); ok {
		return &Decision{State: d.ToTrinary(), Value: val}
	}

	return &Decision{State: trinary.From(val), Value: val}
}
