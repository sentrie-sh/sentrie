// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
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
// WeakMap, WeakSet, RegExp, Error, Promise, typed arrays, ArrayBuffer,
// DataView, an invalid Date, a cyclic value) is an error rather than a lossy
// value. goja has no WeakRef or FinalizationRegistry, so neither can reach here.
func jsToSentrie(vm *goja.Runtime, v goja.Value) (any, error) {
	c := jsConverter{vm: vm, onPath: map[*goja.Object]struct{}{}}
	return c.convert(v, "return value")
}

type jsConverter struct {
	vm     *goja.Runtime
	onPath map[*goja.Object]struct{}
	brands *jsBrandChecks // resolved on first use
}

// jsBrandChecks holds native intrinsics that test for a built-in's internal
// slot. goja reports WeakMap, WeakSet, DataView and typed arrays as class
// "Object" with export types shared by plain objects or Go slices, so neither
// ClassName nor ExportType can tell them apart (#224). A brand check catches
// subclasses and objects whose prototype was swapped, and none of these
// functions runs module code.
//
// They are read from the VM's globals when a conversion first needs them. A
// module that overwrites those globals only gets back the {} it could have
// returned directly, so this is not a trust boundary.
type jsBrandChecks struct {
	weakMapHas goja.Callable // WeakMap.prototype.has: throws without the slot
	weakSetHas goja.Callable // WeakSet.prototype.has: throws without the slot
	isView     goja.Callable // ArrayBuffer.isView: typed arrays and DataView
	isViewThis goja.Value
}

func (c *jsConverter) brandChecks() *jsBrandChecks {
	if c.brands != nil {
		return c.brands
	}
	global := c.vm.GlobalObject()
	method := func(owner goja.Value, name string) goja.Callable {
		obj, ok := owner.(*goja.Object)
		if !ok {
			return nil
		}
		fn, _ := goja.AssertFunction(obj.Get(name))
		return fn
	}
	prototypeMethod := func(ctor, name string) goja.Callable {
		ctorObj, ok := global.Get(ctor).(*goja.Object)
		if !ok {
			return nil
		}
		return method(ctorObj.Get("prototype"), name)
	}
	arrayBuffer := global.Get("ArrayBuffer")
	c.brands = &jsBrandChecks{
		weakMapHas: prototypeMethod("WeakMap", "has"),
		weakSetHas: prototypeMethod("WeakSet", "has"),
		isView:     method(arrayBuffer, "isView"),
		isViewThis: arrayBuffer,
	}
	return c.brands
}

// hasSlot reports whether calling the brand check with obj as its receiver
// succeeds. An interrupt (the call's context ended) is returned, not read as
// "no slot".
func hasSlot(check goja.Callable, obj *goja.Object) (bool, error) {
	if check == nil {
		return false, nil
	}
	_, err := check(obj, obj)
	if err == nil {
		return true, nil
	}
	var interrupted *goja.InterruptedError
	if errors.As(err, &interrupted) {
		return false, err
	}
	return false, nil
}

// opaqueBuiltin names the built-in obj is when it is one whose contents cannot
// be read as a plain object: a WeakMap, WeakSet, ArrayBuffer, typed array or
// DataView. It returns "" for anything else.
func (c *jsConverter) opaqueBuiltin(obj *goja.Object) (string, error) {
	if obj.ExportType() == reflect.TypeFor[goja.ArrayBuffer]() {
		return "ArrayBuffer", nil
	}
	b := c.brandChecks()
	if b.isView != nil {
		isView, err := b.isView(b.isViewThis, obj)
		if err != nil {
			return "", err
		}
		if isView.ToBoolean() {
			// Typed arrays and DataView carry their name as Symbol.toStringTag.
			if tag := obj.GetSymbol(goja.SymToStringTag); tag != nil && !goja.IsUndefined(tag) {
				return tag.String(), nil
			}
			return "ArrayBuffer view", nil
		}
	}
	for _, check := range []struct {
		name string
		fn   goja.Callable
	}{{"WeakMap", b.weakMapHas}, {"WeakSet", b.weakSetHas}} {
		ok, err := hasSlot(check.fn, obj)
		if err != nil {
			return "", err
		}
		if ok {
			return check.name, nil
		}
	}
	return "", nil
}

func unsupportedJS(what, at, why string) error {
	return fmt.Errorf("unexpected return type %s at %s: %s", what, at, why)
}

func (c *jsConverter) convert(v goja.Value, at string) (any, error) {
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
		out := make([]any, 0, n)
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
		return c.convert(c.vm.ToValue(obj.ToNumber().Export()), at)
	case "Boolean":
		return obj.ToBoolean(), nil
	case "Object":
		// goja reports Map, Set and Promise as class "Object"; their export
		// types tell them apart from plain objects and wrapped Go values. The
		// built-ins whose export type does not (#224) need a brand check.
		switch obj.ExportType() {
		case reflect.TypeFor[[][2]any](), reflect.TypeFor[[]any]():
		default:
			what, err := c.opaqueBuiltin(obj)
			if err != nil {
				return nil, err
			}
			if what != "" {
				return nil, unsupportedJS(what, at, opaqueHint(what))
			}
		}
		switch obj.ExportType() {
		case reflect.TypeFor[[][2]any]():
			return c.convertMap(obj, at)
		case reflect.TypeFor[[]any]():
			return nil, unsupportedJS("Set", at, "convert it to an array first")
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

func opaqueHint(what string) string {
	switch what {
	case "WeakMap", "WeakSet":
		return "its entries cannot be enumerated; return a Map, array or plain object instead"
	default:
		return "binary data does not cross into Sentrie; convert it with Array.from first"
	}
}

func (c *jsConverter) convertMap(obj *goja.Object, at string) (any, error) {
	out := map[string]any{}
	var convErr error
	if ex := c.vm.Try(func() {
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
	}); ex != nil {
		return nil, ex
	}
	if convErr != nil {
		return nil, convErr
	}
	return out, nil
}

// exportGoValue keeps the pre-#154 handling for Go values wrapped by builtins.
func exportGoValue(obj *goja.Object, at string) (any, error) {
	accepted := []reflect.Kind{reflect.Map, reflect.Slice, reflect.Array, reflect.String, reflect.Int64, reflect.Float64, reflect.Bool, reflect.Struct}
	if !slices.Contains(accepted, obj.ExportType().Kind()) {
		return nil, unsupportedJS(obj.ExportType().String(), at, "unsupported Go value")
	}
	result := obj.Export()
	if structs.IsStruct(result) {
		result = structs.Map(result)
	}
	return result, nil
}
