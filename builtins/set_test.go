// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package builtins

import (
	"testing"

	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/trinary"
	"github.com/stretchr/testify/require"
)

func strs(xs ...string) box.Value {
	out := make([]box.Value, len(xs))
	for i, x := range xs {
		out[i] = box.String(x)
	}
	return box.List(out)
}

func TestSetBuiltins_Results(t *testing.T) {
	t.Parallel()
	env := noopEnv()

	s, err := invoke(t, "set", env, strs("b", "a", "b"))
	require.NoError(t, err)
	require.True(t, s.IsSet())
	require.Equal(t, []any{"a", "b"}, s.Any())

	cases := []struct {
		name string
		a, b box.Value
		want any
	}{
		{"union", strs("c", "a"), strs("b", "a"), []any{"a", "b", "c"}},
		{"intersection", strs("c", "a", "b"), strs("b", "a", "d"), []any{"a", "b"}},
		{"difference", strs("c", "a", "b"), strs("b"), []any{"a", "c"}},
		{"is_subset", strs("a"), strs("b", "a"), true},
		{"is_subset", strs("a", "z"), strs("b", "a"), false},
		{"is_superset", strs("b", "a"), strs("a"), true},
		{"is_superset", strs("a"), strs("b", "a"), false},
		{"union", s, strs(), []any{"a", "b"}},
	}
	for _, tc := range cases {
		got, err := invoke(t, tc.name, env, tc.a, tc.b)
		require.NoError(t, err, tc.name)
		require.Equal(t, tc.want, got.Any(), tc.name)
		if _, isList := tc.want.([]any); isList {
			require.True(t, got.IsSet(), "%s result must be a set", tc.name)
		}
	}
}

func TestSetBuiltins_Errors(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	cases := []struct {
		name string
		args []box.Value
		want string
	}{
		{"set", nil, "set requires 1 argument"},
		{"set", []box.Value{strs(), strs()}, "set requires 1 argument"},
		{"set", []box.Value{box.String("a")}, "set: argument must be a list"},
		{"set", []box.Value{box.Undefined()}, "set: argument must be a list"},
		{"set", []box.Value{box.List([]box.Value{box.Trinary(trinary.Unknown)})}, "set: a set cannot hold an unknown value"},
		{"union", []box.Value{strs()}, "union requires 2 arguments"},
		{"union", []box.Value{box.Number(1), strs()}, "union: first argument must be a set or list"},
		{"intersection", []box.Value{strs(), box.Null()}, "intersection: second argument must be a set or list"},
		{"difference", []box.Value{box.Undefined(), strs()}, "difference: first argument must be a set or list"},
		{"is_subset", []box.Value{strs(), box.Dict(nil)}, "is_subset: second argument must be a set or list"},
		{"is_superset", []box.Value{strs(), box.List([]box.Value{box.Undefined()})}, "is_superset: second argument: a set cannot hold a value of kind undefined"},
	}
	for _, tc := range cases {
		_, err := invoke(t, tc.name, env, tc.args...)
		require.EqualError(t, err, tc.want, tc.name)
	}
}
