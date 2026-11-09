// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package builtins

import (
	"testing"

	"github.com/sentrie-sh/sentrie/box"
	"github.com/stretchr/testify/require"
)

func TestParseQuantity(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	ok := []struct {
		in   string
		want float64
	}{
		{"0", 0},
		{"42", 42},
		{"+42", 42},
		{"-42", -42},
		{"1.5", 1.5},
		{".5", 0.5},
		{"5.", 5},
		{"500m", 0.5},
		{"100m", 0.1},
		{"1500m", 1.5},
		{"250u", 0.00025},
		{"1n", 1e-9},
		{"1k", 1000},
		{"1.5k", 1500},
		{"2M", 2_000_000},
		{"3G", 3e9},
		{"4T", 4e12},
		{"5P", 5e15},
		{"6E", 6e18},
		{"1Ki", 1024},
		{"1Mi", 1 << 20},
		{"128Mi", 128 << 20},
		{"10Gi", 10 << 30},
		{"2Ti", 2 << 40},
		{"3Pi", 3 << 50},
		{"1Ei", 1 << 60},
		{"0.5Gi", 1 << 29},
		{"-1Gi", -(1 << 30)},
		{"2e3", 2000},
		{"2E3", 2000},
		{"1e-3", 0.001},
		{"12E+2", 1200},
		{"1e0", 1},
	}
	for _, tc := range ok {
		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()
			v, err := invoke(t, "parse_quantity", env, box.String(tc.in))
			require.NoError(t, err)
			require.Equal(t, tc.want, v.Any())
		})
	}

	bad := []struct{ in, msg string }{
		{"", "empty quantity"},
		{"Gi", "no number"},
		{"abc", "no number"},
		{"-", "no number"},
		{".", "no number"},
		{"1.2.3", "unknown suffix"},
		{"10GB", "unknown suffix"},
		{"10K", "unknown suffix"},
		{"10gi", "unknown suffix"},
		{"10 Gi", "unknown suffix"},
		{" 10", "no number"},
		{"10 ", "unknown suffix"},
		{"1e", "unknown suffix"},
		{"1e+", "unknown suffix"},
		{"1e--2", "unknown suffix"},
		{"1ex", "unknown suffix"},
		{"1e400", "out of range"},
		{"1e401", "out of range"},
		{"1e-401", "out of range"},
		{"1E999999999999", "out of range"},
		{"1e308", ""},
	}
	for _, tc := range bad {
		t.Run("bad "+tc.in, func(t *testing.T) {
			t.Parallel()
			v, err := invoke(t, "parse_quantity", env, box.String(tc.in))
			if tc.msg == "" {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			require.True(t, v.IsUndefined(), "an error must not carry a value")
			require.ErrorContains(t, err, "parse_quantity: ")
			require.ErrorContains(t, err, tc.msg)
		})
	}

	t.Run("underflow to zero errors, subnormal is kept", func(t *testing.T) {
		t.Parallel()
		_, err := invoke(t, "parse_quantity", env, box.String("1e-400"))
		require.ErrorContains(t, err, "out of range")
		_, err = invoke(t, "parse_quantity", env, box.String("1n"))
		require.NoError(t, err)
		v, err := invoke(t, "parse_quantity", env, box.String("1e-323"))
		require.NoError(t, err)
		require.Greater(t, v.Any(), 0.0)
	})
	t.Run("overflow to infinity errors", func(t *testing.T) {
		t.Parallel()
		_, err := invoke(t, "parse_quantity", env, box.String("1e400"))
		require.Error(t, err)
	})
	t.Run("too long", func(t *testing.T) {
		t.Parallel()
		long := make([]byte, maxUnitStringLen+1)
		for i := range long {
			long[i] = '1'
		}
		_, err := invoke(t, "parse_quantity", env, box.String(string(long)))
		require.ErrorContains(t, err, "longer than")
	})
	t.Run("wrong kind", func(t *testing.T) {
		t.Parallel()
		for _, v := range []box.Value{box.Number(5), box.Null(), box.FromAny([]any{})} {
			_, err := invoke(t, "parse_quantity", env, v)
			require.EqualError(t, err, "parse_quantity: argument must be a string")
		}
	})
	t.Run("undefined propagates", func(t *testing.T) {
		t.Parallel()
		v, err := invoke(t, "parse_quantity", env, box.Undefined())
		require.NoError(t, err)
		require.True(t, v.IsUndefined())
	})
	t.Run("arity", func(t *testing.T) {
		t.Parallel()
		_, err := invoke(t, "parse_quantity", env)
		require.EqualError(t, err, "parse_quantity requires 1 argument")
		_, err = invoke(t, "parse_quantity", env, box.String("1"), box.String("1"))
		require.EqualError(t, err, "parse_quantity requires 1 argument")
	})
}

func TestParseDuration(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	ok := []struct {
		in   string
		want float64
	}{
		{"0", 0},
		{"2h", 7200},
		{"90m", 5400},
		{"30s", 30},
		{"1h30m", 5400},
		{"1h30m15s", 5415},
		{"1.5h", 5400},
		{"250ms", 0.25},
		{"1500us", 0.0015},
		{"1500µs", 0.0015},
		{"2000ns", 0.000002},
		{"-5m", -300},
		{"+5m", 300},
		{"500m", 30000},
	}
	for _, tc := range ok {
		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()
			v, err := invoke(t, "parse_duration", env, box.String(tc.in))
			require.NoError(t, err)
			require.InDelta(t, tc.want, v.Any(), 1e-12)
		})
	}
	for _, in := range []string{"", "5", "1d", "1w", "abc", "h", "1 h", "5x", "1h 30m", "1e3s"} {
		t.Run("bad "+in, func(t *testing.T) {
			t.Parallel()
			v, err := invoke(t, "parse_duration", env, box.String(in))
			require.Error(t, err)
			require.True(t, v.IsUndefined())
			require.ErrorContains(t, err, "parse_duration: ")
		})
	}
	t.Run("wrong kind", func(t *testing.T) {
		t.Parallel()
		_, err := invoke(t, "parse_duration", env, box.Number(5))
		require.EqualError(t, err, "parse_duration: argument must be a string")
		_, err = invoke(t, "parse_duration", env, box.Null())
		require.EqualError(t, err, "parse_duration: argument must be a string")
	})
	t.Run("undefined propagates", func(t *testing.T) {
		t.Parallel()
		v, err := invoke(t, "parse_duration", env, box.Undefined())
		require.NoError(t, err)
		require.True(t, v.IsUndefined())
	})
	t.Run("arity", func(t *testing.T) {
		t.Parallel()
		_, err := invoke(t, "parse_duration", env)
		require.EqualError(t, err, "parse_duration requires 1 argument")
	})
}
