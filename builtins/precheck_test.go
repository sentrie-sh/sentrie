// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package builtins

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/sentrie-sh/sentrie/box"
	"github.com/stretchr/testify/require"
)

// lenientEnv returns (0, nil) from CallableArity for any value — must not be
// relied on for callable kind rejection; Precheck must error on its own.
type lenientEnv struct{}

func (e *lenientEnv) CallableArity(_ box.Value) (int, error) {
	return 0, nil
}

func (e *lenientEnv) Call(_ context.Context, _ box.Value, _ []box.Value) (box.Value, error) {
	return box.Undefined(), nil
}

func (e *lenientEnv) ExecutionStart() time.Time {
	return time.Unix(0, 0)
}

func TestPrecheckCallableKindMismatchDoesNotRelyOnCallableArity(t *testing.T) {
	t.Parallel()
	decl := Table["collect"]
	list := box.List([]box.Value{box.Number(1)})
	_, _, err := decl.Precheck(&lenientEnv{}, []box.Value{list, box.Number(9)})
	require.Error(t, err)
	require.ErrorContains(t, err, "expected callable, got number")
}

type arityErrorEnv struct {
	lenientEnv
}

func (e *arityErrorEnv) CallableArity(_ box.Value) (int, error) {
	return 0, fmt.Errorf("arity probe failed")
}

func TestPrecheckCallableArityEnvError(t *testing.T) {
	t.Parallel()
	decl := Table["filter"]
	list := box.List([]box.Value{box.Number(1)})
	fn := box.Callable(stubCallable{arity: 1})
	_, _, err := decl.Precheck(&arityErrorEnv{}, []box.Value{list, fn})
	require.Error(t, err)
	require.ErrorContains(t, err, "arity probe failed")
}

func TestPrecheckTooManyFixedArityArgs(t *testing.T) {
	t.Parallel()
	decl := Table["filter"]
	list := box.List([]box.Value{box.Number(1)})
	fn := box.Callable(stubCallable{arity: 1})
	_, _, err := decl.Precheck(noopEnv(), []box.Value{list, fn, box.Number(9)})
	require.Error(t, err)
	require.ErrorContains(t, err, "requires 2 arguments")
}

func TestPrecheckVariadicTail(t *testing.T) {
	t.Parallel()
	decl := Table["error"]
	handled, _, err := decl.Precheck(noopEnv(), []box.Value{box.String("%v"), box.Number(1), box.Number(2)})
	require.False(t, handled)
	require.NoError(t, err)
}

func TestParamSigAtFixedArityOnly(t *testing.T) {
	t.Parallel()
	sig := Sig{Params: []ParamSig{{Name: "only"}}}
	got, ok := paramSigAt(sig, 0)
	require.True(t, ok)
	require.Equal(t, "only", got.Name)
	_, ok = paramSigAt(sig, 1)
	require.False(t, ok)
}

func TestPrecheckTooManyUsesTooFewWhenTooManyEmpty(t *testing.T) {
	t.Parallel()
	decl := &Decl{
		Name: "synthetic",
		Sig: Sig{
			Params:      []ParamSig{{Name: "x"}},
			TooFewError: "synthetic requires 1 argument",
			TooManyError: "",
		},
		Impl: func(_ context.Context, _ Env, _ ...box.Value) (box.Value, error) {
			return box.Undefined(), nil
		},
	}
	_, _, err := decl.Precheck(noopEnv(), []box.Value{box.Number(1), box.Number(2)})
	require.Error(t, err)
	require.ErrorContains(t, err, "synthetic requires 1 argument")
}
