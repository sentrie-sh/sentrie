// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package builtins

import (
	"testing"
	"time"

	"github.com/sentrie-sh/sentrie/box"
	"github.com/stretchr/testify/require"
)

func TestTableWellFormed(t *testing.T) {
	t.Parallel()

	// Issue §5 / Issue B: Result.Kinds must be populated (nil only where noted).
	expectedResultKinds := map[string][]box.ValueKind{
		"all":            kindBool,
		"any":            kindBool,
		"as_list":        kindList,
		"collect":        kindList,
		"count":          kindNumber,
		"distinct":       kindList,
		"error":          nil,
		"filter":         kindList,
		"first":          nil,
		"flatten":        kindList,
		"flatten_deep":   kindList,
		"glob_match":     kindBool,
		"merge":          kindDict,
		"normalise_list": kindList,
		"now":            kindNumber,
		"now_ms":         kindNumber,
		"object_filter":  kindDict,
		"object_get":     nil,
		"object_keys":    kindList,
		"object_remove":  kindDict,
		"object_union":   kindDict,
		"parse_duration": kindNumber,
		"parse_quantity": kindNumber,
		"reduce":         nil,
	}

	for key, d := range Table {
		require.Equal(t, key, d.Name, "map key must equal Decl.Name")
		require.NotEmpty(t, d.Description)
		require.NotNil(t, d.Impl)
		require.True(t, d.DeriveSafe)

		wantKinds, ok := expectedResultKinds[d.Name]
		require.True(t, ok, "builtin %q missing from expectedResultKinds", d.Name)
		require.Equal(t, wantKinds, d.Sig.Result.Kinds, "builtin %q Result.Kinds", d.Name)

		optionalSeen := false
		for _, p := range d.Sig.Params {
			if p.Optional {
				optionalSeen = true
			} else if optionalSeen {
				t.Fatalf("builtin %q: optional params must be trailing", d.Name)
			}
			if len(p.CallableArities) > 0 {
				require.Equal(t, []box.ValueKind{box.ValueCallable}, p.Kinds)
			}
		}
		if d.Sig.Variadic != nil {
			require.False(t, d.Sig.Variadic.Optional, "variadic param must not be optional")
			require.NotEmpty(t, d.Sig.TooFewError, "builtin %q: variadic must set TooFewError", d.Name)
		} else {
			require.NotEmpty(t, d.Sig.TooFewError, "builtin %q: non-variadic must set TooFewError", d.Name)
			require.NotEmpty(t, d.Sig.TooManyError, "builtin %q: non-variadic must set TooManyError", d.Name)
		}
	}
}

func TestDeriveSafeNamesMatchTable(t *testing.T) {
	t.Parallel()
	names := DeriveSafeNames()
	require.Equal(t, len(Table)-len(NondeterministicNames()), len(names))
	for _, name := range names {
		require.Contains(t, Table, name)
		require.False(t, Table[name].Nondeterministic, "builtin %q", name)
	}
}

func TestIsDeriveSafe(t *testing.T) {
	t.Parallel()
	require.True(t, IsDeriveSafe("count"))
	require.False(t, IsDeriveSafe("now"), "nondeterministic builtins are not permitted in derives")
	require.False(t, IsDeriveSafe("now_ms"), "nondeterministic builtins are not permitted in derives")
	require.False(t, IsDeriveSafe("not_a_builtin"))
}

// The clock and randomness sources are the nondeterministic builtins. A new
// builtin that reads either must be added here and marked Nondeterministic.
func TestNondeterministicNames(t *testing.T) {
	t.Parallel()
	require.Equal(t, []string{"now", "now_ms"}, NondeterministicNames())
	require.True(t, IsNondeterministic("now"))
	require.True(t, IsNondeterministic("now_ms"))
	require.False(t, IsNondeterministic("count"))
	require.False(t, IsNondeterministic("not_a_builtin"))
}

func TestPermittedInDerive(t *testing.T) {
	t.Parallel()
	require.True(t, (&Decl{DeriveSafe: true}).PermittedInDerive())
	require.False(t, (&Decl{DeriveSafe: false}).PermittedInDerive())
	require.False(t, (&Decl{DeriveSafe: true, Nondeterministic: true}).PermittedInDerive())
	var nilDecl *Decl
	require.False(t, nilDecl.PermittedInDerive())
}

func TestGoldenBehavior(t *testing.T) {
	t.Parallel()
	env := noopEnv()
	ctx := t.Context()

	t.Run("count multibyte bytes", func(t *testing.T) {
		v, err := invoke(t, "count", env, box.String("café"))
		require.NoError(t, err)
		require.Equal(t, 5.0, v.Any())
	})

	t.Run("count wrong kind undefined", func(t *testing.T) {
		v, err := invoke(t, "count", env, box.Number(5))
		require.NoError(t, err)
		require.True(t, v.IsUndefined())
	})

	t.Run("merge undefined first", func(t *testing.T) {
		_, err := invoke(t, "merge", env, box.Undefined(), box.Dict(map[string]box.Value{}))
		require.Error(t, err)
		require.ErrorContains(t, err, "first argument is not a dict")
	})

	t.Run("merge bad bad arg0 first", func(t *testing.T) {
		_, err := invoke(t, "merge", env, box.Number(1), box.Number(2))
		require.Error(t, err)
		require.ErrorContains(t, err, "first argument is not a dict")
	})

	t.Run("filter undefined empty list", func(t *testing.T) {
		v, err := invoke(t, "filter", env, box.Undefined(), box.Callable(stubCallable{
			arity: 1,
			fn:    func([]box.Value) (box.Value, error) { return box.Bool(true), nil },
		}))
		require.NoError(t, err)
		require.Equal(t, []any{}, v.Any())
	})

	t.Run("any undefined false", func(t *testing.T) {
		v, err := invoke(t, "any", env, box.Undefined(), box.Callable(stubCallable{arity: 1}))
		require.NoError(t, err)
		require.Equal(t, false, v.Any())
	})

	t.Run("all undefined false", func(t *testing.T) {
		v, err := invoke(t, "all", env, box.Undefined(), box.Callable(stubCallable{arity: 1}))
		require.NoError(t, err)
		require.Equal(t, false, v.Any())
	})

	t.Run("merge undefined second", func(t *testing.T) {
		_, err := invoke(t, "merge", env, box.Dict(map[string]box.Value{}), box.Undefined())
		require.Error(t, err)
		require.ErrorContains(t, err, "second argument is not a dict")
	})

	t.Run("now seconds and now_ms milliseconds", func(t *testing.T) {
		fixed := &testEnv{startTime: time.Unix(1_700_000_000, 123_000_000)}
		v, err := Table["now"].Impl(ctx, fixed)
		require.NoError(t, err)
		require.Equal(t, float64(1_700_000_000), v.Any())
		v, err = Table["now_ms"].Impl(ctx, fixed)
		require.NoError(t, err)
		require.Equal(t, float64(1_700_000_000_123), v.Any())
	})

	t.Run("distinct undefined errors", func(t *testing.T) {
		_, err := invoke(t, "distinct", env, box.Undefined())
		require.Error(t, err)
		require.ErrorContains(t, err, "distinct: first argument must be a list")
	})
}
