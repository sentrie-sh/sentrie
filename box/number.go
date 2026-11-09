// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package box

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
)

// A number is boxed as one of two representations under the single kind
// ValueNumber (#166):
//
//   - an exact integer: the int64 bit pattern in u64, with exactIntTag in ref;
//   - a float: the float64 bit pattern in u64, with ref nil.
//
// Integer literals, integral JSON numbers that fit int64 and Go integer types
// box as exact integers; everything else boxes as a float. Kind() is
// ValueNumber for both, so code that only cares that a value is a number is
// unaffected. Code that needs the exact value reads IntValue.

// exactIntTag marks an exact integer. It is zero sized, so storing it in the
// ref interface does not allocate.
type exactIntTag struct{}

// ErrIntegerOverflow is returned when integer arithmetic leaves the int64
// range. Sentrie reports it rather than wrapping or silently rounding.
var ErrIntegerOverflow = errors.New("integer overflow")

const two63 = 9223372036854775808.0 // 2^63, the first float64 above math.MaxInt64

// Int boxes an exact integer.
func Int(x int64) Value {
	return Value{kind: ValueNumber, u64: uint64(x), ref: exactIntTag{}}
}

// Float boxes a float64. An integral float stays a float; it compares equal to
// the exact integer with the same value.
func Float(x float64) Value {
	return Value{kind: ValueNumber, u64: math.Float64bits(x)}
}

// Number boxes any Go numeric value. Signed integers and unsigned integers up
// to math.MaxInt64 box as exact integers; larger unsigned integers and floats
// box as floats.
func Number[T ~int | ~int8 | ~int16 | ~int32 | ~int64 |
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
	~float32 | ~float64](x T) Value {
	// The type set mixes integers and floats, so a type switch cannot see a
	// named type's underlying kind. Probe it by conversion instead: only a
	// float type keeps a half, and only an unsigned type wraps below zero.
	half := 0.5
	if T(half) != 0 {
		return Float(float64(x))
	}
	var below T
	below--
	if below > 0 {
		if u := uint64(x); u > math.MaxInt64 {
			return Float(float64(u))
		}
	}
	return Int(int64(x))
}

// IsInt reports whether v is an exact integer.
func (v Value) IsInt() bool {
	if v.kind != ValueNumber {
		return false
	}
	_, ok := v.ref.(exactIntTag)
	return ok
}

// IntValue returns the exact integer v holds. It reports false for a float,
// even an integral one, and for any other kind; see IntegralValue.
func (v Value) IntValue() (int64, bool) {
	if !v.IsInt() {
		return 0, false
	}
	return int64(v.u64), true
}

// IntegralValue returns v as an exact integer when it has one: an exact
// integer, or a float with no fractional part inside the int64 range.
func (v Value) IntegralValue() (int64, bool) {
	if i, ok := v.IntValue(); ok {
		return i, true
	}
	f, ok := v.NumberValue()
	if !ok || f != math.Trunc(f) || f < -two63 || f >= two63 {
		return 0, false
	}
	return int64(f), true
}

// ParseNumber parses a decimal number: an integer that fits int64 boxes as an
// exact integer, anything else strconv.ParseFloat accepts boxes as a float.
// A magnitude float64 cannot hold (1e400) is an error, never an infinity.
func ParseNumber(s string) (Value, error) {
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return Int(i), nil
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return Undefined(), err
	}
	return Float(f), nil
}

// CheckJSONNumbers walks a tree decoded with json.Decoder.UseNumber and
// reports the first number float64 cannot hold. The plain decoder rejects such
// a number while decoding; with UseNumber the check moves here, so every fact
// decoder must call it before the facts reach FromAny.
func CheckJSONNumbers(x any) error {
	switch t := x.(type) {
	case json.Number:
		if _, err := ParseNumber(string(t)); err != nil {
			return fmt.Errorf("number %s is out of range", t)
		}
	case []any:
		for _, item := range t {
			if err := CheckJSONNumbers(item); err != nil {
				return err
			}
		}
	case map[string]any:
		for _, item := range t {
			if err := CheckJSONNumbers(item); err != nil {
				return err
			}
		}
	}
	return nil
}

// fromJSONNumber boxes a number decoded with json.Decoder.UseNumber. Decoders
// reject an out-of-range number with CheckJSONNumbers first; one that still
// arrives here boxes as the float strconv returns (an infinity), as a Go
// float64 infinity passed to FromAny would.
func fromJSONNumber(n json.Number) Value {
	v, err := ParseNumber(string(n))
	if err != nil {
		f, _ := strconv.ParseFloat(string(n), 64)
		return Float(f)
	}
	return v
}

// numberOperands holds two numeric operands. ints is true when both are exact
// integers, in which case li and ri are set; otherwise lf and rf are.
type numberOperands struct {
	li, ri int64
	lf, rf float64
	ints   bool
}

func operands(lhs, rhs Value) (numberOperands, error) {
	if lhs.kind != ValueNumber {
		return numberOperands{}, fmt.Errorf("left operand is not a number")
	}
	if rhs.kind != ValueNumber {
		return numberOperands{}, fmt.Errorf("right operand is not a number")
	}
	li, lInt := lhs.IntValue()
	ri, rInt := rhs.IntValue()
	if lInt && rInt {
		return numberOperands{li: li, ri: ri, ints: true}, nil
	}
	lf, _ := lhs.NumberValue()
	rf, _ := rhs.NumberValue()
	return numberOperands{lf: lf, rf: rf}, nil
}

func overflow(a int64, op string, b int64) error {
	return fmt.Errorf("%w: %d %s %d", ErrIntegerOverflow, a, op, b)
}

// Add returns lhs + rhs. Two exact integers add exactly and overflow is an
// error; any float operand makes the result a float.
func Add(lhs, rhs Value) (Value, error) {
	o, err := operands(lhs, rhs)
	if err != nil {
		return Undefined(), err
	}
	if !o.ints {
		return Float(o.lf + o.rf), nil
	}
	s := o.li + o.ri
	if (o.ri > 0 && s < o.li) || (o.ri < 0 && s > o.li) {
		return Undefined(), overflow(o.li, "+", o.ri)
	}
	return Int(s), nil
}

// Sub returns lhs - rhs, with the same rules as Add.
func Sub(lhs, rhs Value) (Value, error) {
	o, err := operands(lhs, rhs)
	if err != nil {
		return Undefined(), err
	}
	if !o.ints {
		return Float(o.lf - o.rf), nil
	}
	d := o.li - o.ri
	if (o.ri > 0 && d > o.li) || (o.ri < 0 && d < o.li) {
		return Undefined(), overflow(o.li, "-", o.ri)
	}
	return Int(d), nil
}

// Mul returns lhs * rhs, with the same rules as Add.
func Mul(lhs, rhs Value) (Value, error) {
	o, err := operands(lhs, rhs)
	if err != nil {
		return Undefined(), err
	}
	if !o.ints {
		return Float(o.lf * o.rf), nil
	}
	if o.li == 0 || o.ri == 0 {
		return Int(0), nil
	}
	p := o.li * o.ri
	if p/o.ri != o.li || (o.li == -1 && o.ri == math.MinInt64) || (o.ri == -1 && o.li == math.MinInt64) {
		return Undefined(), overflow(o.li, "*", o.ri)
	}
	return Int(p), nil
}

// Div returns lhs / rhs. A zero divisor is an error. Two exact integers give an
// exact integer when the division is exact and a float otherwise, so 7 / 2 is
// 3.5 as it always was.
func Div(lhs, rhs Value) (Value, error) {
	o, err := operands(lhs, rhs)
	if err != nil {
		return Undefined(), err
	}
	if !o.ints {
		if o.rf == 0 {
			return Undefined(), fmt.Errorf("divide by zero")
		}
		return Float(o.lf / o.rf), nil
	}
	if o.ri == 0 {
		return Undefined(), fmt.Errorf("divide by zero")
	}
	if o.li == math.MinInt64 && o.ri == -1 {
		return Undefined(), overflow(o.li, "/", o.ri)
	}
	if o.li%o.ri == 0 {
		return Int(o.li / o.ri), nil
	}
	return Float(float64(o.li) / float64(o.ri)), nil
}

// Mod returns lhs % rhs, truncated like math.Mod (the result takes the sign of
// lhs). A zero divisor is an error. Two exact integers give an exact integer.
func Mod(lhs, rhs Value) (Value, error) {
	o, err := operands(lhs, rhs)
	if err != nil {
		return Undefined(), err
	}
	if !o.ints {
		if o.rf == 0 {
			return Undefined(), fmt.Errorf("divide by zero")
		}
		return Float(math.Mod(o.lf, o.rf)), nil
	}
	if o.ri == 0 {
		return Undefined(), fmt.Errorf("divide by zero")
	}
	// Go defines math.MinInt64 % -1 as 0, so there is no overflow case.
	return Int(o.li % o.ri), nil
}

// Negate returns -v. Negating math.MinInt64 overflows and is an error.
func Negate(v Value) (Value, error) {
	if i, ok := v.IntValue(); ok {
		if i == math.MinInt64 {
			return Undefined(), fmt.Errorf("%w: -(%d)", ErrIntegerOverflow, i)
		}
		return Int(-i), nil
	}
	f, ok := v.NumberValue()
	if !ok {
		return Undefined(), fmt.Errorf("operand is not a number")
	}
	return Float(-f), nil
}

// CompareNumbers orders two numbers exactly, across the integer and float
// representations: it returns -1, 0 or +1 as lhs is less than, equal to or
// greater than rhs. ordered is false when either side is NaN, which is
// unordered against everything. Either side not being a number is an error.
func CompareNumbers(lhs, rhs Value) (c int, ordered bool, err error) {
	o, err := operands(lhs, rhs)
	if err != nil {
		return 0, false, err
	}
	if o.ints {
		return cmp.Compare(o.li, o.ri), true, nil
	}
	li, lInt := lhs.IntValue()
	ri, rInt := rhs.IntValue()
	switch {
	case lInt:
		if math.IsNaN(o.rf) {
			return 0, false, nil
		}
		return compareIntFloat(li, o.rf), true, nil
	case rInt:
		if math.IsNaN(o.lf) {
			return 0, false, nil
		}
		return -compareIntFloat(ri, o.lf), true, nil
	default:
		if math.IsNaN(o.lf) || math.IsNaN(o.rf) {
			return 0, false, nil
		}
		return cmp.Compare(o.lf, o.rf), true, nil
	}
}

// compareIntFloat orders an int64 against a non-NaN float64 without rounding
// the integer through float64, which would merge neighbours above 2^53.
func compareIntFloat(i int64, f float64) int {
	if f >= two63 {
		return -1
	}
	if f < -two63 {
		return 1
	}
	t := math.Trunc(f)
	if c := cmp.Compare(i, int64(t)); c != 0 {
		return c
	}
	// The integer parts agree; any fractional part decides.
	return cmp.Compare(t, f)
}
