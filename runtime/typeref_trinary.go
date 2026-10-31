// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/tokens"
)

func validateAgainstTrinaryTypeRef(ctx context.Context, ec *ExecutionContext, exec Executor, p *index.Policy, v box.Value, typeRef *ast.TrinaryTypeRef, valueRange tokens.Range) error {
	bound, err := bindConstraints(ctx, ec, exec, p, typeRef)
	if err != nil {
		return err
	}

	// A known truth value is a ValueBool and unknown is a ValueTrinary; TrinaryValue accepts both.
	// An imported decision checked as a truth value is its state (#136).
	tv, ok := v.TrinaryValue()
	if !ok {
		tv, ok = v.DecisionState()
	}
	if !ok {
		return errMismatch("value '%v' is not a bool at %s - expected bool", v, valueRange)
	}

	return checkConstraints(ctx, box.Trinary(tv), bound, valueRange)
}

// bindTruthValue returns the value a binding declared as typeRef holds. An
// imported decision envelope bound to a bool or trinary type becomes its state,
// so the binding holds a truth value rather than a dict that only reads as one
// in truth positions (#136). It is boxed with box.Trinary, like a truth
// literal, so a known state is a ValueBool that compares equal to true and
// false and crosses into JS as a boolean, and unknown stays a ValueTrinary.
// Only the binding itself is rebound: an envelope inside a list, dict or shape
// field stays an envelope. Every other value is unchanged.
func bindTruthValue(v box.Value, typeRef ast.TypeRef) box.Value {
	if typeRef == nil {
		return v
	}
	if ast.IsNullableTypeRef(typeRef) {
		typeRef = ast.UnwrapNullableTypeRef(typeRef)
	}
	if _, ok := typeRef.(*ast.TrinaryTypeRef); !ok {
		return v
	}
	if state, ok := v.DecisionState(); ok {
		return box.Trinary(state)
	}
	return v
}
