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

package js

import (
	"errors"
	"reflect"
	"sort"

	"github.com/dop251/goja"
)

var BuiltinCollectionGo = func(vm *goja.Runtime) (*goja.Object, error) {
	ex := vm.NewObject()

	// List utilities with list_ prefix
	_ = ex.Set("list_includes", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 2 {
			return vm.NewGoError(errors.New("list_includes requires exactly 2 arguments"))
		}
		arrVal := call.Argument(0)
		item := call.Argument(1)

		arr := arrVal.Export()
		arrValue := reflect.ValueOf(arr)
		if arrValue.Kind() != reflect.Slice && arrValue.Kind() != reflect.Array {
			return vm.ToValue(false)
		}

		itemExport := item.Export()
		for i := 0; i < arrValue.Len(); i++ {
			elem := arrValue.Index(i).Interface()
			if reflect.DeepEqual(elem, itemExport) {
				return vm.ToValue(true)
			}
		}
		return vm.ToValue(false)
	})

	_ = ex.Set("list_indexOf", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 2 {
			return vm.NewGoError(errors.New("list_indexOf requires exactly 2 arguments"))
		}
		arrVal := call.Argument(0)
		item := call.Argument(1)

		arr := arrVal.Export()
		arrValue := reflect.ValueOf(arr)
		if arrValue.Kind() != reflect.Slice && arrValue.Kind() != reflect.Array {
			return vm.ToValue(-1)
		}

		itemExport := item.Export()
		for i := 0; i < arrValue.Len(); i++ {
			elem := arrValue.Index(i).Interface()
			if reflect.DeepEqual(elem, itemExport) {
				return vm.ToValue(i)
			}
		}
		return vm.ToValue(-1)
	})

	_ = ex.Set("list_lastIndexOf", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 2 {
			return vm.NewGoError(errors.New("list_lastIndexOf requires exactly 2 arguments"))
		}
		arrVal := call.Argument(0)
		item := call.Argument(1)

		arr := arrVal.Export()
		arrValue := reflect.ValueOf(arr)
		if arrValue.Kind() != reflect.Slice && arrValue.Kind() != reflect.Array {
			return vm.ToValue(-1)
		}

		itemExport := item.Export()
		lastIndex := -1
		for i := 0; i < arrValue.Len(); i++ {
			elem := arrValue.Index(i).Interface()
			if reflect.DeepEqual(elem, itemExport) {
				lastIndex = i
			}
		}
		return vm.ToValue(lastIndex)
	})

	_ = ex.Set("list_sort", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return vm.NewGoError(errors.New("list_sort requires exactly 1 argument"))
		}
		arrVal := call.Argument(0)

		arr := arrVal.Export()
		arrValue := reflect.ValueOf(arr)
		if arrValue.Kind() != reflect.Slice {
			return vm.NewGoError(errors.New("list_sort requires an array"))
		}

		// Convert to []interface{} for sorting
		arrInterface := make([]interface{}, arrValue.Len())
		for i := 0; i < arrValue.Len(); i++ {
			arrInterface[i] = arrValue.Index(i).Interface()
		}

		// Sort using type-aware comparison
		sort.Slice(arrInterface, func(i, j int) bool {
			return compareValues(arrInterface[i], arrInterface[j]) < 0
		})

		return vm.ToValue(arrInterface)
	})

	_ = ex.Set("list_unique", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return vm.NewGoError(errors.New("list_unique requires exactly 1 argument"))
		}
		arrVal := call.Argument(0)

		arr := arrVal.Export()
		arrValue := reflect.ValueOf(arr)
		if arrValue.Kind() != reflect.Slice && arrValue.Kind() != reflect.Array {
			return vm.NewGoError(errors.New("list_unique requires an array"))
		}

		seen := make(map[interface{}]bool)
		var result []interface{}
		for i := 0; i < arrValue.Len(); i++ {
			elem := arrValue.Index(i).Interface()
			// Use deep equal for comparison in map keys would be complex
			// For now, use simple equality check
			key := reflect.ValueOf(elem)
			if !key.IsValid() || key.CanInterface() {
				if !seen[elem] {
					seen[elem] = true
					result = append(result, elem)
				}
			}
		}

		return vm.ToValue(result)
	})

	_ = ex.Set("list_chunk", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 2 {
			return vm.NewGoError(errors.New("list_chunk requires exactly 2 arguments"))
		}
		arrVal := call.Argument(0)
		size := int(call.Argument(1).ToInteger())

		if size <= 0 {
			return vm.NewGoError(errors.New("list_chunk size must be positive"))
		}

		arr := arrVal.Export()
		arrValue := reflect.ValueOf(arr)
		if arrValue.Kind() != reflect.Slice && arrValue.Kind() != reflect.Array {
			return vm.NewGoError(errors.New("list_chunk requires an array"))
		}

		var chunks [][]interface{}
		var currentChunk []interface{}

		for i := 0; i < arrValue.Len(); i++ {
			elem := arrValue.Index(i).Interface()
			currentChunk = append(currentChunk, elem)

			if len(currentChunk) == size {
				chunks = append(chunks, currentChunk)
				currentChunk = nil
			}
		}

		if len(currentChunk) > 0 {
			chunks = append(chunks, currentChunk)
		}

		return vm.ToValue(chunks)
	})

	_ = ex.Set("list_flatten", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return vm.NewGoError(errors.New("list_flatten requires exactly 1 argument"))
		}
		arrVal := call.Argument(0)

		arr := arrVal.Export()
		arrValue := reflect.ValueOf(arr)
		if arrValue.Kind() != reflect.Slice && arrValue.Kind() != reflect.Array {
			return vm.NewGoError(errors.New("list_flatten requires an array"))
		}

		var result []interface{}
		flattenArray(arrValue, &result)

		return vm.ToValue(result)
	})

	// Map utilities with map_ prefix
	_ = ex.Set("map_keys", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return vm.NewGoError(errors.New("map_keys requires exactly 1 argument"))
		}
		mapVal := call.Argument(0)

		mp := mapVal.Export()
		mapValue := reflect.ValueOf(mp)
		if mapValue.Kind() != reflect.Map {
			return vm.NewGoError(errors.New("map_keys requires a map"))
		}

		var keys []interface{}
		for _, key := range mapValue.MapKeys() {
			keys = append(keys, key.Interface())
		}

		return vm.ToValue(keys)
	})

	_ = ex.Set("map_values", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return vm.NewGoError(errors.New("map_values requires exactly 1 argument"))
		}
		mapVal := call.Argument(0)

		mp := mapVal.Export()
		mapValue := reflect.ValueOf(mp)
		if mapValue.Kind() != reflect.Map {
			return vm.NewGoError(errors.New("map_values requires a map"))
		}

		var values []interface{}
		for _, key := range mapValue.MapKeys() {
			values = append(values, mapValue.MapIndex(key).Interface())
		}

		return vm.ToValue(values)
	})

	_ = ex.Set("map_entries", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return vm.NewGoError(errors.New("map_entries requires exactly 1 argument"))
		}
		mapVal := call.Argument(0)

		mp := mapVal.Export()
		mapValue := reflect.ValueOf(mp)
		if mapValue.Kind() != reflect.Map {
			return vm.NewGoError(errors.New("map_entries requires a map"))
		}

		var entries [][]interface{}
		for _, key := range mapValue.MapKeys() {
			entry := []interface{}{key.Interface(), mapValue.MapIndex(key).Interface()}
			entries = append(entries, entry)
		}

		return vm.ToValue(entries)
	})

	_ = ex.Set("map_has", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 2 {
			return vm.NewGoError(errors.New("map_has requires exactly 2 arguments"))
		}
		mapVal := call.Argument(0)
		key := call.Argument(1)

		mp := mapVal.Export()
		mapValue := reflect.ValueOf(mp)
		if mapValue.Kind() != reflect.Map {
			return vm.ToValue(false)
		}

		keyValue := reflect.ValueOf(key.Export())
		if !keyValue.IsValid() {
			return vm.ToValue(false)
		}

		return vm.ToValue(mapValue.MapIndex(keyValue).IsValid())
	})

	_ = ex.Set("map_get", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 || len(call.Arguments) > 3 {
			return vm.NewGoError(errors.New("map_get requires 2 or 3 arguments"))
		}
		mapVal := call.Argument(0)
		key := call.Argument(1)
		var defaultValue goja.Value
		if len(call.Arguments) > 2 {
			defaultValue = call.Argument(2)
		}

		mp := mapVal.Export()
		mapValue := reflect.ValueOf(mp)
		if mapValue.Kind() != reflect.Map {
			if defaultValue != nil && defaultValue != goja.Undefined() {
				return defaultValue
			}
			return goja.Undefined()
		}

		keyValue := reflect.ValueOf(key.Export())
		if !keyValue.IsValid() {
			if defaultValue != nil && defaultValue != goja.Undefined() {
				return defaultValue
			}
			return goja.Undefined()
		}

		value := mapValue.MapIndex(keyValue)
		if !value.IsValid() {
			if defaultValue != nil && defaultValue != goja.Undefined() {
				return defaultValue
			}
			return goja.Undefined()
		}

		return vm.ToValue(value.Interface())
	})

	_ = ex.Set("map_size", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return vm.NewGoError(errors.New("map_size requires exactly 1 argument"))
		}
		mapVal := call.Argument(0)

		mp := mapVal.Export()
		mapValue := reflect.ValueOf(mp)
		if mapValue.Kind() != reflect.Map {
			return vm.NewGoError(errors.New("map_size requires a map"))
		}

		return vm.ToValue(mapValue.Len())
	})

	_ = ex.Set("map_isEmpty", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return vm.NewGoError(errors.New("map_isEmpty requires exactly 1 argument"))
		}
		mapVal := call.Argument(0)

		mp := mapVal.Export()
		mapValue := reflect.ValueOf(mp)
		if mapValue.Kind() != reflect.Map {
			return vm.NewGoError(errors.New("map_isEmpty requires a map"))
		}

		return vm.ToValue(mapValue.Len() == 0)
	})

	_ = ex.Set("map_merge", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			return vm.NewGoError(errors.New("map_merge requires at least 2 arguments"))
		}

		mp1 := call.Argument(0).Export()
		map1Value := reflect.ValueOf(mp1)
		if map1Value.Kind() != reflect.Map {
			return vm.NewGoError(errors.New("map_merge requires maps as arguments"))
		}

		// Create a new map and copy map1 into it
		resultType := reflect.MapOf(map1Value.Type().Key(), map1Value.Type().Elem())
		result := reflect.MakeMap(resultType)

		// Copy all entries from map1
		for _, key := range map1Value.MapKeys() {
			result.SetMapIndex(key, map1Value.MapIndex(key))
		}

		// Merge all other maps
		for i := 1; i < len(call.Arguments); i++ {
			mp2 := call.Argument(i).Export()
			map2Value := reflect.ValueOf(mp2)
			if map2Value.Kind() != reflect.Map {
				return vm.NewGoError(errors.New("map_merge requires maps as arguments"))
			}

			// Copy all entries from map2 (overwrites if key exists)
			for _, key := range map2Value.MapKeys() {
				result.SetMapIndex(key, map2Value.MapIndex(key))
			}
		}

		return vm.ToValue(result.Interface())
	})

	return ex, nil
}

// compareValues compares two values for sorting
func compareValues(a, b interface{}) int {
	aVal := reflect.ValueOf(a)
	bVal := reflect.ValueOf(b)

	// Handle numeric types
	if aVal.Kind() == reflect.Float64 && bVal.Kind() == reflect.Float64 {
		af := aVal.Float()
		bf := bVal.Float()
		if af < bf {
			return -1
		} else if af > bf {
			return 1
		}
		return 0
	}

	// Handle strings
	if aVal.Kind() == reflect.String && bVal.Kind() == reflect.String {
		as := aVal.String()
		bs := bVal.String()
		if as < bs {
			return -1
		} else if as > bs {
			return 1
		}
		return 0
	}

	// Default: convert to string and compare
	as := reflect.ValueOf(a).String()
	bs := reflect.ValueOf(b).String()
	if as < bs {
		return -1
	} else if as > bs {
		return 1
	}
	return 0
}

// flattenArray recursively flattens nested arrays
func flattenArray(arr reflect.Value, result *[]interface{}) {
	for i := 0; i < arr.Len(); i++ {
		elem := arr.Index(i)
		if elem.Kind() == reflect.Slice || elem.Kind() == reflect.Array {
			flattenArray(elem, result)
		} else {
			*result = append(*result, elem.Interface())
		}
	}
}
