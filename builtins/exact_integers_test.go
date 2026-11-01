// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package builtins

import (
	"testing"

	"github.com/sentrie-sh/sentrie/box"
	"github.com/stretchr/testify/require"
)

// TestDistinct_ExactIntegers covers #166: an integer and an equal float are one
// element, and integers above 2^53 that float64 would merge stay distinct.
func TestDistinct_ExactIntegers(t *testing.T) {
	t.Parallel()
	const big = int64(1)<<53 + 1
	list := box.List([]box.Value{box.Int(5), box.Float(5), box.Int(big), box.Int(big - 1), box.Float(2.5)})
	result, err := invoke(t, "distinct", noopEnv(), list)
	require.NoError(t, err)
	require.Equal(t, []any{int64(5), big, big - 1, 2.5}, result.Any())
}

// TestError_FormatsExactIntegers checks that an exact integer argument works
// with both integer and float verbs.
func TestError_FormatsExactIntegers(t *testing.T) {
	t.Parallel()
	_, err := invoke(t, "error", noopEnv(), box.String("%d|%.2f|%v|%5d"), box.Int(9007199254740993), box.Int(3), box.Int(7), box.Int(42))
	require.Error(t, err)
	require.ErrorContains(t, err, "9007199254740993|3.00|7|   42")
}

// TestNow_ExactIntegers: now() and now_ms() return exact integers, like count().
func TestNow_ExactIntegers(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"now", "now_ms"} {
		result, err := invoke(t, name, noopEnv())
		require.NoError(t, err)
		require.True(t, result.IsInt(), name)
	}
}

func TestFlatten_ExactIntegerDepth(t *testing.T) {
	t.Parallel()
	nested := box.List([]box.Value{box.List([]box.Value{box.List([]box.Value{box.Int(1)})})})
	result, err := invoke(t, "flatten", noopEnv(), nested, box.Int(2))
	require.NoError(t, err)
	require.Equal(t, []any{int64(1)}, result.Any())
}
