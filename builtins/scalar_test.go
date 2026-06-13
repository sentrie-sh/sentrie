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

package builtins

import (
	"testing"

	"github.com/sentrie-sh/sentrie/box"
	"github.com/stretchr/testify/require"
)

// Test BuiltinFlatten

func TestFlatten_DefaultDepth(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// flatten(x) should flatten exactly one level
	input := []any{[]any{1.0, 2.0}, []any{3.0, 4.0}}
	result, err := invoke(t, "flatten", env, boxArgs(input)...)
	require.NoError(t, err)
	require.Equal(t, []any{1.0, 2.0, 3.0, 4.0}, result.Any())
}

func TestFlatten_ExplicitDepth1(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// flatten(x, 1) should be equivalent to flatten(x)
	input := []any{[]any{1.0, 2.0}, []any{3.0, 4.0}}
	result, err := invoke(t, "flatten", env, boxArgs(input, 1)...)
	require.NoError(t, err)
	require.Equal(t, []any{1.0, 2.0, 3.0, 4.0}, result.Any())
}

func TestFlatten_Depth0(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// flatten(x, 0) should return x unchanged
	input := []any{[]any{1.0, 2.0}, []any{3.0, 4.0}}
	result, err := invoke(t, "flatten", env, boxArgs(input, 0)...)
	require.NoError(t, err)
	require.Equal(t, input, result.Any())
}

func TestFlatten_Depth2(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// flatten(x, 2) should flatten two levels
	input := []any{[]any{[]any{1.0, 2.0}}, []any{[]any{3.0, 4.0}}}
	result, err := invoke(t, "flatten", env, boxArgs(input, 2)...)
	require.NoError(t, err)
	require.Equal(t, []any{1.0, 2.0, 3.0, 4.0}, result.Any())
}

func TestFlatten_PreservesOrder(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// Flattening should preserve order
	input := []any{[]any{1.0, 2.0}, 5.0, []any{3.0, 4.0}}
	result, err := invoke(t, "flatten", env, boxArgs(input)...)
	require.NoError(t, err)
	require.Equal(t, []any{1.0, 2.0, 5.0, 3.0, 4.0}, result.Any())
}

func TestFlatten_NonListLeaves(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// Non-list values should be treated as leaves
	input := []any{1.0, []any{2.0, 3.0}, 4.0}
	result, err := invoke(t, "flatten", env, boxArgs(input)...)
	require.NoError(t, err)
	require.Equal(t, []any{1.0, 2.0, 3.0, 4.0}, result.Any())
}

func TestFlatten_EmptyList(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// Empty list should return empty list
	input := []any{}
	result, err := invoke(t, "flatten", env, boxArgs(input)...)
	require.NoError(t, err)
	require.Equal(t, []any{}, result.Any())
}

func TestFlatten_UnknownInput(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// Unknown (undefined) input should propagate unknown
	result, err := invoke(t, "flatten", env, boxArgs(box.Undefined())...)
	require.NoError(t, err)
	require.Equal(t, box.Undefined(), result) // Undefined represents unknown
}

func TestFlatten_UnknownInNestedList(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// Unknown in nested list should propagate unknown
	input := []any{[]any{1.0, box.Undefined(), 2.0}}
	result, err := invoke(t, "flatten", env, boxArgs(input)...)
	require.NoError(t, err)
	require.Equal(t, box.Undefined(), result) // Undefined represents unknown
}

func TestFlatten_ErrorNonList(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// Non-list input should return error
	_, err := invoke(t, "flatten", env, boxArgs("not a list")...)
	require.Error(t, err)
	require.ErrorContains(t, err, "must be a list")
}

func TestFlatten_ErrorInvalidDepth(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// Negative depth should return error
	input := []any{[]any{1.0, 2.0}}
	_, err := invoke(t, "flatten", env, boxArgs(input, -1)...)
	require.Error(t, err)
	require.ErrorContains(t, err, "non-negative integer")
}

func TestFlatten_ErrorInvalidDepthType(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// Non-integer depth should return error
	input := []any{[]any{1.0, 2.0}}
	_, err := invoke(t, "flatten", env, boxArgs(input, "not an int")...)
	require.Error(t, err)
	require.ErrorContains(t, err, "non-negative integer")
}

func TestFlatten_ErrorWrongArgCount(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// Wrong argument count should return error
	_, err := invoke(t, "flatten", env, boxArgs()...)
	require.Error(t, err)
	require.ErrorContains(t, err, "1 or 2 arguments")

	_, err = invoke(t, "flatten", env, boxArgs([]any{1.0}, 1, 2)...)
	require.Error(t, err)
	require.ErrorContains(t, err, "1 or 2 arguments")
}

func TestFlatten_UnknownDepth(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// Unknown depth should propagate unknown
	input := []any{[]any{1.0, 2.0}}
	result, err := invoke(t, "flatten", env, boxArgs(input, box.Undefined())...)
	require.NoError(t, err)
	require.Equal(t, box.Undefined(), result) // Undefined represents unknown
}

// Test BuiltinFlattenDeep

func TestFlattenDeep_Simple(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// Should flatten one level
	input := []any{[]any{1.0, 2.0}, []any{3.0, 4.0}}
	result, err := invoke(t, "flatten_deep", env, boxArgs(input)...)
	require.NoError(t, err)
	require.Equal(t, []any{1.0, 2.0, 3.0, 4.0}, result.Any())
}

func TestFlattenDeep_DeeplyNested(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// Should flatten to arbitrary depth
	input := []any{[]any{[]any{[]any{1.0, 2.0}}}, []any{[]any{3.0, 4.0}}}
	result, err := invoke(t, "flatten_deep", env, boxArgs(input)...)
	require.NoError(t, err)
	require.Equal(t, []any{1.0, 2.0, 3.0, 4.0}, result.Any())
}

func TestFlattenDeep_PreservesOrder(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// Should preserve order (depth-first)
	input := []any{1.0, []any{2.0, []any{3.0}}, 4.0}
	result, err := invoke(t, "flatten_deep", env, boxArgs(input)...)
	require.NoError(t, err)
	require.Equal(t, []any{1.0, 2.0, 3.0, 4.0}, result.Any())
}

func TestFlattenDeep_NonListLeaves(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// Non-list values should be preserved
	input := []any{1.0, []any{2.0}, 3.0}
	result, err := invoke(t, "flatten_deep", env, boxArgs(input)...)
	require.NoError(t, err)
	require.Equal(t, []any{1.0, 2.0, 3.0}, result.Any())
}

func TestFlattenDeep_EmptyList(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// Empty list should return empty list
	input := []any{}
	result, err := invoke(t, "flatten_deep", env, boxArgs(input)...)
	require.NoError(t, err)
	require.Equal(t, []any{}, result.Any())
}

func TestFlattenDeep_UnknownInput(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// Unknown (undefined) input should propagate unknown
	result, err := invoke(t, "flatten_deep", env, boxArgs(box.Undefined())...)
	require.NoError(t, err)
	require.Equal(t, box.Undefined(), result) // Undefined represents unknown
}

func TestFlattenDeep_UnknownInNestedList(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// Unknown in nested list should propagate unknown
	input := []any{[]any{[]any{1.0, box.Undefined(), 2.0}}}
	result, err := invoke(t, "flatten_deep", env, boxArgs(input)...)
	require.NoError(t, err)
	require.Equal(t, box.Undefined(), result) // Undefined represents unknown
}

func TestFlattenDeep_ErrorNonList(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// Non-list input should return error
	_, err := invoke(t, "flatten_deep", env, boxArgs("not a list")...)
	require.Error(t, err)
	require.ErrorContains(t, err, "must be a list")
}

func TestFlattenDeep_ErrorWrongArgCount(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// Wrong argument count should return error
	_, err := invoke(t, "flatten_deep", env, boxArgs()...)
	require.Error(t, err)
	require.ErrorContains(t, err, "1 argument")

	_, err = invoke(t, "flatten_deep", env, boxArgs([]any{1.0}, 2)...)
	require.Error(t, err)
	require.ErrorContains(t, err, "1 argument")
}

// Test BuiltinAsList

func TestAsList_ListInput(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// List input should return unchanged
	input := []any{1.0, 2.0, 3.0}
	result, err := invoke(t, "as_list", env, boxArgs(input)...)
	require.NoError(t, err)
	require.Equal(t, input, result.Any())
}

func TestAsList_NonListInput(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// Non-list input should be wrapped
	result, err := invoke(t, "as_list", env, boxArgs(42)...)
	require.NoError(t, err)
	require.Equal(t, []any{42.0}, result.Any())
}

func TestAsList_StringInput(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// String input should be wrapped
	result, err := invoke(t, "as_list", env, boxArgs("hello")...)
	require.NoError(t, err)
	require.Equal(t, []any{"hello"}, result.Any())
}

func TestAsList_MapInput(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// Map input should be wrapped
	input := map[string]any{"key": "value"}
	result, err := invoke(t, "as_list", env, boxArgs(input)...)
	require.NoError(t, err)
	require.Equal(t, []any{input}, result.Any())
}

func TestAsList_EmptyList(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// Empty list should return empty list
	input := []any{}
	result, err := invoke(t, "as_list", env, boxArgs(input)...)
	require.NoError(t, err)
	require.Equal(t, []any{}, result.Any())
}

func TestAsList_UnknownInput(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// Unknown (undefined) input should propagate unknown
	result, err := invoke(t, "as_list", env, boxArgs(box.Undefined())...)
	require.NoError(t, err)
	require.Equal(t, box.Undefined(), result) // Undefined represents unknown
}

func TestAsList_UnknownInList(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// Unknown element in list should propagate unknown
	input := []any{1.0, box.Undefined(), 2.0}
	result, err := invoke(t, "as_list", env, boxArgs(input)...)
	require.NoError(t, err)
	require.Equal(t, box.Undefined(), result) // Undefined represents unknown
}

func TestAsList_ErrorWrongArgCount(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// Wrong argument count should return error
	_, err := invoke(t, "as_list", env, boxArgs()...)
	require.Error(t, err)
	require.ErrorContains(t, err, "1 argument")

	_, err = invoke(t, "as_list", env, boxArgs(1, 2)...)
	require.Error(t, err)
	require.ErrorContains(t, err, "1 argument")
}

// Test BuiltinNormaliseList

func TestNormaliseList_SingleValue(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// Single value should become single-element list
	result, err := invoke(t, "normalise_list", env, boxArgs(42)...)
	require.NoError(t, err)
	require.Equal(t, []any{42.0}, result.Any())
}

func TestNormaliseList_FlatList(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// Flat list should remain unchanged
	input := []any{1.0, 2.0, 3.0}
	result, err := invoke(t, "normalise_list", env, boxArgs(input)...)
	require.NoError(t, err)
	require.Equal(t, input, result.Any())
}

func TestNormaliseList_OneLevelNesting(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// One level of nesting should be flattened
	input := []any{[]any{1, 2.0}, []any{3.0, 4.0}}
	result, err := invoke(t, "normalise_list", env, boxArgs(input)...)
	require.NoError(t, err)
	require.Equal(t, []any{1.0, 2.0, 3.0, 4.0}, result.Any())
}

func TestNormaliseList_MixedOneOrMany(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// Mixed one-or-many should be normalized
	input := []any{1.0, []any{2, 3.0}, 4}
	result, err := invoke(t, "normalise_list", env, boxArgs(input)...)
	require.NoError(t, err)
	require.Equal(t, []any{1.0, 2.0, 3.0, 4.0}, result.Any())
}

func TestNormaliseList_SingleValueThenFlatten(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// Single value wrapped then flattened should work
	result, err := invoke(t, "normalise_list", env, boxArgs(42)...)
	require.NoError(t, err)
	require.Equal(t, []any{42.0}, result.Any())
}

func TestNormaliseList_ErrorDeeperNesting(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// Deeper than one level should return error
	input := []any{[]any{[]any{1, 2.0}}}
	_, err := invoke(t, "normalise_list", env, boxArgs(input)...)
	require.Error(t, err)
	require.ErrorContains(t, err, "deeper than one level")
}

func TestNormaliseList_ErrorDeeperNestingMixed(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// Mixed with deeper nesting should return error
	input := []any{[]any{[]any{1}, 2}}
	_, err := invoke(t, "normalise_list", env, boxArgs(input)...)
	require.Error(t, err)
	require.ErrorContains(t, err, "deeper than one level")
}

func TestNormaliseList_UnknownInput(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// Unknown (undefined) input should propagate unknown
	result, err := invoke(t, "normalise_list", env, boxArgs(box.Undefined())...)
	require.NoError(t, err)
	require.Equal(t, box.Undefined(), result) // Undefined represents unknown
}

func TestNormaliseList_UnknownInNestedList(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// Unknown in nested list should propagate unknown
	input := []any{[]any{1, box.Undefined(), 2.0}}
	result, err := invoke(t, "normalise_list", env, boxArgs(input)...)
	require.NoError(t, err)
	require.Equal(t, box.Undefined(), result) // Undefined represents unknown
}

func TestNormaliseList_ErrorWrongArgCount(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// Wrong argument count should return error
	_, err := invoke(t, "normalise_list", env, boxArgs()...)
	require.Error(t, err)
	require.ErrorContains(t, err, "1 argument")

	_, err = invoke(t, "normalise_list", env, boxArgs(1, 2)...)
	require.Error(t, err)
	require.ErrorContains(t, err, "1 argument")
}

// Integration tests

func TestFlatten_Int64Depth(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// Test that int64 depth values work (common in Go)
	input := []any{[]any{1.0, 2.0}}
	result, err := invoke(t, "flatten", env, boxArgs(input, int64(1))...)
	require.NoError(t, err)
	require.Equal(t, []any{1.0, 2.0}, result.Any())
}

func TestComplexNestedStructures(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// Test with complex nested structures
	input := []any{[]any{1.0, 2.0},
		"string",
		[]any{3.0, []any{4.0, 5.0}},
		6.0,
	}
	result, err := invoke(t, "flatten", env, boxArgs(input)...)
	require.NoError(t, err)
	require.Equal(t, []any{1.0, 2.0, "string", 3.0, []any{4.0, 5.0}, 6.0}, result.Any())
}

func TestNormaliseList_ComplexCase(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// Test normalise_list with complex real-world case
	// T | list<T | list<T>> -> list<T>
	input := []any{1.0, []any{2.0, 3.0},
		[]any{4.0},
	}
	result, err := invoke(t, "normalise_list", env, boxArgs(input)...)
	require.NoError(t, err)
	require.Equal(t, []any{1.0, 2.0, 3.0, 4.0}, result.Any())
}

// Test BuiltinCount

func TestCount_List(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// Count should return length of list
	input := []any{1.0, 2.0, 3.0, 4.0, 5.0}
	result, err := invoke(t, "count", env, boxArgs(input)...)
	require.NoError(t, err)
	require.Equal(t, 5.0, result.Any())
}

func TestCount_EmptyList(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// Count should return 0 for empty list
	input := []any{}
	result, err := invoke(t, "count", env, boxArgs(input)...)
	require.NoError(t, err)
	require.Equal(t, 0.0, result.Any())
}

func TestCount_String(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// Count should return length of string
	result, err := invoke(t, "count", env, boxArgs("hello")...)
	require.NoError(t, err)
	require.Equal(t, 5.0, result.Any())
}

func TestCount_EmptyString(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// Count should return 0 for empty string
	result, err := invoke(t, "count", env, boxArgs("")...)
	require.NoError(t, err)
	require.Equal(t, 0.0, result.Any())
}

func TestCount_Map(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// Count should return number of keys in map
	input := map[string]any{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}
	result, err := invoke(t, "count", env, boxArgs(input)...)
	require.NoError(t, err)
	require.Equal(t, 3.0, result.Any())
}

func TestCount_EmptyMap(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// Count should return 0 for empty map
	input := map[string]any{}
	result, err := invoke(t, "count", env, boxArgs(input)...)
	require.NoError(t, err)
	require.Equal(t, 0.0, result.Any())
}

func TestCount_OtherType(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// Count should return 0 for non-list, non-string, non-map types
	result, err := invoke(t, "count", env, boxArgs(42)...)
	require.NoError(t, err)
	require.True(t, result.IsUndefined())
}

func TestCount_Bool(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// Count should return 0 for bool
	result, err := invoke(t, "count", env, boxArgs(true)...)
	require.NoError(t, err)
	require.True(t, result.IsUndefined())
}

func TestCount_ErrorWrongArgCount(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// Wrong argument count should return error
	_, err := invoke(t, "count", env, boxArgs()...)
	require.Error(t, err)
	require.ErrorContains(t, err, "1 argument")

	_, err = invoke(t, "count", env, boxArgs(1, 2)...)
	require.Error(t, err)
	require.ErrorContains(t, err, "1 argument")
}

// Test BuiltInError

func TestError_SingleArgument(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// Error with single argument should use default format
	result, err := invoke(t, "error", env, boxArgs("test error")...)
	require.True(t, result.IsUndefined())
	require.Error(t, err)
	require.ErrorContains(t, err, "test error")
}

func TestError_FormatString(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// Error with format string should format the message
	result, err := invoke(t, "error", env, boxArgs("error: %s", "test")...)
	require.True(t, result.IsUndefined())
	require.Error(t, err)
	require.ErrorContains(t, err, "error: test")
}

func TestError_MultipleArgs(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// Error with multiple format arguments should format correctly
	result, err := invoke(t, "error", env, boxArgs("%s: %d", "count", 42)...)
	require.True(t, result.IsUndefined())
	require.Error(t, err)
	require.ErrorContains(t, err, "count")
	require.ErrorContains(t, err, "42")
}

func TestError_ErrorWrongArgCount(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// No arguments should return error
	_, err := invoke(t, "error", env, boxArgs()...)
	require.Error(t, err)
	require.ErrorContains(t, err, "at least 1 argument")
}

// Test BuiltinMerge

func TestMerge_Simple(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// Merge should combine two maps
	map1 := map[string]any{
		"a": 1,
		"b": 2,
	}
	map2 := map[string]any{
		"c": 3,
		"d": 4,
	}
	result, err := invoke(t, "merge", env, boxArgs(map1, map2)...)
	require.NoError(t, err)

	merged, ok := result.Any().(map[string]any)
	require.True(t, ok)
	require.Equal(t, 4, len(merged))
	require.Equal(t, 1.0, merged["a"])
	require.Equal(t, 2.0, merged["b"])
	require.Equal(t, 3.0, merged["c"])
	require.Equal(t, 4.0, merged["d"])
}

func TestMerge_Overwrite(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// Merge should overwrite values from map2
	map1 := map[string]any{
		"a": 1,
		"b": 2,
	}
	map2 := map[string]any{
		"b": 20,
		"c": 3,
	}
	result, err := invoke(t, "merge", env, boxArgs(map1, map2)...)
	require.NoError(t, err)

	merged, ok := result.Any().(map[string]any)
	require.True(t, ok)
	require.Equal(t, 3, len(merged))
	require.Equal(t, 1.0, merged["a"])
	require.Equal(t, 20.0, merged["b"]) // overwritten by map2
	require.Equal(t, 3.0, merged["c"])
}

func TestMerge_NestedMaps(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// Merge should recursively merge nested maps
	map1 := map[string]any{
		"nested": map[string]any{
			"a": 1,
			"b": 2,
		},
		"top": "value1",
	}
	map2 := map[string]any{
		"nested": map[string]any{
			"b": 20,
			"c": 3,
		},
		"top": "value2",
	}
	result, err := invoke(t, "merge", env, boxArgs(map1, map2)...)
	require.NoError(t, err)

	merged, ok := result.Any().(map[string]any)
	require.True(t, ok)
	require.Equal(t, "value2", merged["top"]) // overwritten

	nested, ok := merged["nested"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, 3, len(nested))
	require.Equal(t, 1.0, nested["a"])  // from map1
	require.Equal(t, 20.0, nested["b"]) // overwritten by map2
	require.Equal(t, 3.0, nested["c"])  // from map2
}

func TestMerge_DeepNesting(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// Merge should handle deeply nested maps
	map1 := map[string]any{
		"level1": map[string]any{
			"level2": map[string]any{
				"a": 1,
			},
		},
	}
	map2 := map[string]any{
		"level1": map[string]any{
			"level2": map[string]any{
				"b": 2,
			},
		},
	}
	result, err := invoke(t, "merge", env, boxArgs(map1, map2)...)
	require.NoError(t, err)

	merged, ok := result.Any().(map[string]any)
	require.True(t, ok)

	level1, ok := merged["level1"].(map[string]any)
	require.True(t, ok)

	level2, ok := level1["level2"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, 2, len(level2))
	require.Equal(t, 1.0, level2["a"])
	require.Equal(t, 2.0, level2["b"])
}

func TestMerge_NoAliasing(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// Merge should create new maps, not alias the originals
	map1 := map[string]any{
		"nested": map[string]any{
			"a": 1,
		},
	}
	map2 := map[string]any{}

	result, err := invoke(t, "merge", env, boxArgs(map1, map2)...)
	require.NoError(t, err)

	merged, ok := result.Any().(map[string]any)
	require.True(t, ok)

	// Modify the original map
	map1["nested"].(map[string]any)["a"] = 999

	// Result should not be affected (no aliasing)
	nested, ok := merged["nested"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, 1.0, nested["a"]) // original value, not 999
}

func TestMerge_EmptyMaps(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// Merge should handle empty maps
	map1 := map[string]any{}
	map2 := map[string]any{}

	result, err := invoke(t, "merge", env, boxArgs(map1, map2)...)
	require.NoError(t, err)

	merged, ok := result.Any().(map[string]any)
	require.True(t, ok)
	require.Equal(t, 0, len(merged))
}

func TestMerge_FirstEmpty(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// Merge with first map empty should return copy of second
	map1 := map[string]any{}
	map2 := map[string]any{
		"a": 1,
		"b": 2,
	}

	result, err := invoke(t, "merge", env, boxArgs(map1, map2)...)
	require.NoError(t, err)

	merged, ok := result.Any().(map[string]any)
	require.True(t, ok)
	require.Equal(t, 2, len(merged))
	require.Equal(t, 1.0, merged["a"])
	require.Equal(t, 2.0, merged["b"])
}

func TestMerge_SecondEmpty(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// Merge with second map empty should return copy of first
	map1 := map[string]any{
		"a": 1,
		"b": 2,
	}
	map2 := map[string]any{}

	result, err := invoke(t, "merge", env, boxArgs(map1, map2)...)
	require.NoError(t, err)

	merged, ok := result.Any().(map[string]any)
	require.True(t, ok)
	require.Equal(t, 2, len(merged))
	require.Equal(t, 1.0, merged["a"])
	require.Equal(t, 2.0, merged["b"])
}

func TestMerge_ErrorWrongArgCount(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// Wrong argument count should return error
	_, err := invoke(t, "merge", env, boxArgs()...)
	require.Error(t, err)
	require.ErrorContains(t, err, "2 arguments")

	_, err = invoke(t, "merge", env, boxArgs(map[string]any{})...)
	require.Error(t, err)
	require.ErrorContains(t, err, "2 arguments")

	_, err = invoke(t, "merge", env, boxArgs(map[string]any{}, map[string]any{}, map[string]any{})...)
	require.Error(t, err)
	require.ErrorContains(t, err, "2 arguments")
}

func TestMerge_ErrorNonMapFirst(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// First argument not a map should return error
	_, err := invoke(t, "merge", env, boxArgs("not a map", map[string]any{})...)
	require.Error(t, err)
	require.ErrorContains(t, err, "first argument is not a dict")
}

func TestMerge_ErrorNonMapSecond(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	// Second argument not a map should return error
	_, err := invoke(t, "merge", env, boxArgs(map[string]any{}, "not a map")...)
	require.Error(t, err)
	require.ErrorContains(t, err, "second argument is not a dict")
}
