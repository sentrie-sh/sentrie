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
// - nil           → Unknown
// - *Decision     → as-is
// - trinary.Value  → Decision with the same state
// - import envelope → its imported state (via box.TrinaryFrom)
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

// reported returns d as ExecRule reports it to its caller. Inside a policy a rule
// whose body is an import keeps the envelope as its value, so references to the
// rule can still read .state, .value and attachments. Outside, the envelope is
// unwrapped to the imported value, so the reported value agrees with the state
// the decision already carries from the envelope (#136).
func (d *Decision) reported() *Decision {
	if d == nil {
		return nil
	}
	if _, ok := d.Value.DecisionState(); !ok {
		return d
	}
	m, _ := d.Value.DictValue()
	value, ok := m["value"]
	if !ok {
		value = box.Undefined()
	}
	return &Decision{State: d.State, Value: value}
}
