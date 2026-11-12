// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/tokens"
	"github.com/sentrie-sh/sentrie/trinary"
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
		if typeRef.Boolean {
			return errMismatch("value '%v' is not a boolean at %s - expected true or false", v, valueRange)
		}
		return errMismatch("value '%v' is not a trinary at %s - expected true, false or unknown", v, valueRange)
	}
	// A boolean is a trinary that rejects unknown (#144).
	if typeRef.Boolean && tv == trinary.Unknown {
		return errMismatch("value 'unknown' is not a boolean at %s - expected true or false; use trinary to allow unknown", valueRange)
	}

	return checkConstraints(ctx, box.Trinary(tv), bound, valueRange)
}

// isBooleanTypeRef reports whether typeRef is `boolean` or `boolean?`, the
// truth type that rejects unknown, rather than `trinary`.
func isBooleanTypeRef(typeRef ast.TypeRef) bool {
	tr, ok := ast.UnwrapNullableTypeRef(typeRef).(*ast.TrinaryTypeRef)
	return ok && tr.Boolean
}

// bindTruthValue returns the value a binding declared as typeRef holds. An
// imported decision envelope bound to a boolean or trinary type becomes its state,
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
