// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package builtins

import (
	"strings"
	"testing"

	"github.com/sentrie-sh/sentrie/box"
	"github.com/stretchr/testify/require"
)

func globCall(t *testing.T, pattern string, delims []any, value string) (bool, error) {
	t.Helper()
	v, err := invoke(t, "glob_match", noopEnv(), box.String(pattern), box.FromAny(delims), box.String(value))
	if err != nil {
		return false, err
	}
	b, ok := v.BoolValue()
	require.True(t, ok, "result must be a bool, got %s", v.Kind())
	return b, nil
}

func TestGlobMatch(t *testing.T) {
	t.Parallel()
	slash := []any{"/"}
	dot := []any{"."}
	cases := []struct {
		name    string
		pattern string
		delims  []any
		value   string
		want    bool
	}{
		{"literal equal", "/repos/a", slash, "/repos/a", true},
		{"literal differs", "/repos/a", slash, "/repos/b", false},
		{"star within segment", "/repos/*/issues", slash, "/repos/x/issues", true},
		{"star does not cross delimiter", "/repos/*/issues", slash, "/repos/x/y/issues", false},
		{"star matches empty", "/repos/*/issues", slash, "/repos//issues", true},
		{"star prefix in segment", "/repos/ab*", slash, "/repos/abcd", true},
		{"star prefix does not cross", "/repos/ab*", slash, "/repos/ab/cd", false},
		{"double star crosses", "/repos/*/issues/**", slash, "/repos/x/issues/1/comments/2", true},
		{"double star matches empty tail", "/repos/*/issues/**", slash, "/repos/x/issues/", true},
		{"double star needs the literal before it", "/repos/*/issues/**", slash, "/repos/x/issues", false},
		{"double star alone", "**", slash, "/any/thing/at/all", true},
		{"double star alone empty value", "**", slash, "", true},
		{"three stars are a double star", "a/***", slash, "a/b/c", true},
		{"dot delimiter", "*.example.com", dot, "api.example.com", true},
		{"dot delimiter blocks subdomain", "*.example.com", dot, "a.b.example.com", false},
		{"dot delimiter double star", "**.example.com", dot, "a.b.example.com", true},
		{"multiple delimiters", "a*b", []any{"/", ":"}, "a:b", false},
		{"question matches one", "a?c", slash, "abc", true},
		{"question not delimiter", "a?c", slash, "a/c", false},
		{"question needs one char", "a?c", slash, "ac", false},
		{"class match", "file[0-9].txt", dot, "file7.txt", true},
		{"class miss", "file[0-9].txt", dot, "filex.txt", false},
		{"negated class bang", "f[!a-c]", slash, "fd", true},
		{"negated class caret", "f[^a-c]", slash, "fb", false},
		{"negated class not delimiter", "f[!a]", slash, "f/", false},
		{"class literal members", "[xyz]", slash, "y", true},
		{"class trailing dash is literal", "[a-]", slash, "-", true},
		{"class escaped bracket", `[\]]`, slash, "]", true},
		{"escaped star", `a\*b`, slash, "a*b", true},
		{"escaped star is literal only", `a\*b`, slash, "axb", false},
		{"unicode value", "caf?", slash, "café", true},
		{"unicode delimiter", "a*b", []any{"é"}, "aéb", false},
		{"empty pattern empty value", "", slash, "", true},
		{"empty pattern non-empty value", "", slash, "a", false},
		{"adjacent stars with literal", "*a*a*a*b", slash, strings.Repeat("a", 60), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := globCall(t, tc.pattern, tc.delims, tc.value)
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestGlobMatchPathologicalIsLinear(t *testing.T) {
	t.Parallel()
	pattern := strings.Repeat("*a", 100) + "b"
	got, err := globCall(t, pattern, []any{"/"}, strings.Repeat("a", 4000))
	require.NoError(t, err)
	require.False(t, got)
}

func TestGlobMatchErrors(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	s, l := box.String("a"), box.List(nil)

	pattern := []struct{ name, pattern, msg string }{
		{"trailing backslash", `a\`, "unescaped backslash"},
		{"unterminated class", "[abc", "unterminated character class"},
		{"unterminated class after escape", `[a\`, "unterminated character class"},
		{"empty class", "[]", "empty character class"},
		{"empty negated class", "[!]", "empty character class"},
		{"backwards range", "[z-a]", "invalid range"},
		{"open brace", "{a,b}", "alternation"},
		{"close brace", "a}", "alternation"},
		{"pattern too long", strings.Repeat("a", maxGlobPatternRunes+1), "longer than"},
	}
	for _, tc := range pattern {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := globCall(t, tc.pattern, []any{"/"}, "a")
			require.ErrorContains(t, err, tc.msg)
		})
	}

	t.Run("value too long", func(t *testing.T) {
		t.Parallel()
		_, err := globCall(t, "a", []any{"/"}, strings.Repeat("a", maxGlobValueRunes+1))
		require.ErrorContains(t, err, "longer than")
	})
	t.Run("pattern wrong kind", func(t *testing.T) {
		t.Parallel()
		_, err := invoke(t, "glob_match", env, box.Number(1), l, s)
		require.EqualError(t, err, "glob_match: pattern must be a string")
	})
	t.Run("delimiters wrong kind", func(t *testing.T) {
		t.Parallel()
		_, err := invoke(t, "glob_match", env, s, box.String("/"), s)
		require.EqualError(t, err, "glob_match: delimiters must be a list of single-character strings")
	})
	t.Run("empty delimiter list", func(t *testing.T) {
		t.Parallel()
		_, err := globCall(t, "*.example.com", []any{}, "evil.attacker.example.com")
		require.ErrorContains(t, err, "delimiters must not be empty")
	})
	t.Run("delimiter too long", func(t *testing.T) {
		t.Parallel()
		_, err := globCall(t, "a", []any{"//"}, "a")
		require.ErrorContains(t, err, "single-character")
	})
	t.Run("delimiter empty", func(t *testing.T) {
		t.Parallel()
		_, err := globCall(t, "a", []any{""}, "a")
		require.ErrorContains(t, err, "single-character")
	})
	t.Run("delimiter not a string", func(t *testing.T) {
		t.Parallel()
		_, err := globCall(t, "a", []any{1.0}, "a")
		require.ErrorContains(t, err, "single-character")
	})
	t.Run("value wrong kind", func(t *testing.T) {
		t.Parallel()
		_, err := invoke(t, "glob_match", env, s, l, box.Number(1))
		require.EqualError(t, err, "glob_match: value must be a string")
	})
	t.Run("null pattern errors", func(t *testing.T) {
		t.Parallel()
		_, err := invoke(t, "glob_match", env, box.Null(), l, s)
		require.EqualError(t, err, "glob_match: pattern must be a string")
	})
	t.Run("arity", func(t *testing.T) {
		t.Parallel()
		_, err := invoke(t, "glob_match", env, s, l)
		require.EqualError(t, err, "glob_match requires 3 arguments")
		_, err = invoke(t, "glob_match", env, s, l, s, s)
		require.EqualError(t, err, "glob_match requires 3 arguments")
	})
}

func TestGlobMatchUndefinedPropagates(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	s, l := box.String("a"), box.List(nil)
	for name, args := range map[string][]box.Value{
		"pattern":    {box.Undefined(), l, s},
		"delimiters": {s, box.Undefined(), s},
		"value":      {s, l, box.Undefined()},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			v, err := invoke(t, "glob_match", env, args...)
			require.NoError(t, err)
			require.True(t, v.IsUndefined())
		})
	}
}
