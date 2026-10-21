// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"time"

	"github.com/dop251/goja"
	"github.com/fatih/structs"
	"github.com/sentrie-sh/sentrie/box"
)

// jsToSentrie converts a value returned by a module export into a boundary
// value (see box.FromBoundaryAny). The conversion table (#154):
//
//	undefined                 -> undefined (nested too)
//	null                      -> null
//	boolean, string           -> bool, string
//	number                    -> number (int64 when integral, else float64)
//	Array                     -> list, element by element
//	plain object              -> dict of its own enumerable string keys
//	Map with string keys      -> dict
//	Date                      -> Unix seconds as a number, like @sentrie/time.now()
//	String/Number/Boolean obj -> the wrapped primitive
//
// Anything else (functions, symbols, BigInt, Set, a Map with a non-string key,
// RegExp, Error, Promise, typed arrays, an invalid Date, a cyclic value) is an
// error rather than a lossy value.
//
// Conversion reads the value through goja from Go, which can run module code
// (getters, Proxy traps). A JS exception, an interrupt from the deadline or a
// stack overflow raised there becomes an error naming the path, never a panic,
// and ctx is checked as the walk goes so a huge value cannot outlive it.
func jsToSentrie(ctx context.Context, vm *goja.Runtime, v goja.Value) (_ any, err error) {
	c := jsConverter{ctx: ctx, vm: vm, onPath: map[*goja.Object]struct{}{}, at: "return value"}
	defer func() {
		if r := recover(); r != nil {
			jsErr, ok := r.(error)
			if !ok || !isGojaThrow(jsErr) {
				panic(r)
			}
			err = fmt.Errorf("converting %s: %w", c.at, jsErr)
		}
	}()
	return c.convert(v, c.at)
}

// maxReturnedListLen bounds the length of an array a module may return. A
// sparse array is cheap in JS but would be walked, and allocated, in full here.
const maxReturnedListLen = 1 << 20

// convertCheckEvery is how many values the walk converts between ctx checks.
const convertCheckEvery = 1024

// isGojaThrow reports whether a recovered panic is goja raising a JS throw,
// an interrupt or a stack overflow from Go-side property access.
func isGojaThrow(err error) bool {
	var ex *goja.Exception
	var interrupted *goja.InterruptedError
	var overflow *goja.StackOverflowError
	return errors.As(err, &ex) || errors.As(err, &interrupted) || errors.As(err, &overflow)
}

type jsConverter struct {
	ctx    context.Context
	vm     *goja.Runtime
	onPath map[*goja.Object]struct{}
	at     string // path being converted, for errors raised by a panic
	steps  int
}

// step counts one converted value and checks ctx every convertCheckEvery.
func (c *jsConverter) step(at string) error {
	c.steps++
	if c.steps%convertCheckEvery == 0 {
		if err := c.ctx.Err(); err != nil {
			return fmt.Errorf("converting %s: %w", at, err)
		}
	}
	return nil
}

func unsupportedJS(what, at, why string) error {
	return fmt.Errorf("unexpected return type %s at %s: %s", what, at, why)
}

func (c *jsConverter) convert(v goja.Value, at string) (any, error) {
	c.at = at
	if err := c.step(at); err != nil {
		return nil, err
	}
	if v == nil || goja.IsUndefined(v) {
		return box.ToBoundaryAny(box.Undefined()), nil
	}
	if goja.IsNull(v) {
		return nil, nil
	}
	obj, isObj := v.(*goja.Object)
	if !isObj {
		if _, ok := v.(*goja.Symbol); ok {
			return nil, unsupportedJS("symbol", at, "symbols cannot cross into Sentrie")
		}
		switch x := v.Export().(type) {
		case bool, string, int64, float64:
			return x, nil
		default:
			return nil, unsupportedJS(fmt.Sprintf("%T", x), at, "only booleans, strings and numbers are supported primitives")
		}
	}

	if _, seen := c.onPath[obj]; seen {
		return nil, unsupportedJS("cyclic value", at, "a value cannot contain itself")
	}
	c.onPath[obj] = struct{}{}
	defer delete(c.onPath, obj)

	switch class := obj.ClassName(); class {
	case "Date":
		t, ok := obj.Export().(time.Time)
		if !ok {
			return nil, unsupportedJS("Date", at, "invalid Date")
		}
		return t.Unix(), nil
	case "Array":
		n := obj.Get("length").ToInteger()
		if n > maxReturnedListLen {
			return nil, unsupportedJS("Array", at, fmt.Sprintf("length %d exceeds the limit of %d", n, maxReturnedListLen))
		}
		out := make([]any, 0, min(n, convertCheckEvery))
		for i := range n {
			item, err := c.convert(obj.Get(strconv.FormatInt(i, 10)), fmt.Sprintf("%s[%d]", at, i))
			if err != nil {
				return nil, err
			}
			out = append(out, item)
		}
		return out, nil
	case "String":
		return obj.String(), nil
	case "Number":
		return obj.ToNumber().Export(), nil
	case "Boolean":
		return obj.ToBoolean(), nil
	case "Object":
		// goja reports Map, Set and Promise as class "Object"; their export
		// types tell them apart from plain objects and wrapped Go values.
		switch obj.ExportType() {
		case reflect.TypeFor[[][2]any]():
			return c.convertMap(obj, at)
		case reflect.TypeFor[[]any]():
			return nil, unsupportedJS("Set", at, "convert it to an array first")
		case reflect.TypeFor[goja.Proxy]():
			return nil, unsupportedJS("Proxy", at, "return the target object instead")
		case reflect.TypeFor[map[string]any]():
		default:
			// A Go value a builtin wrapped (struct, typed map or slice).
			return exportGoValue(obj, at)
		}
		out := map[string]any{}
		for _, k := range obj.Keys() {
			item, err := c.convert(obj.Get(k), at+"."+k)
			if err != nil {
				return nil, err
			}
			out[k] = item
		}
		return out, nil
	case "Function":
		return nil, unsupportedJS("Function", at, "functions cannot cross into Sentrie")
	default:
		return nil, unsupportedJS(class, at, "convert it to a plain object, array or primitive first")
	}
}

func (c *jsConverter) convertMap(obj *goja.Object, at string) (any, error) {
	out := map[string]any{}
	var convErr error
	// ForOf panics on a JS throw; jsToSentrie's recover turns that into an error.
	c.vm.ForOf(obj, func(entry goja.Value) bool {
		pair := entry.ToObject(c.vm)
		key := pair.Get("0")
		if _, isObj := key.(*goja.Object); isObj || key.ExportType() == nil || key.ExportType().Kind() != reflect.String {
			convErr = unsupportedJS("Map", at, fmt.Sprintf("key %s is not a string", key.String()))
			return false
		}
		k := key.String()
		item, err := c.convert(pair.Get("1"), at+".get("+strconv.Quote(k)+")")
		if err != nil {
			convErr = err
			return false
		}
		out[k] = item
		return true
	})
	if convErr != nil {
		return nil, convErr
	}
	return out, nil
}

// acceptedGoKinds are the Go kinds a builtin-wrapped value may have.
var acceptedGoKinds = []reflect.Kind{reflect.Map, reflect.Slice, reflect.Array, reflect.String, reflect.Int64, reflect.Float64, reflect.Bool, reflect.Struct}

// exportGoValue keeps the pre-#154 handling for Go values wrapped by builtins.
func exportGoValue(obj *goja.Object, at string) (any, error) {
	if !slices.Contains(acceptedGoKinds, obj.ExportType().Kind()) {
		return nil, unsupportedJS(obj.ExportType().String(), at, "unsupported Go value")
	}
	result := obj.Export()
	if structs.IsStruct(result) {
		result = structs.Map(result)
	}
	return result, nil
}
