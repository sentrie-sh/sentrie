// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package builtins

import (
	"testing"

	"github.com/sentrie-sh/sentrie/box"
	"github.com/stretchr/testify/require"
)

func objArg(m map[string]any) box.Value { return box.FromAny(m) }

func TestObjectGet(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	doc := objArg(map[string]any{
		"a":  map[string]any{"b": map[string]any{"c": 1.0}},
		"xs": []any{"zero", map[string]any{"k": "v"}},
		"n":  nil,
		"s":  "str",
	})
	def := box.String("DEFAULT")

	cases := []struct {
		name string
		obj  box.Value
		path box.Value
		want any
	}{
		{"nested list path", doc, box.FromAny([]any{"a", "b", "c"}), 1.0},
		{"single string key", doc, box.String("s"), "str"},
		{"dotted string is one key", doc, box.String("a.b.c"), "DEFAULT"},
		{"list index", doc, box.FromAny([]any{"xs", 0.0}), "zero"},
		{"index then key", doc, box.FromAny([]any{"xs", 1.0, "k"}), "v"},
		{"missing key", doc, box.FromAny([]any{"a", "x"}), "DEFAULT"},
		{"missing deep key", doc, box.FromAny([]any{"a", "b", "x", "y"}), "DEFAULT"},
		{"index out of range", doc, box.FromAny([]any{"xs", 5.0}), "DEFAULT"},
		{"negative index", doc, box.FromAny([]any{"xs", -1.0}), "DEFAULT"},
		{"fractional index", doc, box.FromAny([]any{"xs", 0.5}), "DEFAULT"},
		{"number segment on dict", doc, box.FromAny([]any{"a", 0.0}), "DEFAULT"},
		{"string segment on list", doc, box.FromAny([]any{"xs", "0"}), "DEFAULT"},
		{"through a scalar", doc, box.FromAny([]any{"s", "x"}), "DEFAULT"},
		{"present null is returned", doc, box.String("n"), nil},
		{"through a null", doc, box.FromAny([]any{"n", "x"}), "DEFAULT"},
		{"empty path is the object", objArg(map[string]any{"a": 1.0}), box.FromAny([]any{}), map[string]any{"a": 1.0}},
		{"undefined object", box.Undefined(), box.String("a"), "DEFAULT"},
		{"null object", box.Null(), box.String("a"), "DEFAULT"},
		{"null object empty path", box.Null(), box.FromAny([]any{}), "DEFAULT"},
		{"undefined object empty path", box.Undefined(), box.FromAny([]any{}), "DEFAULT"},
		{"single number segment on a list", box.FromAny([]any{"p", "q"}), box.Number(1), "q"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			v, err := invoke(t, "object_get", env, tc.obj, tc.path, def)
			require.NoError(t, err)
			require.Equal(t, tc.want, v.Any())
		})
	}

	t.Run("default may be any kind", func(t *testing.T) {
		t.Parallel()
		v, err := invoke(t, "object_get", env, doc, box.String("zz"), box.Number(7))
		require.NoError(t, err)
		require.Equal(t, 7.0, v.Any())
	})
	t.Run("undefined path propagates", func(t *testing.T) {
		t.Parallel()
		v, err := invoke(t, "object_get", env, doc, box.Undefined(), def)
		require.NoError(t, err)
		require.True(t, v.IsUndefined())
	})
	t.Run("null path errors", func(t *testing.T) {
		t.Parallel()
		_, err := invoke(t, "object_get", env, doc, box.Null(), def)
		require.ErrorContains(t, err, "object_get: path must be")
	})
	t.Run("bool path errors", func(t *testing.T) {
		t.Parallel()
		_, err := invoke(t, "object_get", env, doc, box.Bool(true), def)
		require.EqualError(t, err, "object_get: path must be a string, a number or a list of segments")
	})
	t.Run("bad segment errors", func(t *testing.T) {
		t.Parallel()
		_, err := invoke(t, "object_get", env, doc, box.FromAny([]any{"a", true}), def)
		require.ErrorContains(t, err, "path segments must be strings or non-negative integers")
	})
	t.Run("arity", func(t *testing.T) {
		t.Parallel()
		_, err := invoke(t, "object_get", env, doc, box.String("a"))
		require.EqualError(t, err, "object_get requires 3 arguments")
		_, err = invoke(t, "object_get", env, doc, box.String("a"), def, def)
		require.EqualError(t, err, "object_get requires 3 arguments")
	})
}

func TestObjectKeys(t *testing.T) {
	t.Parallel()
	env := noopEnv()

	t.Run("sorted and deterministic", func(t *testing.T) {
		t.Parallel()
		m := objArg(map[string]any{"b": 1.0, "c": 2.0, "a": 3.0, "B": 4.0, "": 5.0})
		for range 20 {
			v, err := invoke(t, "object_keys", env, m)
			require.NoError(t, err)
			require.Equal(t, []any{"", "B", "a", "b", "c"}, v.Any())
		}
	})
	t.Run("empty dict", func(t *testing.T) {
		t.Parallel()
		v, err := invoke(t, "object_keys", env, objArg(map[string]any{}))
		require.NoError(t, err)
		require.Equal(t, []any{}, v.Any())
	})
	t.Run("undefined propagates", func(t *testing.T) {
		t.Parallel()
		v, err := invoke(t, "object_keys", env, box.Undefined())
		require.NoError(t, err)
		require.True(t, v.IsUndefined())
	})
	t.Run("wrong kind", func(t *testing.T) {
		t.Parallel()
		for _, bad := range []box.Value{box.String("x"), box.Null(), box.FromAny([]any{1.0}), box.Number(1)} {
			_, err := invoke(t, "object_keys", env, bad)
			require.EqualError(t, err, "object_keys: argument must be a dict")
		}
	})
	t.Run("arity", func(t *testing.T) {
		t.Parallel()
		_, err := invoke(t, "object_keys", env)
		require.EqualError(t, err, "object_keys requires 1 argument")
		_, err = invoke(t, "object_keys", env, objArg(nil), objArg(nil))
		require.EqualError(t, err, "object_keys requires 1 argument")
	})
}

func TestObjectRemoveAndFilter(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	src := map[string]any{"a": 1.0, "b": map[string]any{"x": 1.0}, "c": 3.0}
	keys := box.FromAny([]any{"a", "b", "zzz"})

	t.Run("remove", func(t *testing.T) {
		t.Parallel()
		v, err := invoke(t, "object_remove", env, objArg(src), keys)
		require.NoError(t, err)
		require.Equal(t, map[string]any{"c": 3.0}, v.Any())
	})
	t.Run("filter", func(t *testing.T) {
		t.Parallel()
		v, err := invoke(t, "object_filter", env, objArg(src), keys)
		require.NoError(t, err)
		require.Equal(t, map[string]any{"a": 1.0, "b": map[string]any{"x": 1.0}}, v.Any())
	})
	t.Run("empty key list", func(t *testing.T) {
		t.Parallel()
		v, err := invoke(t, "object_remove", env, objArg(src), box.FromAny([]any{}))
		require.NoError(t, err)
		require.Equal(t, src, v.Any())
		v, err = invoke(t, "object_filter", env, objArg(src), box.FromAny([]any{}))
		require.NoError(t, err)
		require.Equal(t, map[string]any{}, v.Any())
	})
	t.Run("input is not mutated", func(t *testing.T) {
		t.Parallel()
		in := objArg(map[string]any{"a": 1.0, "b": map[string]any{"x": 1.0}})
		_, err := invoke(t, "object_remove", env, in, box.FromAny([]any{"a"}))
		require.NoError(t, err)
		_, err = invoke(t, "object_filter", env, in, box.FromAny([]any{"a"}))
		require.NoError(t, err)
		require.Equal(t, map[string]any{"a": 1.0, "b": map[string]any{"x": 1.0}}, in.Any())
	})
	for _, name := range []string{"object_remove", "object_filter"} {
		t.Run(name+" errors", func(t *testing.T) {
			t.Parallel()
			_, err := invoke(t, name, env, box.String("x"), keys)
			require.EqualError(t, err, name+": first argument must be a dict")
			_, err = invoke(t, name, env, box.Null(), keys)
			require.EqualError(t, err, name+": first argument must be a dict")
			_, err = invoke(t, name, env, objArg(src), box.String("a"))
			require.EqualError(t, err, name+": keys must be a list of strings")
			_, err = invoke(t, name, env, objArg(src), box.FromAny([]any{"a", 1.0}))
			require.EqualError(t, err, name+": keys must be a list of strings")
			_, err = invoke(t, name, env, objArg(src))
			require.EqualError(t, err, name+" requires 2 arguments")
		})
		t.Run(name+" undefined propagates", func(t *testing.T) {
			t.Parallel()
			v, err := invoke(t, name, env, box.Undefined(), keys)
			require.NoError(t, err)
			require.True(t, v.IsUndefined())
			v, err = invoke(t, name, env, objArg(src), box.Undefined())
			require.NoError(t, err)
			require.True(t, v.IsUndefined())
		})
	}
}

func TestObjectUnion(t *testing.T) {
	t.Parallel()
	env := noopEnv()

	t.Run("deep merge second wins", func(t *testing.T) {
		t.Parallel()
		a := objArg(map[string]any{"x": 1.0, "n": map[string]any{"p": 1.0, "q": 2.0}, "l": []any{1.0, 2.0}, "only_a": true})
		b := objArg(map[string]any{"x": 2.0, "n": map[string]any{"q": 3.0, "r": 4.0}, "l": []any{9.0}, "only_b": "b"})
		v, err := invoke(t, "object_union", env, a, b)
		require.NoError(t, err)
		require.Equal(t, map[string]any{
			"x":      2.0,
			"n":      map[string]any{"p": 1.0, "q": 3.0, "r": 4.0},
			"l":      []any{9.0},
			"only_a": true,
			"only_b": "b",
		}, v.Any())
	})
	t.Run("scalar replaces dict and back", func(t *testing.T) {
		t.Parallel()
		v, err := invoke(t, "object_union", env, objArg(map[string]any{"k": map[string]any{"a": 1.0}}), objArg(map[string]any{"k": 5.0}))
		require.NoError(t, err)
		require.Equal(t, map[string]any{"k": 5.0}, v.Any())
		v, err = invoke(t, "object_union", env, objArg(map[string]any{"k": 5.0}), objArg(map[string]any{"k": map[string]any{"a": 1.0}}))
		require.NoError(t, err)
		require.Equal(t, map[string]any{"k": map[string]any{"a": 1.0}}, v.Any())
	})
	t.Run("inputs not mutated", func(t *testing.T) {
		t.Parallel()
		a := objArg(map[string]any{"n": map[string]any{"p": 1.0}})
		b := objArg(map[string]any{"n": map[string]any{"q": 2.0}})
		_, err := invoke(t, "object_union", env, a, b)
		require.NoError(t, err)
		require.Equal(t, map[string]any{"n": map[string]any{"p": 1.0}}, a.Any())
		require.Equal(t, map[string]any{"n": map[string]any{"q": 2.0}}, b.Any())
	})
	t.Run("errors", func(t *testing.T) {
		t.Parallel()
		_, err := invoke(t, "object_union", env, box.String("x"), objArg(nil))
		require.EqualError(t, err, "object_union: first argument must be a dict")
		_, err = invoke(t, "object_union", env, objArg(nil), box.Null())
		require.EqualError(t, err, "object_union: second argument must be a dict")
		_, err = invoke(t, "object_union", env, objArg(nil))
		require.EqualError(t, err, "object_union requires 2 arguments")
	})
	t.Run("undefined propagates", func(t *testing.T) {
		t.Parallel()
		v, err := invoke(t, "object_union", env, box.Undefined(), objArg(nil))
		require.NoError(t, err)
		require.True(t, v.IsUndefined())
	})
}
