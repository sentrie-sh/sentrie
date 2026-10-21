// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package js

import (
	"fmt"
	"math"
	"strconv"

	"github.com/dop251/goja"
)

// BuiltinMathGo provides @sentrie/math: number helpers for policy code.
//
// The module never lets a non-finite number in or out, so every value it
// handles stays ordered, comparable and JSON-encodable in a decision:
//   - every argument must be a finite JS number; NaN, Infinity, and any
//     non-number (including numeric strings) throw;
//   - an operation whose result would be NaN or Infinity (sqrt of a negative,
//     an overflowing pow) throws;
//   - a negative-zero result is returned as 0.
//
// Argument counts are exact, and every failure is a thrown JS exception; no
// function returns an Error object. `round` rounds half away from zero, which
// differs from JavaScript's Math.round for negative halves (-2.5 gives -3).
// The IEEE-754 behaviour of JavaScript's Math stays available through
// @sentrie/js.
var BuiltinMathGo = func(vm *goja.Runtime) (*goja.Object, error) {
	ex := vm.NewObject()

	unary := func(name string, fn func(float64) float64) {
		_ = ex.Set(name, func(call goja.FunctionCall) goja.Value {
			mathArgCount(vm, call, name, 1)
			return mathResult(vm, name, fn(mathArg(vm, call, name, 0)))
		})
	}

	unary("abs", math.Abs)
	unary("floor", math.Floor)
	unary("ceil", math.Ceil)
	unary("round", math.Round)
	unary("trunc", math.Trunc)
	unary("sqrt", func(x float64) float64 {
		if x < 0 {
			mathThrow(vm, "sqrt: argument must be non-negative, got %s", mathFormat(x))
		}
		return math.Sqrt(x)
	})

	_ = ex.Set("pow", func(call goja.FunctionCall) goja.Value {
		mathArgCount(vm, call, "pow", 2)
		return mathResult(vm, "pow", math.Pow(mathArg(vm, call, "pow", 0), mathArg(vm, call, "pow", 1)))
	})

	_ = ex.Set("min", func(call goja.FunctionCall) goja.Value {
		values := mathValues(vm, call, "min")
		out := values[0]
		for _, v := range values[1:] {
			out = math.Min(out, v)
		}
		return mathResult(vm, "min", out)
	})

	_ = ex.Set("max", func(call goja.FunctionCall) goja.Value {
		values := mathValues(vm, call, "max")
		out := values[0]
		for _, v := range values[1:] {
			out = math.Max(out, v)
		}
		return mathResult(vm, "max", out)
	})

	_ = ex.Set("clamp", func(call goja.FunctionCall) goja.Value {
		mathArgCount(vm, call, "clamp", 3)
		x := mathArg(vm, call, "clamp", 0)
		lo := mathArg(vm, call, "clamp", 1)
		hi := mathArg(vm, call, "clamp", 2)
		if lo > hi {
			mathThrow(vm, "clamp: lower bound %s is greater than upper bound %s", mathFormat(lo), mathFormat(hi))
		}
		return mathResult(vm, "clamp", math.Min(math.Max(x, lo), hi))
	})

	return ex, nil
}

// mathThrow throws a JS exception carrying a Go error.
func mathThrow(vm *goja.Runtime, format string, args ...any) {
	panic(vm.NewGoError(fmt.Errorf(format, args...)))
}

func mathArgCount(vm *goja.Runtime, call goja.FunctionCall, name string, want int) {
	if got := len(call.Arguments); got != want {
		mathThrow(vm, "%s: expected %d argument(s), got %d", name, want, got)
	}
}

func mathArg(vm *goja.Runtime, call goja.FunctionCall, name string, i int) float64 {
	return mathNumber(vm, call.Argument(i), "%s: argument %d", name, i+1)
}

// mathNumber returns v as a finite float64, or throws naming the value with
// the given description.
func mathNumber(vm *goja.Runtime, v goja.Value, what string, whatArgs ...any) float64 {
	if !goja.IsNumber(v) {
		mathThrow(vm, "%s must be a number, got %s", fmt.Sprintf(what, whatArgs...), stringTypeName(v))
	}
	f := v.ToFloat()
	if math.IsNaN(f) || math.IsInf(f, 0) {
		mathThrow(vm, "%s must be a finite number, got %s", fmt.Sprintf(what, whatArgs...), v.String())
	}
	return f
}

// mathValues collects the operands of min and max: either one or more number
// arguments, or a single non-empty array of numbers.
func mathValues(vm *goja.Runtime, call goja.FunctionCall, name string) []float64 {
	args := call.Arguments
	if len(args) == 0 {
		mathThrow(vm, "%s: expected at least 1 argument, got 0", name)
	}
	if obj, ok := args[0].(*goja.Object); ok && obj.ClassName() == "Array" {
		if len(args) != 1 {
			mathThrow(vm, "%s: an array must be the only argument, got %d arguments", name, len(args))
		}
		n := int(obj.Get("length").ToInteger())
		if n == 0 {
			mathThrow(vm, "%s: the array must not be empty", name)
		}
		out := make([]float64, n)
		for i := range n {
			out[i] = mathNumber(vm, obj.Get(strconv.Itoa(i)), "%s: array element %d", name, i)
		}
		return out
	}
	out := make([]float64, len(args))
	for i, v := range args {
		out[i] = mathNumber(vm, v, "%s: argument %d", name, i+1)
	}
	return out
}

// mathResult rejects non-finite results and normalizes negative zero.
func mathResult(vm *goja.Runtime, name string, f float64) goja.Value {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		mathThrow(vm, "%s: result is not a finite number", name)
	}
	if f == 0 {
		f = 0 // drop the sign of negative zero
	}
	return vm.ToValue(f)
}

func mathFormat(f float64) string {
	return strconv.FormatFloat(f, 'g', -1, 64)
}
