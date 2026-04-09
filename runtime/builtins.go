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

package runtime

import (
	"context"
	"fmt"
	"slices"

	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/xerr"
)

func isUndefinedV(v box.Value) bool {
	return v.IsUndefined()
}

func toIntV(v box.Value) (int64, bool) {
	n, ok := v.NumberValue()
	if !ok {
		return 0, false
	}
	return int64(n), true
}

func copyMapDeep(m map[string]box.Value) map[string]box.Value {
	out := make(map[string]box.Value, len(m))
	for k, v := range m {
		if vm, ok := v.MapValue(); ok {
			out[k] = box.Map(copyMapDeep(vm))
		} else {
			out[k] = v
		}
	}
	return out
}

func mergeValueMaps(map1, map2 map[string]box.Value) map[string]box.Value {
	result := copyMapDeep(map1)
	for key, value2 := range map2 {
		if existing, exists := result[key]; exists {
			m1, ok1 := existing.MapValue()
			m2, ok2 := value2.MapValue()
			if ok1 && ok2 {
				result[key] = box.Map(mergeValueMaps(m1, m2))
				continue
			}
		}
		if nestedMap, ok := value2.MapValue(); ok {
			result[key] = box.Map(copyMapDeep(nestedMap))
			continue
		}
		result[key] = value2
	}
	return result
}

// BuiltinMerge merges two maps into a new map recursively.
func BuiltinMerge(_ context.Context, _ *CallSite, args ...box.Value) (box.Value, error) {
	if len(args) != 2 {
		return box.Undefined(), fmt.Errorf("merge requires 2 arguments")
	}
	m1, ok := args[0].MapValue()
	if !ok {
		return box.Undefined(), fmt.Errorf("first argument is not a map")
	}
	m2, ok := args[1].MapValue()
	if !ok {
		return box.Undefined(), fmt.Errorf("second argument is not a map")
	}
	return box.Map(mergeValueMaps(m1, m2)), nil
}

// BuiltinCount returns the length of a list, string, or map.
func BuiltinCount(_ context.Context, _ *CallSite, args ...box.Value) (box.Value, error) {
	if len(args) != 1 {
		return box.Undefined(), fmt.Errorf("count requires 1 argument")
	}
	if xs, ok := args[0].ListValue(); ok {
		return box.Number(len(xs)), nil
	}
	if s, ok := args[0].StringValue(); ok {
		return box.Number(len(s)), nil
	}
	if m, ok := args[0].MapValue(); ok {
		return box.Number(len(m)), nil
	}
	return box.Number(0), nil
}

// BuiltInError short-circuits execution with a formatted error.
func BuiltInError(_ context.Context, _ *CallSite, args ...box.Value) (box.Value, error) {
	if len(args) == 0 {
		return box.Undefined(), fmt.Errorf("error requires at least 1 argument")
	}
	fa := args
	if len(fa) == 1 {
		fa = append([]box.Value{box.String("%v")}, fa...)
	}
	format, ok := fa[0].StringValue()
	if !ok {
		return box.Undefined(), fmt.Errorf("error: first argument must be a format string")
	}
	rest := make([]any, 0, len(fa)-1)
	for _, a := range fa[1:] {
		x, err := box.TryToBoundaryAny(a)
		if err != nil {
			return box.Undefined(), fmt.Errorf("error: %w", err)
		}
		rest = append(rest, x)
	}
	return box.Undefined(), xerr.ErrInjected(format, rest...)
}

// BuiltinFlatten flattens nested lists to a controlled depth.
func BuiltinFlatten(_ context.Context, _ *CallSite, args ...box.Value) (box.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return box.Undefined(), fmt.Errorf("flatten requires 1 or 2 arguments")
	}
	if isUndefinedV(args[0]) {
		return box.Undefined(), nil
	}
	x, ok := args[0].ListValue()
	if !ok {
		return box.Undefined(), fmt.Errorf("flatten: first argument must be a list")
	}
	var depth int64 = 1
	if len(args) == 2 {
		if isUndefinedV(args[1]) {
			return box.Undefined(), nil
		}
		n, ok := toIntV(args[1])
		if !ok {
			return box.Undefined(), fmt.Errorf("flatten: second argument must be a non-negative integer")
		}
		if n < 0 {
			return box.Undefined(), fmt.Errorf("flatten: depth must be a non-negative integer")
		}
		depth = n
	}
	if depth == 0 {
		return box.List(x), nil
	}
	return flattenListBox(x, depth)
}

func flattenListBox(x []box.Value, depth int64) (box.Value, error) {
	if depth == 0 {
		return box.List(x), nil
	}
	result := make([]box.Value, 0)
	for _, elem := range x {
		if isUndefinedV(elem) {
			return box.Undefined(), nil
		}
		if nestedList, ok := elem.ListValue(); ok {
			for _, nestedElem := range nestedList {
				if isUndefinedV(nestedElem) {
					return box.Undefined(), nil
				}
			}
			flattened, err := flattenListBox(nestedList, depth-1)
			if err != nil {
				return box.Undefined(), err
			}
			if flattened.IsUndefined() {
				return box.Undefined(), nil
			}
			sub, _ := flattened.ListValue()
			result = append(result, sub...)
		} else {
			result = append(result, elem)
		}
	}
	return box.List(result), nil
}

// BuiltinFlattenDeep recursively flattens nested lists.
func BuiltinFlattenDeep(_ context.Context, _ *CallSite, args ...box.Value) (box.Value, error) {
	if len(args) != 1 {
		return box.Undefined(), fmt.Errorf("flatten_deep requires 1 argument")
	}
	if isUndefinedV(args[0]) {
		return box.Undefined(), nil
	}
	x, ok := args[0].ListValue()
	if !ok {
		return box.Undefined(), fmt.Errorf("flatten_deep: argument must be a list")
	}
	return flattenDeepBox(x)
}

func flattenDeepBox(x []box.Value) (box.Value, error) {
	result := make([]box.Value, 0)
	for _, elem := range x {
		if isUndefinedV(elem) {
			return box.Undefined(), nil
		}
		if nestedList, ok := elem.ListValue(); ok {
			flattened, err := flattenDeepBox(nestedList)
			if err != nil {
				return box.Undefined(), err
			}
			if flattened.IsUndefined() {
				return box.Undefined(), nil
			}
			sub, _ := flattened.ListValue()
			result = append(result, sub...)
		} else {
			result = append(result, elem)
		}
	}
	return box.List(result), nil
}

// BuiltinAsList normalizes one-or-many inputs to a list.
func BuiltinAsList(_ context.Context, _ *CallSite, args ...box.Value) (box.Value, error) {
	if len(args) != 1 {
		return box.Undefined(), fmt.Errorf("as_list requires 1 argument")
	}
	if isUndefinedV(args[0]) {
		return box.Undefined(), nil
	}
	v := args[0]
	if list, ok := v.ListValue(); ok {
		for _, elem := range list {
			if isUndefinedV(elem) {
				return box.Undefined(), nil
			}
		}
		return box.List(list), nil
	}
	return box.List([]box.Value{v}), nil
}

// BuiltinNormaliseList normalizes messy list inputs with one level of nesting.
func BuiltinNormaliseList(_ context.Context, _ *CallSite, args ...box.Value) (box.Value, error) {
	if len(args) != 1 {
		return box.Undefined(), fmt.Errorf("normalise_list requires 1 argument")
	}
	if isUndefinedV(args[0]) {
		return box.Undefined(), nil
	}
	v := args[0]
	var list []box.Value
	if l, ok := v.ListValue(); ok {
		list = l
	} else {
		list = []box.Value{v}
	}
	if slices.ContainsFunc(list, isUndefinedV) {
		return box.Undefined(), nil
	}
	for _, elem := range list {
		if nestedList, ok := elem.ListValue(); ok {
			for _, nestedElem := range nestedList {
				if isUndefinedV(nestedElem) {
					return box.Undefined(), nil
				}
				if _, ok := nestedElem.ListValue(); ok {
					return box.Undefined(), fmt.Errorf("normalise_list: input contains deeper than one level of nesting")
				}
			}
		}
	}
	result := make([]box.Value, 0)
	for _, elem := range list {
		if nestedList, ok := elem.ListValue(); ok {
			result = append(result, nestedList...)
		} else {
			result = append(result, elem)
		}
	}
	return box.List(result), nil
}

// Builtins is the registry of global built-in functions.
var Builtins = map[string]Builtin{
	"all":            BuiltinAll,
	"any":            BuiltinAny,
	"as_list":        BuiltinAsList,
	"count":          BuiltinCount,
	"distinct":       BuiltinDistinct,
	"error":          BuiltInError,
	"filter":         BuiltinFilter,
	"first":          BuiltinFirst,
	"flatten":        BuiltinFlatten,
	"flatten_deep":   BuiltinFlattenDeep,
	"map":            BuiltinMap,
	"merge":          BuiltinMerge,
	"normalise_list": BuiltinNormaliseList,
	"reduce":         BuiltinReduce,
}
