// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2026 Binaek Sarkar
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

package box

import (
	"fmt"
	"regexp"
	"strings"
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
func EqualValues(a, b Value) bool {
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
	case ValueMap:
		am, _ := a.MapValue()
		bm, _ := b.MapValue()
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

// ContainsValue implements infix `contains` / `in` semantics for string, list, and map haystacks.
func ContainsValue(haystack, needle Value) bool {
	switch haystack.Kind() {
	case ValueString:
		h, _ := haystack.StringValue()
		n, ok := needle.StringValue()
		return ok && n != "" && strings.Contains(h, n)
	case ValueList:
		xs, _ := haystack.ListValue()
		for _, v := range xs {
			if EqualValues(v, needle) {
				return true
			}
		}
		return false
	case ValueMap:
		m, _ := haystack.MapValue()
		if s, ok := needle.StringValue(); ok {
			_, ok2 := m[s]
			return ok2
		}
		if sub, ok := needle.MapValue(); ok {
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
		for _, v := range m {
			if EqualValues(v, needle) {
				return true
			}
		}
		return false
	default:
		return false
	}
}
