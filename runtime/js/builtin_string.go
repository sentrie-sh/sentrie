// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package js

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/dop251/goja"
)

// BuiltinStringGo provides @sentrie/string: Unicode-aware string helpers for
// policy code.
//
// Every function checks its argument count and types exactly and throws a JS
// exception on a mismatch; none returns an Error object. Strings cross into Go
// as UTF-8, so a lone UTF-16 surrogate in a JS string becomes U+FFFD. Case
// mapping and folding use Go's simple (one-to-one) Unicode mappings, and the
// trim functions strip Unicode White_Space.
var BuiltinStringGo = func(vm *goja.Runtime) (*goja.Object, error) {
	ex := vm.NewObject()

	unary := func(name string, fn func(string) goja.Value) {
		_ = ex.Set(name, func(call goja.FunctionCall) goja.Value {
			stringArgCount(vm, call, name, 1)
			return fn(stringArg(vm, call, name, 0))
		})
	}
	binary := func(name string, fn func(a, b string) goja.Value) {
		_ = ex.Set(name, func(call goja.FunctionCall) goja.Value {
			stringArgCount(vm, call, name, 2)
			return fn(stringArg(vm, call, name, 0), stringArg(vm, call, name, 1))
		})
	}
	ternary := func(name string, fn func(a, b, c string) goja.Value) {
		_ = ex.Set(name, func(call goja.FunctionCall) goja.Value {
			stringArgCount(vm, call, name, 3)
			return fn(stringArg(vm, call, name, 0), stringArg(vm, call, name, 1), stringArg(vm, call, name, 2))
		})
	}
	str := func(s string) goja.Value { return vm.ToValue(s) }
	boolean := func(b bool) goja.Value { return vm.ToValue(b) }

	unary("lower", func(s string) goja.Value { return str(strings.ToLower(s)) })
	unary("upper", func(s string) goja.Value { return str(strings.ToUpper(s)) })
	unary("trim", func(s string) goja.Value { return str(strings.TrimFunc(s, unicode.IsSpace)) })
	unary("trimStart", func(s string) goja.Value { return str(strings.TrimLeftFunc(s, unicode.IsSpace)) })
	unary("trimEnd", func(s string) goja.Value { return str(strings.TrimRightFunc(s, unicode.IsSpace)) })

	binary("startsWith", func(s, prefix string) goja.Value { return boolean(strings.HasPrefix(s, prefix)) })
	binary("endsWith", func(s, suffix string) goja.Value { return boolean(strings.HasSuffix(s, suffix)) })
	binary("includes", func(s, sub string) goja.Value { return boolean(strings.Contains(s, sub)) })
	binary("equalFold", func(a, b string) goja.Value { return boolean(strings.EqualFold(a, b)) })
	binary("split", func(s, sep string) goja.Value {
		// An empty separator splits into Unicode code points, not UTF-16 units.
		parts := strings.Split(s, sep)
		items := make([]any, len(parts))
		for i, p := range parts {
			items[i] = p
		}
		return vm.NewArray(items...)
	})

	ternary("replace", func(s, old, repl string) goja.Value { return str(strings.Replace(s, old, repl, 1)) })
	ternary("replaceAll", func(s, old, repl string) goja.Value { return str(strings.ReplaceAll(s, old, repl)) })

	_ = ex.Set("join", func(call goja.FunctionCall) goja.Value {
		stringArgCount(vm, call, "join", 2)
		parts := stringListArg(vm, call, "join", 0)
		sep := stringArg(vm, call, "join", 1)
		return str(strings.Join(parts, sep))
	})

	_ = ex.Set("format", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			stringThrow(vm, "format: expected a format string and its arguments, got no arguments")
		}
		layout := stringArg(vm, call, "format", 0)
		return str(stringFormat(vm, layout, call.Arguments[1:]))
	})

	return ex, nil
}

// stringMaxListLength caps the number of elements a string function reads
// from an array argument. The length of a JS array is caller-controlled (a
// sparse `a.length = 2**32 - 1` costs nothing in JS), so it is checked before
// anything is allocated for it.
const stringMaxListLength = 1 << 20

// stringFormatMaxPrecision caps the precision of `%.Nf`.
const stringFormatMaxPrecision = 20

// stringFormat implements @sentrie/string.format. It supports a closed set of
// verbs and throws on anything else:
//
//	%s   a string argument, as is
//	%q   a string argument, double-quoted with Go escapes
//	%d   an integral, finite number, without a fraction or exponent
//	%f   a finite number with 6 decimal places
//	%.Nf a finite number with N decimal places, 0 <= N <= 20
//	     (%f and %.Nf round the exact binary value to nearest, ties to even)
//	%v   a string, number, boolean or null, as JavaScript's String() renders it
//	%%   a literal percent sign
//
// There are no flags or widths. The number of arguments must equal the number
// of verbs that consume one.
func stringFormat(vm *goja.Runtime, layout string, args []goja.Value) string {
	var b strings.Builder
	next := 0
	take := func(verb string) goja.Value {
		if next >= len(args) {
			stringThrow(vm, "format: missing argument for %s (verb %d)", verb, next+1)
		}
		v := args[next]
		next++
		return v
	}

	for i := 0; i < len(layout); i++ {
		c := layout[i]
		if c != '%' {
			b.WriteByte(c)
			continue
		}
		i++
		if i >= len(layout) {
			stringThrow(vm, "format: trailing '%%' at the end of the format string")
		}

		precision := -1
		if layout[i] == '.' {
			start := i + 1
			j := start
			for j < len(layout) && layout[j] >= '0' && layout[j] <= '9' {
				j++
			}
			if j == start || j >= len(layout) || layout[j] != 'f' {
				stringThrow(vm, "format: a precision is only supported as %%.Nf")
			}
			n, err := strconv.Atoi(layout[start:j])
			if err != nil || n > stringFormatMaxPrecision {
				stringThrow(vm, "format: precision in %%.%sf must be between 0 and %d", layout[start:j], stringFormatMaxPrecision)
			}
			precision = n
			i = j
		}

		switch verb := layout[i]; verb {
		case '%':
			b.WriteByte('%')
		case 's', 'q':
			v := take("%" + string(verb))
			if !goja.IsString(v) {
				stringThrow(vm, "format: %%%c expects a string, got %s", verb, stringTypeName(v))
			}
			if verb == 'q' {
				b.WriteString(strconv.Quote(v.String()))
			} else {
				b.WriteString(v.String())
			}
		case 'd':
			f := stringFormatNumber(vm, "%d", take("%d"))
			if f != math.Trunc(f) {
				stringThrow(vm, "format: %%d expects an integer, got %s", strconv.FormatFloat(f, 'g', -1, 64))
			}
			if f == 0 {
				f = 0 // drop the sign of negative zero
			}
			b.WriteString(strconv.FormatFloat(f, 'f', 0, 64))
		case 'f':
			if precision < 0 {
				precision = 6
			}
			f := stringFormatNumber(vm, "%f", take("%f"))
			if f == 0 {
				f = 0 // drop the sign of negative zero, as %d and %v do
			}
			b.WriteString(strconv.FormatFloat(f, 'f', precision, 64))
		case 'v':
			v := take("%v")
			switch {
			case goja.IsString(v), goja.IsNumber(v), goja.IsNull(v):
				b.WriteString(v.String())
			default:
				if bv, ok := v.Export().(bool); ok {
					b.WriteString(strconv.FormatBool(bv))
				} else {
					stringThrow(vm, "format: %%v expects a string, number, boolean or null, got %s", stringTypeName(v))
				}
			}
		default:
			r, _ := utf8.DecodeRuneInString(layout[i:])
			stringThrow(vm, "format: unsupported verb %%%c", r)
		}
	}

	if next != len(args) {
		stringThrow(vm, "format: %d arguments given but the format uses %d", len(args), next)
	}
	return b.String()
}

func stringFormatNumber(vm *goja.Runtime, verb string, v goja.Value) float64 {
	if !goja.IsNumber(v) {
		stringThrow(vm, "format: %s expects a number, got %s", verb, stringTypeName(v))
	}
	f := v.ToFloat()
	if math.IsNaN(f) || math.IsInf(f, 0) {
		stringThrow(vm, "format: %s expects a finite number, got %s", verb, v.String())
	}
	return f
}

// stringThrow throws a JS exception carrying a Go error.
func stringThrow(vm *goja.Runtime, format string, args ...any) {
	panic(vm.NewGoError(fmt.Errorf(format, args...)))
}

func stringArgCount(vm *goja.Runtime, call goja.FunctionCall, name string, want int) {
	if got := len(call.Arguments); got != want {
		stringThrow(vm, "%s: expected %d argument(s), got %d", name, want, got)
	}
}

func stringArg(vm *goja.Runtime, call goja.FunctionCall, name string, i int) string {
	v := call.Argument(i)
	if !goja.IsString(v) {
		stringThrow(vm, "%s: argument %d must be a string, got %s", name, i+1, stringTypeName(v))
	}
	return v.String()
}

func stringListArg(vm *goja.Runtime, call goja.FunctionCall, name string, i int) []string {
	obj, ok := call.Argument(i).(*goja.Object)
	if !ok || obj.ClassName() != "Array" {
		stringThrow(vm, "%s: argument %d must be an array of strings, got %s", name, i+1, stringTypeName(call.Argument(i)))
	}
	n := obj.Get("length").ToInteger()
	if n < 0 || n > stringMaxListLength {
		stringThrow(vm, "%s: argument %d has %d elements, more than the limit of %d", name, i+1, n, stringMaxListLength)
	}
	// Grow while validating, so a bad element throws before a large buffer exists.
	out := make([]string, 0, min(n, 64))
	for j := range int(n) {
		el := obj.Get(strconv.Itoa(j))
		if !goja.IsString(el) {
			stringThrow(vm, "%s: argument %d element %d must be a string, got %s", name, i+1, j, stringTypeName(el))
		}
		out = append(out, el.String())
	}
	return out
}

// stringTypeName names a value's JS type for error messages.
func stringTypeName(v goja.Value) string {
	switch {
	case v == nil || goja.IsUndefined(v):
		return "undefined"
	case goja.IsNull(v):
		return "null"
	case goja.IsString(v):
		return "string"
	case goja.IsNumber(v):
		return "number"
	case goja.IsBigInt(v):
		return "bigint"
	}
	if _, ok := v.(*goja.Symbol); ok {
		return "symbol"
	}
	if obj, ok := v.(*goja.Object); ok {
		if obj.ClassName() == "Array" {
			return "array"
		}
		return "object"
	}
	if _, ok := v.Export().(bool); ok {
		return "boolean"
	}
	return v.ExportType().String()
}
