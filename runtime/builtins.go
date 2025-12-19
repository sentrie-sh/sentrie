// SPDX-License-Identifier: Apache-2.0

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

var Builtins = map[string]Builtin{
	"count": BuiltinCount,
	"error": BuiltInError,
	"merge": BuiltinMerge,
}
