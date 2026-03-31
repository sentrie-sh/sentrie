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

package box_test

import (
	"testing"

	"github.com/sentrie-sh/sentrie/box"
	"github.com/stretchr/testify/require"
)

func TestMustNumbers(t *testing.T) {
	t.Run("returns both numbers when operands are numeric", func(t *testing.T) {
		lhs, rhs, err := box.MustNumbers(box.Number(12), box.Number(2.5))
		require.NoError(t, err)
		require.Equal(t, 12.0, lhs)
		require.Equal(t, 2.5, rhs)
	})

	t.Run("returns error when left operand is not a number", func(t *testing.T) {
		lhs, rhs, err := box.MustNumbers(box.String("12"), box.Number(2))
		require.EqualError(t, err, "left operand is not a number")
		require.Zero(t, lhs)
		require.Zero(t, rhs)
	})

	t.Run("returns error when right operand is not a number", func(t *testing.T) {
		lhs, rhs, err := box.MustNumbers(box.Number(2), box.Bool(true))
		require.EqualError(t, err, "right operand is not a number")
		require.Zero(t, lhs)
		require.Zero(t, rhs)
	})
}

func TestEqualValues(t *testing.T) {
	t.Run("supports cross kind numeric equality only for numbers", func(t *testing.T) {
		require.True(t, box.EqualValues(box.Number(42), box.Number(42.0)))
		require.False(t, box.EqualValues(box.Number(42), box.String("42")))
	})

	t.Run("treats undefined and null as equal within each kind", func(t *testing.T) {
		require.True(t, box.EqualValues(box.Undefined(), box.Undefined()))
		require.True(t, box.EqualValues(box.Null(), box.Null()))
		require.False(t, box.EqualValues(box.Undefined(), box.Null()))
	})

	t.Run("compares nested lists recursively", func(t *testing.T) {
		left := box.List([]box.Value{
			box.Number(1),
			box.List([]box.Value{box.String("x"), box.Number(2)}),
			box.Map(map[string]box.Value{"k": box.Bool(true)}),
		})
		rightEqual := box.List([]box.Value{
			box.Number(1.0),
			box.List([]box.Value{box.String("x"), box.Number(2)}),
			box.Map(map[string]box.Value{"k": box.Bool(true)}),
		})
		rightDifferent := box.List([]box.Value{
			box.Number(1),
			box.List([]box.Value{box.String("x"), box.Number(3)}),
			box.Map(map[string]box.Value{"k": box.Bool(true)}),
		})

		require.True(t, box.EqualValues(left, rightEqual))
		require.False(t, box.EqualValues(left, rightDifferent))
	})

	t.Run("compares maps recursively and checks key set", func(t *testing.T) {
		left := box.Map(map[string]box.Value{
			"a": box.Number(7),
			"b": box.Map(map[string]box.Value{
				"nested": box.String("ok"),
			}),
		})
		rightEqual := box.Map(map[string]box.Value{
			"a": box.Number(7.0),
			"b": box.Map(map[string]box.Value{
				"nested": box.String("ok"),
			}),
		})
		rightMissingKey := box.Map(map[string]box.Value{
			"a": box.Number(7),
		})
		rightDifferentValue := box.Map(map[string]box.Value{
			"a": box.Number(7),
			"b": box.Map(map[string]box.Value{
				"nested": box.String("nope"),
			}),
		})

		require.True(t, box.EqualValues(left, rightEqual))
		require.False(t, box.EqualValues(left, rightMissingKey))
		require.False(t, box.EqualValues(left, rightDifferentValue))
	})

	t.Run("uses document reference identity semantics", func(t *testing.T) {
		type doc struct {
			id string
		}
		shared := &doc{id: "same"}
		equalByValue := &doc{id: "same"}

		require.True(t, box.EqualValues(box.Document(shared), box.Document(shared)))
		require.False(t, box.EqualValues(box.Document(shared), box.Document(equalByValue)))
	})
}

func TestMatchesValue(t *testing.T) {
	t.Run("returns error when haystack is not a string", func(t *testing.T) {
		matched, err := box.MatchesValue(box.Number(12), box.String("^12$"))
		require.EqualError(t, err, "haystack must be a string")
		require.False(t, matched)
	})

	t.Run("returns error when pattern is not a string", func(t *testing.T) {
		matched, err := box.MatchesValue(box.String("12"), box.Number(12))
		require.EqualError(t, err, "pattern must be a string")
		require.False(t, matched)
	})

	t.Run("returns true when regex matches", func(t *testing.T) {
		matched, err := box.MatchesValue(box.String("abc123"), box.String("^[a-z]+\\d+$"))
		require.NoError(t, err)
		require.True(t, matched)
	})

	t.Run("returns false when regex does not match", func(t *testing.T) {
		matched, err := box.MatchesValue(box.String("abc"), box.String("^\\d+$"))
		require.NoError(t, err)
		require.False(t, matched)
	})

	t.Run("returns regex compile error for invalid pattern", func(t *testing.T) {
		matched, err := box.MatchesValue(box.String("abc"), box.String("("))
		require.Error(t, err)
		require.False(t, matched)
	})
}

func TestContainsValue(t *testing.T) {
	t.Run("string contains requires non empty string needle", func(t *testing.T) {
		haystack := box.String("hello world")
		require.True(t, box.ContainsValue(haystack, box.String("world")))
		require.False(t, box.ContainsValue(haystack, box.String("")))
		require.False(t, box.ContainsValue(haystack, box.Number(1)))
	})

	t.Run("list uses semantic equality for contains", func(t *testing.T) {
		haystack := box.List([]box.Value{
			box.Number(1),
			box.Map(map[string]box.Value{"x": box.Number(2)}),
		})
		require.True(t, box.ContainsValue(haystack, box.Number(1.0)))
		require.True(t, box.ContainsValue(haystack, box.Map(map[string]box.Value{"x": box.Number(2.0)})))
		require.False(t, box.ContainsValue(haystack, box.Map(map[string]box.Value{"x": box.Number(3)})))
	})

	t.Run("map string needle performs key lookup", func(t *testing.T) {
		haystack := box.Map(map[string]box.Value{
			"k1": box.Number(1),
			"k2": box.String("v2"),
		})
		require.True(t, box.ContainsValue(haystack, box.String("k1")))
		require.False(t, box.ContainsValue(haystack, box.String("missing")))
	})

	t.Run("map needle as map checks subset semantics", func(t *testing.T) {
		haystack := box.Map(map[string]box.Value{
			"a": box.Number(1),
			"b": box.Map(map[string]box.Value{"nested": box.Bool(true)}),
		})
		require.True(t, box.ContainsValue(haystack, box.Map(map[string]box.Value{
			"a": box.Number(1.0),
		})))
		require.True(t, box.ContainsValue(haystack, box.Map(map[string]box.Value{
			"b": box.Map(map[string]box.Value{"nested": box.Bool(true)}),
		})))
		require.False(t, box.ContainsValue(haystack, box.Map(map[string]box.Value{
			"missing": box.Number(1),
		})))
		require.False(t, box.ContainsValue(haystack, box.Map(map[string]box.Value{
			"b": box.Map(map[string]box.Value{"nested": box.Bool(false)}),
		})))
	})

	t.Run("map falls back to value membership for non string non map needles", func(t *testing.T) {
		haystack := box.Map(map[string]box.Value{
			"a": box.Number(1),
			"b": box.List([]box.Value{box.String("x")}),
		})
		require.True(t, box.ContainsValue(haystack, box.Number(1.0)))
		require.True(t, box.ContainsValue(haystack, box.List([]box.Value{box.String("x")})))
		require.False(t, box.ContainsValue(haystack, box.Number(2)))
	})

	t.Run("returns false for unsupported haystack kinds", func(t *testing.T) {
		require.False(t, box.ContainsValue(box.Number(123), box.Number(123)))
		require.False(t, box.ContainsValue(box.Bool(true), box.Bool(true)))
		require.False(t, box.ContainsValue(box.Null(), box.Null()))
	})
}
