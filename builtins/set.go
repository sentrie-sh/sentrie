// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package builtins

import (
	"context"
	"fmt"

	"github.com/sentrie-sh/sentrie/box"
)

// Set builtins (#162). A set is a list with distinct elements in canonical
// order (see box.NewSet). The set operations take a set or a plain list for
// each operand, so facts that arrive as JSON arrays can be compared without
// first calling set(); the result is always a set. All of them are strict:
// an undefined or null operand is an error, as for merge and distinct.

var (
	declSet = &Decl{
		Name:        "set",
		Description: "Builds a set from a list: distinct elements in canonical sorted order.",
		DeriveSafe:  true,
		Sig: Sig{
			Params:       []ParamSig{{Name: "elements", Kinds: kindList, KindError: "set: argument must be a list"}},
			TooFewError:  "set requires 1 argument",
			TooManyError: "set requires 1 argument",
			Result:       ParamSig{Name: "result", Kinds: kindList},
		},
		Impl: implSet,
	}

	declUnion        = setOpDecl("union", "Returns the set of elements in either operand.", kindList, box.SetUnion)
	declIntersection = setOpDecl("intersection", "Returns the set of elements in both operands.", kindList, box.SetIntersection)
	declDifference   = setOpDecl("difference", "Returns the set of elements in the first operand that are not in the second.", kindList, box.SetDifference)
	declIsSubset     = setOpDecl("is_subset", "Reports whether every element of the first operand is in the second.", kindBool, setPredicate(false))
	declIsSuperset   = setOpDecl("is_superset", "Reports whether every element of the second operand is in the first.", kindBool, setPredicate(true))
)

// setOpDecl declares a two-operand set builtin; each operand is a set or list.
func setOpDecl(name, description string, result []box.ValueKind, op func(a, b box.Value) (box.Value, error)) *Decl {
	first := name + ": first argument must be a set or list"
	second := name + ": second argument must be a set or list"
	return &Decl{
		Name:        name,
		Description: description,
		DeriveSafe:  true,
		Sig: Sig{
			Params: []ParamSig{
				{Name: "first", Kinds: kindList, KindError: first},
				{Name: "second", Kinds: kindList, KindError: second},
			},
			TooFewError:  name + " requires 2 arguments",
			TooManyError: name + " requires 2 arguments",
			Result:       ParamSig{Name: "result", Kinds: result},
		},
		Impl: func(ctx context.Context, env Env, args ...box.Value) (box.Value, error) {
			// Precheck lets undefined and null through; these builtins reject them.
			if args[0].Kind() != box.ValueList {
				return box.Undefined(), fmt.Errorf("%s", first)
			}
			if args[1].Kind() != box.ValueList {
				return box.Undefined(), fmt.Errorf("%s", second)
			}
			// Convert here, not in op, so an element error names its argument.
			a, err := box.ToSet(args[0])
			if err != nil {
				return box.Undefined(), fmt.Errorf("%s: first argument: %w", name, err)
			}
			b, err := box.ToSet(args[1])
			if err != nil {
				return box.Undefined(), fmt.Errorf("%s: second argument: %w", name, err)
			}
			return op(a, b)
		},
	}
}

// setPredicate is is_subset, or is_superset (the operands swapped) when swap is set.
func setPredicate(swap bool) func(a, b box.Value) (box.Value, error) {
	return func(a, b box.Value) (box.Value, error) {
		if swap {
			a, b = b, a
		}
		ok, err := box.SetIsSubset(a, b)
		if err != nil {
			return box.Undefined(), err
		}
		return box.Bool(ok), nil
	}
}

func implSet(ctx context.Context, env Env, args ...box.Value) (box.Value, error) {
	xs, ok := args[0].ListValue()
	if !ok {
		return box.Undefined(), fmt.Errorf("set: argument must be a list")
	}
	v, err := box.NewSet(xs)
	if err != nil {
		return box.Undefined(), fmt.Errorf("set: %w", err)
	}
	return v, nil
}
