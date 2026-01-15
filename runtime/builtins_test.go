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
	"testing"

	"github.com/stretchr/testify/suite"
)

type BuiltinsTestSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *BuiltinsTestSuite) SetupTest() {
	s.ctx = context.Background()
}

func TestBuiltinsTestSuite(t *testing.T) {
	suite.Run(t, new(BuiltinsTestSuite))
}

// Test BuiltinFlatten

func (s *BuiltinsTestSuite) TestFlatten_DefaultDepth() {
	// flatten(x) should flatten exactly one level
	input := []any{[]any{1, 2}, []any{3, 4}}
	result, err := BuiltinFlatten(s.ctx, []any{input})
	s.NoError(err)
	s.Equal([]any{1, 2, 3, 4}, result)
}

func (s *BuiltinsTestSuite) TestFlatten_ExplicitDepth1() {
	// flatten(x, 1) should be equivalent to flatten(x)
	input := []any{[]any{1, 2}, []any{3, 4}}
	result, err := BuiltinFlatten(s.ctx, []any{input, 1})
	s.NoError(err)
	s.Equal([]any{1, 2, 3, 4}, result)
}

func (s *BuiltinsTestSuite) TestFlatten_Depth0() {
	// flatten(x, 0) should return x unchanged
	input := []any{[]any{1, 2}, []any{3, 4}}
	result, err := BuiltinFlatten(s.ctx, []any{input, 0})
	s.NoError(err)
	s.Equal(input, result)
}

func (s *BuiltinsTestSuite) TestFlatten_Depth2() {
	// flatten(x, 2) should flatten two levels
	input := []any{[]any{[]any{1, 2}}, []any{[]any{3, 4}}}
	result, err := BuiltinFlatten(s.ctx, []any{input, 2})
	s.NoError(err)
	s.Equal([]any{1, 2, 3, 4}, result)
}

func (s *BuiltinsTestSuite) TestFlatten_PreservesOrder() {
	// Flattening should preserve order
	input := []any{[]any{1, 2}, 5, []any{3, 4}}
	result, err := BuiltinFlatten(s.ctx, []any{input})
	s.NoError(err)
	s.Equal([]any{1, 2, 5, 3, 4}, result)
}

func (s *BuiltinsTestSuite) TestFlatten_NonListLeaves() {
	// Non-list values should be treated as leaves
	input := []any{1, []any{2, 3}, 4}
	result, err := BuiltinFlatten(s.ctx, []any{input})
	s.NoError(err)
	s.Equal([]any{1, 2, 3, 4}, result)
}

func (s *BuiltinsTestSuite) TestFlatten_EmptyList() {
	// Empty list should return empty list
	input := []any{}
	result, err := BuiltinFlatten(s.ctx, []any{input})
	s.NoError(err)
	s.Equal([]any{}, result)
}

func (s *BuiltinsTestSuite) TestFlatten_UnknownInput() {
	// Unknown (undefined) input should propagate unknown
	result, err := BuiltinFlatten(s.ctx, []any{Undefined})
	s.NoError(err)
	s.Equal(Undefined, result) // Undefined represents unknown
}

func (s *BuiltinsTestSuite) TestFlatten_UnknownInNestedList() {
	// Unknown in nested list should propagate unknown
	input := []any{[]any{1, Undefined, 2}}
	result, err := BuiltinFlatten(s.ctx, []any{input})
	s.NoError(err)
	s.Equal(Undefined, result) // Undefined represents unknown
}

func (s *BuiltinsTestSuite) TestFlatten_ErrorNonList() {
	// Non-list input should return error
	_, err := BuiltinFlatten(s.ctx, []any{"not a list"})
	s.Error(err)
	s.Contains(err.Error(), "must be a list")
}

func (s *BuiltinsTestSuite) TestFlatten_ErrorInvalidDepth() {
	// Negative depth should return error
	input := []any{[]any{1, 2}}
	_, err := BuiltinFlatten(s.ctx, []any{input, -1})
	s.Error(err)
	s.Contains(err.Error(), "non-negative integer")
}

func (s *BuiltinsTestSuite) TestFlatten_ErrorInvalidDepthType() {
	// Non-integer depth should return error
	input := []any{[]any{1, 2}}
	_, err := BuiltinFlatten(s.ctx, []any{input, "not an int"})
	s.Error(err)
	s.Contains(err.Error(), "non-negative integer")
}

func (s *BuiltinsTestSuite) TestFlatten_ErrorWrongArgCount() {
	// Wrong argument count should return error
	_, err := BuiltinFlatten(s.ctx, []any{})
	s.Error(err)
	s.Contains(err.Error(), "1 or 2 arguments")

	_, err = BuiltinFlatten(s.ctx, []any{[]any{1}, 1, 2})
	s.Error(err)
	s.Contains(err.Error(), "1 or 2 arguments")
}

func (s *BuiltinsTestSuite) TestFlatten_UnknownDepth() {
	// Unknown depth should propagate unknown
	input := []any{[]any{1, 2}}
	result, err := BuiltinFlatten(s.ctx, []any{input, Undefined})
	s.NoError(err)
	s.Equal(Undefined, result) // Undefined represents unknown
}

// Test BuiltinFlattenDeep

func (s *BuiltinsTestSuite) TestFlattenDeep_Simple() {
	// Should flatten one level
	input := []any{[]any{1, 2}, []any{3, 4}}
	result, err := BuiltinFlattenDeep(s.ctx, []any{input})
	s.NoError(err)
	s.Equal([]any{1, 2, 3, 4}, result)
}

func (s *BuiltinsTestSuite) TestFlattenDeep_DeeplyNested() {
	// Should flatten to arbitrary depth
	input := []any{[]any{[]any{[]any{1, 2}}}, []any{[]any{3, 4}}}
	result, err := BuiltinFlattenDeep(s.ctx, []any{input})
	s.NoError(err)
	s.Equal([]any{1, 2, 3, 4}, result)
}

func (s *BuiltinsTestSuite) TestFlattenDeep_PreservesOrder() {
	// Should preserve order (depth-first)
	input := []any{1, []any{2, []any{3}}, 4}
	result, err := BuiltinFlattenDeep(s.ctx, []any{input})
	s.NoError(err)
	s.Equal([]any{1, 2, 3, 4}, result)
}

func (s *BuiltinsTestSuite) TestFlattenDeep_NonListLeaves() {
	// Non-list values should be preserved
	input := []any{1, []any{2}, 3}
	result, err := BuiltinFlattenDeep(s.ctx, []any{input})
	s.NoError(err)
	s.Equal([]any{1, 2, 3}, result)
}

func (s *BuiltinsTestSuite) TestFlattenDeep_EmptyList() {
	// Empty list should return empty list
	input := []any{}
	result, err := BuiltinFlattenDeep(s.ctx, []any{input})
	s.NoError(err)
	s.Equal([]any{}, result)
}

func (s *BuiltinsTestSuite) TestFlattenDeep_UnknownInput() {
	// Unknown (undefined) input should propagate unknown
	result, err := BuiltinFlattenDeep(s.ctx, []any{Undefined})
	s.NoError(err)
	s.Equal(Undefined, result) // Undefined represents unknown
}

func (s *BuiltinsTestSuite) TestFlattenDeep_UnknownInNestedList() {
	// Unknown in nested list should propagate unknown
	input := []any{[]any{[]any{1, Undefined, 2}}}
	result, err := BuiltinFlattenDeep(s.ctx, []any{input})
	s.NoError(err)
	s.Equal(Undefined, result) // Undefined represents unknown
}

func (s *BuiltinsTestSuite) TestFlattenDeep_ErrorNonList() {
	// Non-list input should return error
	_, err := BuiltinFlattenDeep(s.ctx, []any{"not a list"})
	s.Error(err)
	s.Contains(err.Error(), "must be a list")
}

func (s *BuiltinsTestSuite) TestFlattenDeep_ErrorWrongArgCount() {
	// Wrong argument count should return error
	_, err := BuiltinFlattenDeep(s.ctx, []any{})
	s.Error(err)
	s.Contains(err.Error(), "1 argument")

	_, err = BuiltinFlattenDeep(s.ctx, []any{[]any{1}, 2})
	s.Error(err)
	s.Contains(err.Error(), "1 argument")
}

// Test BuiltinAsList

func (s *BuiltinsTestSuite) TestAsList_ListInput() {
	// List input should return unchanged
	input := []any{1, 2, 3}
	result, err := BuiltinAsList(s.ctx, []any{input})
	s.NoError(err)
	s.Equal(input, result)
}

func (s *BuiltinsTestSuite) TestAsList_NonListInput() {
	// Non-list input should be wrapped
	result, err := BuiltinAsList(s.ctx, []any{42})
	s.NoError(err)
	s.Equal([]any{42}, result)
}

func (s *BuiltinsTestSuite) TestAsList_StringInput() {
	// String input should be wrapped
	result, err := BuiltinAsList(s.ctx, []any{"hello"})
	s.NoError(err)
	s.Equal([]any{"hello"}, result)
}

func (s *BuiltinsTestSuite) TestAsList_MapInput() {
	// Map input should be wrapped
	input := map[string]any{"key": "value"}
	result, err := BuiltinAsList(s.ctx, []any{input})
	s.NoError(err)
	s.Equal([]any{input}, result)
}

func (s *BuiltinsTestSuite) TestAsList_EmptyList() {
	// Empty list should return empty list
	input := []any{}
	result, err := BuiltinAsList(s.ctx, []any{input})
	s.NoError(err)
	s.Equal([]any{}, result)
}

func (s *BuiltinsTestSuite) TestAsList_UnknownInput() {
	// Unknown (undefined) input should propagate unknown
	result, err := BuiltinAsList(s.ctx, []any{Undefined})
	s.NoError(err)
	s.Equal(Undefined, result) // Undefined represents unknown
}

func (s *BuiltinsTestSuite) TestAsList_UnknownInList() {
	// Unknown element in list should propagate unknown
	input := []any{1, Undefined, 2}
	result, err := BuiltinAsList(s.ctx, []any{input})
	s.NoError(err)
	s.Equal(Undefined, result) // Undefined represents unknown
}

func (s *BuiltinsTestSuite) TestAsList_ErrorWrongArgCount() {
	// Wrong argument count should return error
	_, err := BuiltinAsList(s.ctx, []any{})
	s.Error(err)
	s.Contains(err.Error(), "1 argument")

	_, err = BuiltinAsList(s.ctx, []any{1, 2})
	s.Error(err)
	s.Contains(err.Error(), "1 argument")
}

// Test BuiltinNormaliseList

func (s *BuiltinsTestSuite) TestNormaliseList_SingleValue() {
	// Single value should become single-element list
	result, err := BuiltinNormaliseList(s.ctx, []any{42})
	s.NoError(err)
	s.Equal([]any{42}, result)
}

func (s *BuiltinsTestSuite) TestNormaliseList_FlatList() {
	// Flat list should remain unchanged
	input := []any{1, 2, 3}
	result, err := BuiltinNormaliseList(s.ctx, []any{input})
	s.NoError(err)
	s.Equal(input, result)
}

func (s *BuiltinsTestSuite) TestNormaliseList_OneLevelNesting() {
	// One level of nesting should be flattened
	input := []any{[]any{1, 2}, []any{3, 4}}
	result, err := BuiltinNormaliseList(s.ctx, []any{input})
	s.NoError(err)
	s.Equal([]any{1, 2, 3, 4}, result)
}

func (s *BuiltinsTestSuite) TestNormaliseList_MixedOneOrMany() {
	// Mixed one-or-many should be normalized
	input := []any{1, []any{2, 3}, 4}
	result, err := BuiltinNormaliseList(s.ctx, []any{input})
	s.NoError(err)
	s.Equal([]any{1, 2, 3, 4}, result)
}

func (s *BuiltinsTestSuite) TestNormaliseList_SingleValueThenFlatten() {
	// Single value wrapped then flattened should work
	result, err := BuiltinNormaliseList(s.ctx, []any{42})
	s.NoError(err)
	s.Equal([]any{42}, result)
}

func (s *BuiltinsTestSuite) TestNormaliseList_ErrorDeeperNesting() {
	// Deeper than one level should return error
	input := []any{[]any{[]any{1, 2}}}
	_, err := BuiltinNormaliseList(s.ctx, []any{input})
	s.Error(err)
	s.Contains(err.Error(), "deeper than one level")
}

func (s *BuiltinsTestSuite) TestNormaliseList_ErrorDeeperNestingMixed() {
	// Mixed with deeper nesting should return error
	input := []any{[]any{[]any{1}, 2}}
	_, err := BuiltinNormaliseList(s.ctx, []any{input})
	s.Error(err)
	s.Contains(err.Error(), "deeper than one level")
}

func (s *BuiltinsTestSuite) TestNormaliseList_UnknownInput() {
	// Unknown (undefined) input should propagate unknown
	result, err := BuiltinNormaliseList(s.ctx, []any{Undefined})
	s.NoError(err)
	s.Equal(Undefined, result) // Undefined represents unknown
}

func (s *BuiltinsTestSuite) TestNormaliseList_UnknownInNestedList() {
	// Unknown in nested list should propagate unknown
	input := []any{[]any{1, Undefined, 2}}
	result, err := BuiltinNormaliseList(s.ctx, []any{input})
	s.NoError(err)
	s.Equal(Undefined, result) // Undefined represents unknown
}

func (s *BuiltinsTestSuite) TestNormaliseList_ErrorWrongArgCount() {
	// Wrong argument count should return error
	_, err := BuiltinNormaliseList(s.ctx, []any{})
	s.Error(err)
	s.Contains(err.Error(), "1 argument")

	_, err = BuiltinNormaliseList(s.ctx, []any{1, 2})
	s.Error(err)
	s.Contains(err.Error(), "1 argument")
}

// Integration tests

func (s *BuiltinsTestSuite) TestFlatten_Int64Depth() {
	// Test that int64 depth values work (common in Go)
	input := []any{[]any{1, 2}}
	result, err := BuiltinFlatten(s.ctx, []any{input, int64(1)})
	s.NoError(err)
	s.Equal([]any{1, 2}, result)
}

func (s *BuiltinsTestSuite) TestComplexNestedStructures() {
	// Test with complex nested structures
	input := []any{
		[]any{1, 2},
		"string",
		[]any{3, []any{4, 5}},
		6,
	}
	result, err := BuiltinFlatten(s.ctx, []any{input})
	s.NoError(err)
	s.Equal([]any{1, 2, "string", 3, []any{4, 5}, 6}, result)
}

func (s *BuiltinsTestSuite) TestNormaliseList_ComplexCase() {
	// Test normalise_list with complex real-world case
	// T | list<T | list<T>> -> list<T>
	input := []any{
		1,
		[]any{2, 3},
		[]any{4},
	}
	result, err := BuiltinNormaliseList(s.ctx, []any{input})
	s.NoError(err)
	s.Equal([]any{1, 2, 3, 4}, result)
}

// Test BuiltinCount

func (s *BuiltinsTestSuite) TestCount_List() {
	// Count should return length of list
	input := []any{1, 2, 3, 4, 5}
	result, err := BuiltinCount(s.ctx, []any{input})
	s.NoError(err)
	s.Equal(5, result)
}

func (s *BuiltinsTestSuite) TestCount_EmptyList() {
	// Count should return 0 for empty list
	input := []any{}
	result, err := BuiltinCount(s.ctx, []any{input})
	s.NoError(err)
	s.Equal(0, result)
}

func (s *BuiltinsTestSuite) TestCount_String() {
	// Count should return length of string
	result, err := BuiltinCount(s.ctx, []any{"hello"})
	s.NoError(err)
	s.Equal(5, result)
}

func (s *BuiltinsTestSuite) TestCount_EmptyString() {
	// Count should return 0 for empty string
	result, err := BuiltinCount(s.ctx, []any{""})
	s.NoError(err)
	s.Equal(0, result)
}

func (s *BuiltinsTestSuite) TestCount_Map() {
	// Count should return number of keys in map
	input := map[string]any{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}
	result, err := BuiltinCount(s.ctx, []any{input})
	s.NoError(err)
	s.Equal(3, result)
}

func (s *BuiltinsTestSuite) TestCount_EmptyMap() {
	// Count should return 0 for empty map
	input := map[string]any{}
	result, err := BuiltinCount(s.ctx, []any{input})
	s.NoError(err)
	s.Equal(0, result)
}

func (s *BuiltinsTestSuite) TestCount_OtherType() {
	// Count should return 0 for non-list, non-string, non-map types
	result, err := BuiltinCount(s.ctx, []any{42})
	s.NoError(err)
	s.Equal(0, result)
}

func (s *BuiltinsTestSuite) TestCount_Bool() {
	// Count should return 0 for bool
	result, err := BuiltinCount(s.ctx, []any{true})
	s.NoError(err)
	s.Equal(0, result)
}

func (s *BuiltinsTestSuite) TestCount_ErrorWrongArgCount() {
	// Wrong argument count should return error
	_, err := BuiltinCount(s.ctx, []any{})
	s.Error(err)
	s.Contains(err.Error(), "1 argument")

	_, err = BuiltinCount(s.ctx, []any{1, 2})
	s.Error(err)
	s.Contains(err.Error(), "1 argument")
}

// Test BuiltInError

func (s *BuiltinsTestSuite) TestError_SingleArgument() {
	// Error with single argument should use default format
	result, err := BuiltInError(s.ctx, []any{"test error"})
	s.Nil(result)
	s.Error(err)
	s.Contains(err.Error(), "test error")
}

func (s *BuiltinsTestSuite) TestError_FormatString() {
	// Error with format string should format the message
	result, err := BuiltInError(s.ctx, []any{"error: %s", "test"})
	s.Nil(result)
	s.Error(err)
	s.Contains(err.Error(), "error: test")
}

func (s *BuiltinsTestSuite) TestError_MultipleArgs() {
	// Error with multiple format arguments should format correctly
	result, err := BuiltInError(s.ctx, []any{"%s: %d", "count", 42})
	s.Nil(result)
	s.Error(err)
	s.Contains(err.Error(), "count")
	s.Contains(err.Error(), "42")
}

func (s *BuiltinsTestSuite) TestError_ErrorWrongArgCount() {
	// No arguments should return error
	_, err := BuiltInError(s.ctx, []any{})
	s.Error(err)
	s.Contains(err.Error(), "at least 1 argument")
}

// Test BuiltinMerge

func (s *BuiltinsTestSuite) TestMerge_Simple() {
	// Merge should combine two maps
	map1 := map[string]any{
		"a": 1,
		"b": 2,
	}
	map2 := map[string]any{
		"c": 3,
		"d": 4,
	}
	result, err := BuiltinMerge(s.ctx, []any{map1, map2})
	s.NoError(err)

	merged, ok := result.(map[string]any)
	s.True(ok)
	s.Equal(4, len(merged))
	s.Equal(1, merged["a"])
	s.Equal(2, merged["b"])
	s.Equal(3, merged["c"])
	s.Equal(4, merged["d"])
}

func (s *BuiltinsTestSuite) TestMerge_Overwrite() {
	// Merge should overwrite values from map2
	map1 := map[string]any{
		"a": 1,
		"b": 2,
	}
	map2 := map[string]any{
		"b": 20,
		"c": 3,
	}
	result, err := BuiltinMerge(s.ctx, []any{map1, map2})
	s.NoError(err)

	merged, ok := result.(map[string]any)
	s.True(ok)
	s.Equal(3, len(merged))
	s.Equal(1, merged["a"])
	s.Equal(20, merged["b"]) // overwritten by map2
	s.Equal(3, merged["c"])
}

func (s *BuiltinsTestSuite) TestMerge_NestedMaps() {
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
	result, err := BuiltinMerge(s.ctx, []any{map1, map2})
	s.NoError(err)

	merged, ok := result.(map[string]any)
	s.True(ok)
	s.Equal("value2", merged["top"]) // overwritten

	nested, ok := merged["nested"].(map[string]any)
	s.True(ok)
	s.Equal(3, len(nested))
	s.Equal(1, nested["a"])  // from map1
	s.Equal(20, nested["b"]) // overwritten by map2
	s.Equal(3, nested["c"])  // from map2
}

func (s *BuiltinsTestSuite) TestMerge_DeepNesting() {
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
	result, err := BuiltinMerge(s.ctx, []any{map1, map2})
	s.NoError(err)

	merged, ok := result.(map[string]any)
	s.True(ok)

	level1, ok := merged["level1"].(map[string]any)
	s.True(ok)

	level2, ok := level1["level2"].(map[string]any)
	s.True(ok)
	s.Equal(2, len(level2))
	s.Equal(1, level2["a"])
	s.Equal(2, level2["b"])
}

func (s *BuiltinsTestSuite) TestMerge_NoAliasing() {
	// Merge should create new maps, not alias the originals
	map1 := map[string]any{
		"nested": map[string]any{
			"a": 1,
		},
	}
	map2 := map[string]any{}

	result, err := BuiltinMerge(s.ctx, []any{map1, map2})
	s.NoError(err)

	merged, ok := result.(map[string]any)
	s.True(ok)

	// Modify the original map
	map1["nested"].(map[string]any)["a"] = 999

	// Result should not be affected (no aliasing)
	nested, ok := merged["nested"].(map[string]any)
	s.True(ok)
	s.Equal(1, nested["a"]) // original value, not 999
}

func (s *BuiltinsTestSuite) TestMerge_EmptyMaps() {
	// Merge should handle empty maps
	map1 := map[string]any{}
	map2 := map[string]any{}

	result, err := BuiltinMerge(s.ctx, []any{map1, map2})
	s.NoError(err)

	merged, ok := result.(map[string]any)
	s.True(ok)
	s.Equal(0, len(merged))
}

func (s *BuiltinsTestSuite) TestMerge_FirstEmpty() {
	// Merge with first map empty should return copy of second
	map1 := map[string]any{}
	map2 := map[string]any{
		"a": 1,
		"b": 2,
	}

	result, err := BuiltinMerge(s.ctx, []any{map1, map2})
	s.NoError(err)

	merged, ok := result.(map[string]any)
	s.True(ok)
	s.Equal(2, len(merged))
	s.Equal(1, merged["a"])
	s.Equal(2, merged["b"])
}

func (s *BuiltinsTestSuite) TestMerge_SecondEmpty() {
	// Merge with second map empty should return copy of first
	map1 := map[string]any{
		"a": 1,
		"b": 2,
	}
	map2 := map[string]any{}

	result, err := BuiltinMerge(s.ctx, []any{map1, map2})
	s.NoError(err)

	merged, ok := result.(map[string]any)
	s.True(ok)
	s.Equal(2, len(merged))
	s.Equal(1, merged["a"])
	s.Equal(2, merged["b"])
}

func (s *BuiltinsTestSuite) TestMerge_ErrorWrongArgCount() {
	// Wrong argument count should return error
	_, err := BuiltinMerge(s.ctx, []any{})
	s.Error(err)
	s.Contains(err.Error(), "2 arguments")

	_, err = BuiltinMerge(s.ctx, []any{map[string]any{}})
	s.Error(err)
	s.Contains(err.Error(), "2 arguments")

	_, err = BuiltinMerge(s.ctx, []any{map[string]any{}, map[string]any{}, map[string]any{}})
	s.Error(err)
	s.Contains(err.Error(), "2 arguments")
}

func (s *BuiltinsTestSuite) TestMerge_ErrorNonMapFirst() {
	// First argument not a map should return error
	_, err := BuiltinMerge(s.ctx, []any{"not a map", map[string]any{}})
	s.Error(err)
	s.Contains(err.Error(), "first argument is not a map")
}

func (s *BuiltinsTestSuite) TestMerge_ErrorNonMapSecond() {
	// Second argument not a map should return error
	_, err := BuiltinMerge(s.ctx, []any{map[string]any{}, "not a map"})
	s.Error(err)
	s.Contains(err.Error(), "second argument is not a map")
}
