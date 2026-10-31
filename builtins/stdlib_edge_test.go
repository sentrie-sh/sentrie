// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package builtins

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/sentrie-sh/sentrie/box"
	"github.com/stretchr/testify/require"
)

// TestGlobMatchEdges covers where `**` sits, delimiter runs, escapes of
// special characters and class corner cases that the main table does not.
func TestGlobMatchEdges(t *testing.T) {
	t.Parallel()
	slash := []any{"/"}
	cases := []struct {
		name    string
		pattern string
		delims  []any
		value   string
		want    bool
	}{
		{"double star leading crosses", "**/x", slash, "a/b/x", true},
		{"double star leading needs the slash", "**/x", slash, "x", false},
		{"double star leading matches empty prefix", "**/x", slash, "/x", true},
		{"double star trailing after slash", "/a/**", slash, "/a/b/c", true},
		{"double star trailing needs the slash", "/a/**", slash, "/a", false},
		{"double star middle spans segments", "a/**/b", slash, "a/x/y/b", true},
		{"double star middle matches one segment", "a/**/b", slash, "a/x/b", true},
		{"double star middle matches empty between slashes", "a/**/b", slash, "a//b", true},
		{"double star middle needs both slashes", "a/**/b", slash, "a/b", false},
		{"double star glued to literals", "a**b", slash, "a/x/b", true},
		{"two double stars", "**/mid/**", slash, "a/b/mid/c/d", true},
		{"two double stars miss", "**/mid/**", slash, "a/b/min/c/d", false},
		{"star one segment only", "a/*/b", slash, "a///b", false},
		{"star empty between consecutive slashes", "a/*/b", slash, "a//b", true},
		{"literal consecutive delimiters", "a//b", slash, "a//b", true},
		{"literal consecutive delimiters differ", "a//b", slash, "a/b", false},
		{"star after trailing delimiter", "a/*", slash, "a/", true},
		{"star does not cross any of several delimiters", "a*b", []any{"/", ".", ":"}, "a.b", false},
		{"star crosses undeclared delimiter", "a*b", []any{"/", ":"}, "a.b", true},
		{"duplicate delimiters are harmless", "a*b", []any{"/", "/"}, "a/b", false},
		{"question mark does not match any of several delimiters", "a?b", []any{"/", "."}, "a.b", false},
		{"class does not match a delimiter", "a[/x]b", slash, "a/b", false},
		{"class may hold a non-delimiter", "a[/x]b", slash, "axb", true},
		{"escaped question mark", `a\?b`, slash, "a?b", true},
		{"escaped question mark is literal only", `a\?b`, slash, "axb", false},
		{"escaped slash", `a\/b`, slash, "a/b", true},
		{"escaped backslash", `a\\b`, slash, `a\b`, true},
		{"escaped open bracket", `a\[b`, slash, "a[b", true},
		{"escaped open brace", `a\{b\}`, slash, "a{b}", true},
		{"escaped regular character", `\a`, slash, "a", true},
		{"escaped double star is two literals", `\*\*`, slash, "**", true},
		{"escaped double star does not cross", `\*\*`, slash, "ab", false},
		{"star after an escaped star", `\**`, slash, "*abc", true},
		{"class escaped dash is a member", `[a\-c]`, slash, "-", true},
		{"class escaped dash is not a range", `[a\-c]`, slash, "b", false},
		{"class dash first is literal", `[-a]`, slash, "-", true},
		{"class multiple ranges", "[a-cx-z]", slash, "y", true},
		{"class multiple ranges miss", "[a-cx-z]", slash, "m", false},
		{"class unicode range", "[α-ω]", slash, "λ", true},
		{"class unicode range miss", "[α-ω]", slash, "a", false},
		{"class single rune range", "[a-a]", slash, "a", true},
		{"negated class several ranges", "[!a-cx-z]", slash, "m", true},
		{"unicode star", "*é*", slash, "caféx", true},
		{"unicode pattern literal", "日本/*", slash, "日本/語", true},
		{"unicode question counts runes not bytes", "?", slash, "日", true},
		{"unicode question two runes", "?", slash, "日本", false},
		{"unicode double star", "/a/**", slash, "/a/日/本", true},
		{"value longer than pattern", "abc", slash, "abcd", false},
		{"value shorter than pattern", "abcd", slash, "abc", false},
		{"star at end matches empty", "abc*", slash, "abc", true},
		{"star at start matches empty", "*abc", slash, "abc", true},
		{"only star matches empty", "*", slash, "", true},
		{"only star rejects delimiter", "*", slash, "/", false},
		{"only question mark rejects empty", "?", slash, "", false},
		{"question marks fix the length", "???", slash, "abc", true},
		{"question marks fix the length miss", "???", slash, "ab", false},
		{"question crosses a non-delimiter slash", "a?b", []any{"."}, "a/b", true},
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

// TestGlobMatchCapBoundaries pins the limits: exactly at a cap is accepted,
// one over is rejected, and the caps count characters, not bytes.
func TestGlobMatchCapBoundaries(t *testing.T) {
	t.Parallel()
	slash := []any{"/"}

	t.Run("pattern exactly at the cap", func(t *testing.T) {
		t.Parallel()
		got, err := globCall(t, strings.Repeat("a", maxGlobPatternRunes), slash, strings.Repeat("a", maxGlobPatternRunes))
		require.NoError(t, err)
		require.True(t, got)
	})
	t.Run("pattern one over the cap", func(t *testing.T) {
		t.Parallel()
		_, err := globCall(t, strings.Repeat("a", maxGlobPatternRunes+1), slash, "a")
		require.ErrorContains(t, err, "pattern is longer than")
	})
	t.Run("pattern cap counts runes not bytes", func(t *testing.T) {
		t.Parallel()
		got, err := globCall(t, strings.Repeat("é", maxGlobPatternRunes), slash, strings.Repeat("é", maxGlobPatternRunes))
		require.NoError(t, err)
		require.True(t, got)
		_, err = globCall(t, strings.Repeat("é", maxGlobPatternRunes+1), slash, "a")
		require.ErrorContains(t, err, "pattern is longer than")
	})
	t.Run("value exactly at the cap", func(t *testing.T) {
		t.Parallel()
		got, err := globCall(t, "**", slash, strings.Repeat("a/", maxGlobValueRunes/2))
		require.NoError(t, err)
		require.True(t, got)
	})
	t.Run("value one over the cap", func(t *testing.T) {
		t.Parallel()
		_, err := globCall(t, "**", slash, strings.Repeat("a", maxGlobValueRunes+1))
		require.ErrorContains(t, err, "value is longer than")
	})
	t.Run("value cap counts runes not bytes", func(t *testing.T) {
		t.Parallel()
		got, err := globCall(t, "*", slash, strings.Repeat("é", maxGlobValueRunes))
		require.NoError(t, err)
		require.True(t, got)
		_, err = globCall(t, "*", slash, strings.Repeat("é", maxGlobValueRunes+1))
		require.ErrorContains(t, err, "value is longer than")
	})
	t.Run("value cap is checked before the pattern is parsed", func(t *testing.T) {
		t.Parallel()
		_, err := globCall(t, "{", slash, strings.Repeat("a", maxGlobValueRunes+1))
		require.ErrorContains(t, err, "value is longer than")
	})
	t.Run("worst case at both caps stays correct", func(t *testing.T) {
		t.Parallel()
		pattern := strings.Repeat("*a", maxGlobPatternRunes/2-1) + "*b"
		got, err := globCall(t, pattern, slash, strings.Repeat("a", maxGlobValueRunes))
		require.NoError(t, err)
		require.False(t, got)
		got, err = globCall(t, pattern, slash, strings.Repeat("a", maxGlobValueRunes-1)+"b")
		require.NoError(t, err)
		require.True(t, got)
	})
}

// TestGlobMatchSyntaxErrorsAreIndependentOfValue pins that a malformed pattern
// errors even when the value could never have reached the bad token.
func TestGlobMatchSyntaxErrorsAreIndependentOfValue(t *testing.T) {
	t.Parallel()
	for _, value := range []string{"", "x", "a/b"} {
		_, err := globCall(t, "a[", []any{"/"}, value)
		require.ErrorContains(t, err, "unterminated character class", "value %q", value)
	}
	_, err := globCall(t, "[!", []any{"/"}, "x")
	require.ErrorContains(t, err, "unterminated character class")
}

// TestParseQuantityPrecision pins that quantities are computed exactly and
// rounded once to the nearest float64, and how negatives and edge numbers
// behave.
func TestParseQuantityPrecision(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	cases := []struct {
		in   string
		want float64
	}{
		{"300m", 0.3},
		{"700m", 0.7},
		{"0.1", 0.1},
		{"0.3", 0.3},
		{"1.1Gi", 1181116006.4},
		{"1.5Mi", 1572864},
		{"0.1k", 100},
		{"1.25M", 1_250_000},
		{"-500m", -0.5},
		{"-1.5Ki", -1536},
		{"-2e3", -2000},
		{"-3k", -3000},
		{"+500m", 0.5},
		{"007", 7},
		{"0.500", 0.5},
		{"0", 0},
		{"0Gi", 0},
		{"9007199254740993", 9007199254740992},
		{"1e18", 1e18},
		{"1E", 1e18},
		{"1Ei", 1 << 60},
		{"8Ei", 1 << 63},
		{"1e308", 1e308},
		{"123456789u", 123.456789},
		{"1000000000n", 1},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()
			v, err := invoke(t, "parse_quantity", env, box.String(tc.in))
			require.NoError(t, err)
			got, ok := v.NumberValue()
			require.True(t, ok)
			require.Equal(t, tc.want, got)
		})
	}

	for _, in := range []string{"1.8e308", "1e309"} {
		t.Run("overflow "+in, func(t *testing.T) {
			t.Parallel()
			_, err := invoke(t, "parse_quantity", env, box.String(in))
			require.ErrorContains(t, err, "out of range")
		})
	}

	t.Run("trailing dot before an exponent is a valid decimal", func(t *testing.T) {
		t.Parallel()
		v, err := invoke(t, "parse_quantity", env, box.String("1.e3"))
		require.NoError(t, err)
		require.Equal(t, 1000.0, v.Any())
	})

	for _, in := range []string{"++1", "+-1", "--1", "1.5e3Gi", "1e3.5", "1,5", "0x10", "1_000", "NaN", "Inf", "\t1", "1\n"} {
		t.Run("rejects "+fmt.Sprintf("%q", in), func(t *testing.T) {
			t.Parallel()
			v, err := invoke(t, "parse_quantity", env, box.String(in))
			require.Error(t, err)
			require.True(t, v.IsUndefined())
		})
	}

	t.Run("string at the length cap", func(t *testing.T) {
		t.Parallel()
		s := strings.Repeat("0", maxUnitStringLen-1) + "1"
		v, err := invoke(t, "parse_quantity", env, box.String(s))
		require.NoError(t, err)
		require.Equal(t, 1.0, v.Any())
	})
}

func TestParseDurationEdges(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	ok := []struct {
		in   string
		want float64
	}{
		{"0s", 0},
		{"-0", 0},
		{"2562047h", 2562047 * 3600},
		{"-2562047h", -2562047 * 3600},
		{"1μs", 0.000001},
		{"1ns", 1e-9},
		{"1h1m1s1ms", 3661.001},
		{"0.5m", 30},
		{".5s", 0.5},
		{"-1h30m", -5400},
		{"1m1m", 120},
	}
	for _, tc := range ok {
		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()
			v, err := invoke(t, "parse_duration", env, box.String(tc.in))
			require.NoError(t, err)
			require.InDelta(t, tc.want, v.Any(), 1e-9)
		})
	}
	for _, in := range []string{"2562048h", "9999999999h", "-2562048h", "1.5", "--5m", "5m-", "1h-5m", " 1h", "1h ", "ms", "1M", "1H", "1S", "NaN", "Inf"} {
		t.Run("bad "+fmt.Sprintf("%q", in), func(t *testing.T) {
			t.Parallel()
			v, err := invoke(t, "parse_duration", env, box.String(in))
			require.Error(t, err)
			require.True(t, v.IsUndefined())
			require.ErrorContains(t, err, "parse_duration: ")
		})
	}
}

// TestObjectGetPathEdges covers list index and intermediate-key edge cases.
func TestObjectGetPathEdges(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	def := box.String("DEFAULT")
	doc := objArg(map[string]any{
		"grid": []any{[]any{"a", "b"}, []any{"c"}},
		"rows": []any{map[string]any{"id": 1.0}, map[string]any{"id": 2.0, "tags": []any{"x", "y"}}},
		"deep": map[string]any{"l1": map[string]any{"l2": map[string]any{"l3": "bottom"}}},
		"":     "empty key",
		"e":    []any{},
		"d":    map[string]any{},
	})
	cases := []struct {
		name string
		obj  box.Value
		path box.Value
		want any
	}{
		{"list of lists", doc, box.FromAny([]any{"grid", 0.0, 1.0}), "b"},
		{"second row second tag", doc, box.FromAny([]any{"rows", 1.0, "tags", 1.0}), "y"},
		{"first row has no tags", doc, box.FromAny([]any{"rows", 0.0, "tags", 0.0}), "DEFAULT"},
		{"inner list out of range", doc, box.FromAny([]any{"grid", 1.0, 1.0}), "DEFAULT"},
		{"missing at the first step", doc, box.FromAny([]any{"nope", "x", "y"}), "DEFAULT"},
		{"missing at the last step", doc, box.FromAny([]any{"deep", "l1", "l2", "nope"}), "DEFAULT"},
		{"three levels deep", doc, box.FromAny([]any{"deep", "l1", "l2", "l3"}), "bottom"},
		{"past the leaf", doc, box.FromAny([]any{"deep", "l1", "l2", "l3", "x"}), "DEFAULT"},
		{"empty string key", doc, box.String(""), "empty key"},
		{"empty list indexed", doc, box.FromAny([]any{"e", 0.0}), "DEFAULT"},
		{"empty dict keyed", doc, box.FromAny([]any{"d", "k"}), "DEFAULT"},
		{"huge index", doc, box.FromAny([]any{"rows", 1e30}), "DEFAULT"},
		{"nan index", doc, box.FromAny([]any{"rows", math.NaN()}), "DEFAULT"},
		{"infinite index", doc, box.FromAny([]any{"rows", math.Inf(1)}), "DEFAULT"},
		{"negative zero index", doc, box.FromAny([]any{"grid", math.Copysign(0, -1), 0.0}), "a"},
		{"whole-valued float index", doc, box.FromAny([]any{"grid", 1.0, 0.0}), "c"},
		{"top level list by single number", box.FromAny([]any{"p", "q", "r"}), box.Number(2), "r"},
		{"top level list by single number out of range", box.FromAny([]any{"p"}), box.Number(1), "DEFAULT"},
		{"top level scalar object", box.String("s"), box.FromAny([]any{"a"}), "DEFAULT"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			v, err := invoke(t, "object_get", env, tc.obj, tc.path, def)
			require.NoError(t, err)
			require.Equal(t, tc.want, v.Any())
		})
	}

	t.Run("non-dict default is returned as given", func(t *testing.T) {
		t.Parallel()
		v, err := invoke(t, "object_get", env, doc, box.FromAny([]any{"nope"}), box.FromAny([]any{1.0, 2.0}))
		require.NoError(t, err)
		require.Equal(t, []any{1.0, 2.0}, v.Any())
	})
	t.Run("path error is raised even when the object is empty", func(t *testing.T) {
		t.Parallel()
		_, err := invoke(t, "object_get", env, objArg(map[string]any{}), box.FromAny([]any{"a", true}), def)
		require.ErrorContains(t, err, "path segments must be strings or non-negative integers")
	})
	t.Run("nested list segment is rejected", func(t *testing.T) {
		t.Parallel()
		_, err := invoke(t, "object_get", env, doc, box.FromAny([]any{[]any{"a"}}), def)
		require.ErrorContains(t, err, "path segments must be strings or non-negative integers")
	})
}

func TestObjectKeysAndSelectEdges(t *testing.T) {
	t.Parallel()
	env := noopEnv()

	t.Run("keys sort by byte order", func(t *testing.T) {
		t.Parallel()
		v, err := invoke(t, "object_keys", env, objArg(map[string]any{"é": 1.0, "z": 2.0, "Z": 3.0, "10": 4.0, "9": 5.0}))
		require.NoError(t, err)
		require.Equal(t, []any{"10", "9", "Z", "z", "é"}, v.Any())
	})
	t.Run("keys do not descend", func(t *testing.T) {
		t.Parallel()
		v, err := invoke(t, "object_keys", env, objArg(map[string]any{"a": map[string]any{"b": 1.0}}))
		require.NoError(t, err)
		require.Equal(t, []any{"a"}, v.Any())
	})
	t.Run("keys is stable across many calls on a large dict", func(t *testing.T) {
		t.Parallel()
		m := map[string]any{}
		for i := range 200 {
			m[fmt.Sprintf("k%03d", i)] = float64(i)
		}
		first, err := invoke(t, "object_keys", env, objArg(m))
		require.NoError(t, err)
		for range 20 {
			again, err := invoke(t, "object_keys", env, objArg(m))
			require.NoError(t, err)
			require.Equal(t, first.Any(), again.Any())
		}
		got := first.Any().([]any)
		require.Len(t, got, 200)
		require.Equal(t, "k000", got[0])
		require.Equal(t, "k199", got[199])
	})
	t.Run("filter with duplicate keys", func(t *testing.T) {
		t.Parallel()
		v, err := invoke(t, "object_filter", env, objArg(map[string]any{"a": 1.0, "b": 2.0}), box.FromAny([]any{"a", "a"}))
		require.NoError(t, err)
		require.Equal(t, map[string]any{"a": 1.0}, v.Any())
	})
	t.Run("remove with duplicate keys", func(t *testing.T) {
		t.Parallel()
		v, err := invoke(t, "object_remove", env, objArg(map[string]any{"a": 1.0, "b": 2.0}), box.FromAny([]any{"a", "a"}))
		require.NoError(t, err)
		require.Equal(t, map[string]any{"b": 2.0}, v.Any())
	})
	t.Run("filter is top level only", func(t *testing.T) {
		t.Parallel()
		v, err := invoke(t, "object_filter", env, objArg(map[string]any{"a": map[string]any{"x": 1.0}}), box.FromAny([]any{"x"}))
		require.NoError(t, err)
		require.Equal(t, map[string]any{}, v.Any())
	})
	t.Run("remove on an empty dict", func(t *testing.T) {
		t.Parallel()
		v, err := invoke(t, "object_remove", env, objArg(map[string]any{}), box.FromAny([]any{"a"}))
		require.NoError(t, err)
		require.Equal(t, map[string]any{}, v.Any())
	})
	t.Run("filter result is a copy", func(t *testing.T) {
		t.Parallel()
		in := objArg(map[string]any{"a": map[string]any{"x": 1.0}})
		out, err := invoke(t, "object_filter", env, in, box.FromAny([]any{"a"}))
		require.NoError(t, err)
		inner, ok := out.DictValue()
		require.True(t, ok)
		nested, ok := inner["a"].DictValue()
		require.True(t, ok)
		nested["x"] = box.Number(99)
		require.Equal(t, map[string]any{"a": map[string]any{"x": 1.0}}, in.Any())
	})
	t.Run("union with empty dicts", func(t *testing.T) {
		t.Parallel()
		a := objArg(map[string]any{"a": 1.0})
		v, err := invoke(t, "object_union", env, a, objArg(map[string]any{}))
		require.NoError(t, err)
		require.Equal(t, map[string]any{"a": 1.0}, v.Any())
		v, err = invoke(t, "object_union", env, objArg(map[string]any{}), a)
		require.NoError(t, err)
		require.Equal(t, map[string]any{"a": 1.0}, v.Any())
	})
	t.Run("union second null overrides a dict", func(t *testing.T) {
		t.Parallel()
		v, err := invoke(t, "object_union", env, objArg(map[string]any{"a": map[string]any{"x": 1.0}}), objArg(map[string]any{"a": nil}))
		require.NoError(t, err)
		require.Equal(t, map[string]any{"a": nil}, v.Any())
	})
	t.Run("union merges three levels", func(t *testing.T) {
		t.Parallel()
		a := objArg(map[string]any{"a": map[string]any{"b": map[string]any{"c": 1.0, "d": 2.0}}})
		b := objArg(map[string]any{"a": map[string]any{"b": map[string]any{"d": 3.0, "e": 4.0}}})
		v, err := invoke(t, "object_union", env, a, b)
		require.NoError(t, err)
		require.Equal(t, map[string]any{"a": map[string]any{"b": map[string]any{"c": 1.0, "d": 3.0, "e": 4.0}}}, v.Any())
	})
}
