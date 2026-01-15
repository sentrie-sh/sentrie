// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2025 Binaek Sarkar
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

	"github.com/sentrie-sh/sentrie/xerr"
)

type Builtin func(ctx context.Context, args []any) (any, error)

// builtin merge - merge two maps into a new map recursively
func BuiltinMerge(ctx context.Context, args []any) (any, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("merge requires 2 arguments")
	}

	map1, ok := args[0].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("first argument is not a map")
	}

	map2, ok := args[1].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("second argument is not a map")
	}

	return mergeMaps(map1, map2), nil
}

// mergeMaps merges two maps into a new map recursively
func mergeMaps(map1, map2 map[string]any) map[string]any {
	// Start with a fresh result map sized approximately to hold both inputs
	result := make(map[string]any, len(map1)+len(map2))

	// Copy entries from map1. If a value is a map, copy it recursively so we
	// do not alias nested maps from the inputs.
	for key, value := range map1 {
		if nestedMap, ok := value.(map[string]any); ok {
			result[key] = mergeMaps(nestedMap, map[string]any{})
			continue
		}
		result[key] = value
	}

	// Merge/overwrite with entries from map2. When both values are maps, merge
	// recursively. Otherwise, values from map2 replace those from map1.
	for key, value2 := range map2 {
		if existing, exists := result[key]; exists {
			m1, ok1 := existing.(map[string]any)
			m2, ok2 := value2.(map[string]any)
			if ok1 && ok2 {
				result[key] = mergeMaps(m1, m2)
				continue
			}
		}

		// If the incoming value is a map, copy it to avoid aliasing
		if nestedMap, ok := value2.(map[string]any); ok {
			result[key] = mergeMaps(nestedMap, map[string]any{})
			continue
		}
		result[key] = value2
	}

	return result
}

func BuiltinCount(ctx context.Context, args []any) (any, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("count requires 1 argument")
	}

	asList, ok := args[0].([]any)
	if ok {
		return len(asList), nil
	}

	asString, ok := args[0].(string)
	if ok {
		return len(asString), nil
	}

	asMap, ok := args[0].(map[string]any)
	if ok {
		return len(asMap), nil
	}

	return 0, nil
}

// builtin error - short-circuit the execution and float up the error
func BuiltInError(ctx context.Context, args []any) (any, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("error requires at least 1 argument")
	}

	if len(args) == 1 {
		args = append([]any{"%v"}, args...)
	}

	format := args[0].(string)
	args = args[1:]
	return nil, xerr.ErrInjected(format, args...)
}

// builtin flatten - flatten nested lists to a controlled depth
func BuiltinFlatten(ctx context.Context, args []any) (any, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("flatten requires 1 or 2 arguments")
	}

	// Check for unknown (undefined) input
	if IsUndefined(args[0]) {
		return Undefined, nil
	}

	x, ok := args[0].([]any)
	if !ok {
		return nil, fmt.Errorf("flatten: first argument must be a list")
	}

	var depth int64 = 1 // default depth
	if len(args) == 2 {
		if IsUndefined(args[1]) {
			return Undefined, nil
		}
		n, ok := toInt(args[1])
		if !ok {
			return nil, fmt.Errorf("flatten: second argument must be a non-negative integer")
		}
		if n < 0 {
			return nil, fmt.Errorf("flatten: depth must be a non-negative integer")
		}
		depth = n
	}

	if depth == 0 {
		return x, nil
	}

	return flattenList(x, depth)
}

// flattenList flattens a list to the specified depth
func flattenList(x []any, depth int64) (any, error) {
	if depth == 0 {
		return x, nil
	}

	result := make([]any, 0)
	for _, elem := range x {
		// Check for unknown (undefined) - propagate unknown
		if IsUndefined(elem) {
			return Undefined, nil
		}

		// If element is a list, flatten it
		if nestedList, ok := elem.([]any); ok {
			// Check if nested list contains unknown
			for _, nestedElem := range nestedList {
				if IsUndefined(nestedElem) {
					return Undefined, nil
				}
			}
			// Recursively flatten with depth-1
			flattened, err := flattenList(nestedList, depth-1)
			if err != nil {
				return nil, err
			}
			if IsUndefined(flattened) {
				return Undefined, nil
			}
			flattenedList, ok := flattened.([]any)
			if !ok {
				return nil, fmt.Errorf("flatten: internal error - expected list result")
			}
			result = append(result, flattenedList...)
		} else {
			// Non-list element, preserve as-is
			result = append(result, elem)
		}
	}

	return result, nil
}

// builtin flatten_deep - recursively flatten nested lists
func BuiltinFlattenDeep(ctx context.Context, args []any) (any, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("flatten_deep requires 1 argument")
	}

	// Check for unknown (undefined) input
	if IsUndefined(args[0]) {
		return Undefined, nil
	}

	x, ok := args[0].([]any)
	if !ok {
		return nil, fmt.Errorf("flatten_deep: argument must be a list")
	}

	return flattenDeep(x)
}

// flattenDeep recursively flattens a list to arbitrary depth
func flattenDeep(x []any) (any, error) {
	result := make([]any, 0)
	for _, elem := range x {
		// Check for unknown (undefined) - propagate unknown
		if IsUndefined(elem) {
			return Undefined, nil
		}

		// If element is a list, recursively flatten it
		if nestedList, ok := elem.([]any); ok {
			flattened, err := flattenDeep(nestedList)
			if err != nil {
				return nil, err
			}
			if IsUndefined(flattened) {
				return Undefined, nil
			}
			flattenedList, ok := flattened.([]any)
			if !ok {
				return nil, fmt.Errorf("flatten_deep: internal error - expected list result")
			}
			result = append(result, flattenedList...)
		} else {
			// Non-list element, preserve as-is
			result = append(result, elem)
		}
	}

	return result, nil
}

// builtin as_list - normalize "one-or-many" inputs
func BuiltinAsList(ctx context.Context, args []any) (any, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("as_list requires 1 argument")
	}

	// Check for unknown (undefined) input
	if IsUndefined(args[0]) {
		return Undefined, nil
	}

	v := args[0]

	// If v is already a list, return it unchanged
	if list, ok := v.([]any); ok {
		// Check for unknown elements in the list
		for _, elem := range list {
			if IsUndefined(elem) {
				return Undefined, nil
			}
		}
		return list, nil
	}

	// Otherwise, wrap in a single-element list
	return []any{v}, nil
}

// builtin normalise_list - normalize messy list inputs with one level of nesting
func BuiltinNormaliseList(ctx context.Context, args []any) (any, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("normalise_list requires 1 argument")
	}

	// Check for unknown (undefined) input
	if IsUndefined(args[0]) {
		return Undefined, nil
	}

	v := args[0]

	// First apply as_list: wrap non-list values
	var list []any
	if l, ok := v.([]any); ok {
		list = l
	} else {
		list = []any{v}
	}

	// Check for unknown elements
	if slices.ContainsFunc(list, IsUndefined) {
		return Undefined, nil
	}

	// Check for deeper than one level of nesting before flattening
	// This ensures we error on list<list<list<T>>> structures
	for _, elem := range list {
		if nestedList, ok := elem.([]any); ok {
			for _, nestedElem := range nestedList {
				if IsUndefined(nestedElem) {
					return Undefined, nil
				}
				// Check for deeper nesting (error case)
				if _, ok := nestedElem.([]any); ok {
					return nil, fmt.Errorf("normalise_list: input contains deeper than one level of nesting")
				}
			}
		}
	}

	// Then flatten exactly one level
	result := make([]any, 0)
	for _, elem := range list {
		if nestedList, ok := elem.([]any); ok {
			// We already checked for unknown and deeper nesting above
			result = append(result, nestedList...)
		} else {
			result = append(result, elem)
		}
	}

	return result, nil
}

var Builtins = map[string]Builtin{
	"as_list":        BuiltinAsList,
	"count":          BuiltinCount,
	"error":          BuiltInError,
	"flatten":        BuiltinFlatten,
	"flatten_deep":   BuiltinFlattenDeep,
	"merge":          BuiltinMerge,
	"normalise_list": BuiltinNormaliseList,
}
