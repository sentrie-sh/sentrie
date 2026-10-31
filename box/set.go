// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package box

import (
	"cmp"
	"fmt"
	"math"
	"slices"
	"strings"
)

// setTag marks a list built by NewSet or NewSetUnique in the otherwise unused
// u64 slot. List never sets u64, so no list built from policy data, JSON or a
// module can carry the tag.
//
// A set is a list, not a kind of its own (#162): its elements are distinct
// and held in canonical order (see compareSetElements), so every path that
// reads a list (`in`, `contains`, `count`, indexing, iteration, JSON, the JS
// boundary) reads a set unchanged, and serialises it as a sorted array. The
// tag lives on this Value only: a list rebuilt from it (by `map`, `filter` or
// a module) is a plain list again.
const setTag uint64 = 1 << 9

// IsSet reports whether v is a set built by NewSet or NewSetUnique.
func (v Value) IsSet() bool {
	return v.kind == ValueList && v.u64&setTag != 0
}

// NewSet builds a set from xs: the elements are copied, sorted into canonical
// order and deduplicated. It fails when an element cannot be held in a set
// (see checkSetElement).
func NewSet(xs []Value) (Value, error) {
	return canonicalSet(xs, true)
}

// NewSetUnique builds a set from xs like NewSet, but fails on a duplicate
// element instead of dropping it. It converts a list that must already hold
// distinct elements, such as a fact declared as set[T].
func NewSetUnique(xs []Value) (Value, error) {
	return canonicalSet(xs, false)
}

// ToSet returns v when it is already a set, and otherwise builds a set from a
// list with NewSet. Any other kind is an error.
func ToSet(v Value) (Value, error) {
	if v.IsSet() {
		return v, nil
	}
	xs, ok := v.ListValue()
	if !ok {
		return Value{}, fmt.Errorf("value of kind %s is not a set or list", v.kind)
	}
	return NewSet(xs)
}

func canonicalSet(xs []Value, dedupe bool) (Value, error) {
	out := make([]Value, 0, len(xs))
	for _, x := range xs {
		if err := checkSetElement(x); err != nil {
			return Value{}, err
		}
		x, _ = canonicalElement(x)
		out = append(out, x)
	}
	slices.SortStableFunc(out, compareSetElements)
	w := 0
	for i := range out {
		if w > 0 && compareSetElements(out[w-1], out[i]) == 0 {
			if !dedupe {
				return Value{}, fmt.Errorf("duplicate set element %s", out[i])
			}
			continue
		}
		out[w] = out[i]
		w++
	}
	return Value{kind: ValueList, u64: setTag, ref: out[:w]}, nil
}

// canonicalElement returns x as a set holds it, and whether that differs from
// x. At every depth, an imported decision is the plain dict it reads as, so
// EqualValues on the held elements never compares a decision's state with a
// bool (an untagged `[impT]` would equal `[true]` and break dedupe), and -0 is
// 0, so the JSON a set produces does not depend on which of the two equal
// zeros came first. A nested list keeps its set tag.
func canonicalElement(x Value) (Value, bool) {
	switch x.kind {
	case ValueNumber:
		if n, _ := x.NumberValue(); n == 0 && math.Signbit(n) {
			return Number(0), true
		}
	case ValueList:
		xs, _ := x.ListValue()
		var out []Value
		for i, e := range xs {
			c, changed := canonicalElement(e)
			if changed && out == nil {
				out = make([]Value, i, len(xs))
				copy(out, xs[:i])
			}
			if out != nil {
				out = append(out, c)
			}
		}
		if out != nil {
			return Value{kind: ValueList, u64: x.u64 & setTag, ref: out}, true
		}
	case ValueDict:
		m, _ := x.DictValue()
		_, changed := x.DecisionState()
		var repl map[string]Value
		for k, e := range m {
			if c, ok := canonicalElement(e); ok {
				if repl == nil {
					repl = map[string]Value{}
				}
				repl[k] = c
			}
		}
		if changed || repl != nil {
			out := make(map[string]Value, len(m))
			for k, e := range m {
				if c, ok := repl[k]; ok {
					e = c
				}
				out[k] = e
			}
			return Dict(out), true
		}
	}
	return x, false
}

// checkSetElement reports whether v can be held in a set: null, a bool, a
// number other than NaN, a string, or a list or dict whose members can be
// held in a set. These are exactly the values on which compareSetElements is
// a total order that agrees with EqualValues, so membership and duplicates
// are well defined. Undefined, unknown, NaN, documents and callables have no
// such order and are rejected.
func checkSetElement(v Value) error {
	switch v.kind {
	case ValueNull, ValueBool, ValueString:
		return nil
	case ValueNumber:
		if n, _ := v.NumberValue(); math.IsNaN(n) {
			return fmt.Errorf("a set cannot hold NaN")
		}
		return nil
	case ValueList:
		xs, _ := v.ListValue()
		for _, x := range xs {
			if err := checkSetElement(x); err != nil {
				return err
			}
		}
		return nil
	case ValueDict:
		m, _ := v.DictValue()
		for _, x := range m {
			if err := checkSetElement(x); err != nil {
				return err
			}
		}
		return nil
	case ValueTrinary:
		return fmt.Errorf("a set cannot hold an unknown value")
	default:
		return fmt.Errorf("a set cannot hold a value of kind %s", v.kind)
	}
}

func setKindRank(k ValueKind) int {
	switch k {
	case ValueNull:
		return 0
	case ValueBool:
		return 1
	case ValueNumber:
		return 2
	case ValueString:
		return 3
	case ValueList:
		return 4
	default: // ValueDict; checkSetElement admits nothing else
		return 5
	}
}

// compareSetElements is the canonical order of set elements: by kind (null,
// bool, number, string, list, dict), then by value. Lists compare element by
// element, then by length; dicts compare their entries in key order, key
// first, then by size. For values checkSetElement admits it returns 0 exactly
// when EqualValues holds.
func compareSetElements(a, b Value) int {
	if c := cmp.Compare(setKindRank(a.kind), setKindRank(b.kind)); c != 0 {
		return c
	}
	switch a.kind {
	case ValueBool:
		av, _ := a.BoolValue()
		bv, _ := b.BoolValue()
		switch {
		case av == bv:
			return 0
		case !av:
			return -1
		default:
			return 1
		}
	case ValueNumber:
		av, _ := a.NumberValue()
		bv, _ := b.NumberValue()
		return cmp.Compare(av, bv)
	case ValueString:
		av, _ := a.StringValue()
		bv, _ := b.StringValue()
		return strings.Compare(av, bv)
	case ValueList:
		al, _ := a.ListValue()
		bl, _ := b.ListValue()
		return slices.CompareFunc(al, bl, compareSetElements)
	case ValueDict:
		am, _ := a.DictValue()
		bm, _ := b.DictValue()
		ak := sortedKeys(am)
		bk := sortedKeys(bm)
		for i := 0; i < len(ak) && i < len(bk); i++ {
			if c := strings.Compare(ak[i], bk[i]); c != 0 {
				return c
			}
			if c := compareSetElements(am[ak[i]], bm[bk[i]]); c != 0 {
				return c
			}
		}
		return cmp.Compare(len(ak), len(bk))
	default: // ValueNull
		return 0
	}
}

func sortedKeys(m map[string]Value) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

// setContains reports whether the set elements xs hold needle. It is only
// used for scalar needles, whose EqualValues matches are exactly the elements
// compareSetElements finds equal.
func setContains(xs []Value, needle Value) bool {
	_, found := slices.BinarySearchFunc(xs, needle, compareSetElements)
	return found
}

func isScalarSetNeedle(v Value) bool {
	switch v.kind {
	case ValueNull, ValueBool, ValueString:
		return true
	case ValueNumber:
		n, _ := v.NumberValue()
		return !math.IsNaN(n)
	default:
		return false
	}
}

// setOperands returns the elements of two sets. The operations merge over the
// canonical order, so a plain list is refused rather than converted: the
// caller converts with ToSet, which is where an element error can name its
// argument.
func setOperands(a, b Value) ([]Value, []Value, error) {
	if !a.IsSet() {
		return nil, nil, fmt.Errorf("first operand is not a set")
	}
	if !b.IsSet() {
		return nil, nil, fmt.Errorf("second operand is not a set")
	}
	ax, _ := a.ListValue()
	bx, _ := b.ListValue()
	return ax, bx, nil
}

func setOf(xs []Value) Value {
	return Value{kind: ValueList, u64: setTag, ref: xs}
}

// SetUnion returns the set of elements in a or b. Both operands must be sets.
func SetUnion(a, b Value) (Value, error) {
	ax, bx, err := setOperands(a, b)
	if err != nil {
		return Value{}, err
	}
	out := make([]Value, 0, len(ax)+len(bx))
	i, j := 0, 0
	for i < len(ax) && j < len(bx) {
		switch c := compareSetElements(ax[i], bx[j]); {
		case c < 0:
			out = append(out, ax[i])
			i++
		case c > 0:
			out = append(out, bx[j])
			j++
		default:
			out = append(out, ax[i])
			i++
			j++
		}
	}
	out = append(out, ax[i:]...)
	out = append(out, bx[j:]...)
	return setOf(out), nil
}

// SetIntersection returns the set of elements in both a and b. Both operands
// must be sets.
func SetIntersection(a, b Value) (Value, error) {
	ax, bx, err := setOperands(a, b)
	if err != nil {
		return Value{}, err
	}
	out := make([]Value, 0, min(len(ax), len(bx)))
	i, j := 0, 0
	for i < len(ax) && j < len(bx) {
		switch c := compareSetElements(ax[i], bx[j]); {
		case c < 0:
			i++
		case c > 0:
			j++
		default:
			out = append(out, ax[i])
			i++
			j++
		}
	}
	return setOf(out), nil
}

// SetDifference returns the set of elements in a that are not in b. Both
// operands must be sets.
func SetDifference(a, b Value) (Value, error) {
	ax, bx, err := setOperands(a, b)
	if err != nil {
		return Value{}, err
	}
	out := make([]Value, 0, len(ax))
	i, j := 0, 0
	for i < len(ax) {
		if j >= len(bx) {
			out = append(out, ax[i:]...)
			break
		}
		switch c := compareSetElements(ax[i], bx[j]); {
		case c < 0:
			out = append(out, ax[i])
			i++
		case c > 0:
			j++
		default:
			i++
			j++
		}
	}
	return setOf(out), nil
}

// SetIsSubset reports whether every element of a is in b. Both operands must
// be sets.
func SetIsSubset(a, b Value) (bool, error) {
	ax, bx, err := setOperands(a, b)
	if err != nil {
		return false, err
	}
	if len(ax) > len(bx) {
		return false, nil
	}
	j := 0
	for _, x := range ax {
		for j < len(bx) && compareSetElements(bx[j], x) < 0 {
			j++
		}
		if j >= len(bx) || compareSetElements(bx[j], x) != 0 {
			return false, nil
		}
		j++
	}
	return true, nil
}
