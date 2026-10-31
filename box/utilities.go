// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package box

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/sentrie-sh/sentrie/trinary"
)

// MustNumbers returns both operands as float64 numbers, or an error if either is not numeric.
func MustNumbers(lhs, rhs Value) (float64, float64, error) {
	l, ok := lhs.NumberValue()
	if !ok {
		return 0, 0, fmt.Errorf("left operand is not a number")
	}
	r, ok := rhs.NumberValue()
	if !ok {
		return 0, 0, fmt.Errorf("right operand is not a number")
	}
	return l, r, nil
}

// EqualValues compares two boxed values for semantic equality (including cross-kind number equality).
// A DecisionEnvelope compared with a bool or trinary compares its state, so an imported decision
// equals true only when its state is true; an unknown state equals neither true nor false.
func EqualValues(a, b Value) bool {
	if eq, ok := envelopeStateEquals(a, b); ok {
		return eq
	}
	if eq, ok := envelopeStateEquals(b, a); ok {
		return eq
	}
	if a.Kind() != b.Kind() {
		an, aok := a.NumberValue()
		bn, bok := b.NumberValue()
		return aok && bok && an == bn
	}

	switch a.Kind() {
	case ValueUndefined, ValueNull:
		return true
	case ValueBool:
		av, _ := a.BoolValue()
		bv, _ := b.BoolValue()
		return av == bv
	case ValueNumber:
		av, _ := a.NumberValue()
		bv, _ := b.NumberValue()
		return av == bv
	case ValueString:
		av, _ := a.StringValue()
		bv, _ := b.StringValue()
		return av == bv
	case ValueTrinary:
		av, _ := a.TrinaryValue()
		bv, _ := b.TrinaryValue()
		return av == bv
	case ValueList:
		al, _ := a.ListValue()
		bl, _ := b.ListValue()
		if len(al) != len(bl) {
			return false
		}
		for i := range al {
			if !EqualValues(al[i], bl[i]) {
				return false
			}
		}
		return true
	case ValueDict:
		am, _ := a.DictValue()
		bm, _ := b.DictValue()
		if len(am) != len(bm) {
			return false
		}
		for k, av := range am {
			bv, ok := bm[k]
			if !ok || !EqualValues(av, bv) {
				return false
			}
		}
		return true
	case ValueDocument:
		return a.SameDocumentRef(b)
	default:
		return false
	}
}

// envelopeStateEquals compares env's decision state with a bool or trinary operand. It
// reports ok false when env is not a DecisionEnvelope or other is not a truth literal.
func envelopeStateEquals(env, other Value) (eq bool, ok bool) {
	state, ok := env.DecisionState()
	if !ok {
		return false, false
	}
	if b, ok := other.BoolValue(); ok {
		return state == trinary.From(b), true
	}
	if t, ok := other.TrinaryValue(); ok {
		return state == t, true
	}
	return false, false
}

// MatchesValue returns whether haystack matches the regexp pattern; both must be strings.
func MatchesValue(haystack, pattern Value) (bool, error) {
	h, ok := haystack.StringValue()
	if !ok {
		return false, fmt.Errorf("haystack must be a string")
	}
	p, ok := pattern.StringValue()
	if !ok {
		return false, fmt.Errorf("pattern must be a string")
	}
	return regexp.MatchString(p, h)
}

// ContainsValue implements infix `contains` / `in` semantics for string, list, and dict haystacks.
// For dict haystacks, only string-key lookup and dict-subset containment are supported. A set
// haystack (a list built by NewSet) is searched by binary search for a scalar needle.
func ContainsValue(haystack, needle Value) bool {
	switch haystack.Kind() {
	case ValueString:
		h, _ := haystack.StringValue()
		n, ok := needle.StringValue()
		return ok && n != "" && strings.Contains(h, n)
	case ValueList:
		xs, _ := haystack.ListValue()
		if haystack.IsSet() && isScalarSetNeedle(needle) {
			return setContains(xs, needle)
		}
		for _, v := range xs {
			if EqualValues(v, needle) {
				return true
			}
		}
		return false
	case ValueDict:
		m, _ := haystack.DictValue()
		if s, ok := needle.StringValue(); ok {
			_, ok2 := m[s]
			return ok2
		}
		if sub, ok := needle.DictValue(); ok {
			for k, v := range sub {
				mv, ok2 := m[k]
				if !ok2 {
					return false
				}
				if !EqualValues(v, mv) {
					return false
				}
			}
			return true
		}
		return false
	default:
		return false
	}
}
