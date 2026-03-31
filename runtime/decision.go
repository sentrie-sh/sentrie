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

package runtime

import (
	"encoding/json"

	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/trinary"
)

type Decision struct {
	State trinary.Value `json:"state"`
	Value box.Value     `json:"-"`
}

func (d Decision) ToTrinary() trinary.Value {
	return d.State
}

type DecisionAttachments map[string]box.Value

// Behaviour:
// - nil           → Unknown
// - *Decision     → as-is
// - trinary.Value  → Decision with the same state
// - anything else → trinary.From(val)
func DecisionOf(val box.Value) *Decision {
	if val.IsUndefined() || val.IsNull() {
		return &Decision{State: trinary.Unknown, Value: val}
	}

	if d, ok := val.TrinaryValue(); ok {
		return &Decision{State: d, Value: val}
	}

	return &Decision{State: box.TrinaryFrom(val), Value: val}
}

func (d Decision) MarshalJSON() ([]byte, error) {
	type dto struct {
		State trinary.Value `json:"state"`
		Value box.Value     `json:"value"`
	}
	return json.Marshal(dto{
		State: d.State,
		Value: d.Value,
	})
}
