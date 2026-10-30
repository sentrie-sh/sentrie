// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/trinary"
)

type Decision struct {
	State trinary.Value `json:"state"`
	Value box.Value     `json:"value"`
}

func (d Decision) ToTrinary() trinary.Value {
	return d.State
}

type DecisionAttachments map[string]box.Value

// Behaviour:
// - undefined or null      → Unknown, keeping the original value
// - boolean or unknown     → Decision with the same truth value
// - anything else          → box.TrinaryFrom(val)
func DecisionOf(val box.Value) *Decision {
	if val.IsUndefined() || val.IsNull() {
		return &Decision{State: trinary.Unknown, Value: val}
	}

	if d, ok := val.TrinaryValue(); ok {
		return &Decision{State: d, Value: val}
	}

	return &Decision{State: box.TrinaryFrom(val), Value: val}
}
