// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"
	"testing"
	"time"

	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/builtins"
	"github.com/stretchr/testify/require"
)

func TestCallSiteImplementsBuiltinsEnv(t *testing.T) {
	t.Parallel()
	var _ builtins.Env = (*CallSite)(nil)
}

func TestCallSiteEnvExecutionStartAndCallableArity(t *testing.T) {
	t.Parallel()
	p := newEvalTestPolicy()
	ec := NewExecutionContext(p, &executorImpl{})
	start := time.Unix(1_700_000_000, 123_000_000)
	ec.createdAt = start
	site := &CallSite{EC: ec, Exec: &executorImpl{}, Policy: p}

	require.Equal(t, start, site.ExecutionStart())

	_, err := site.CallableArity(box.Number(1))
	require.Error(t, err)
	require.ErrorContains(t, err, "expected callable, got number")
}

func TestCallSiteEnvCall(t *testing.T) {
	t.Parallel()
	p := newEvalTestPolicy()
	ec := NewExecutionContext(p, &executorImpl{})
	site := &CallSite{EC: ec, Exec: &executorImpl{}, Policy: p}

	fn := box.Callable(callableStub{
		arity: 1,
		fn: func(_ context.Context, args []box.Value) (box.Value, error) {
			return box.Number(len(args)), nil
		},
	})
	out, err := site.Call(t.Context(), fn, []box.Value{box.Number(1)})
	require.NoError(t, err)
	require.Equal(t, 1.0, out.Any())

	_, err = site.Call(t.Context(), box.Number(1), nil)
	require.Error(t, err)
	require.ErrorContains(t, err, "expected callable, got number")
}

type callableStub struct {
	arity int
	fn    func(context.Context, []box.Value) (box.Value, error)
}

func (s callableStub) Arity() int { return s.arity }

func (s callableStub) Invoke(ctx context.Context, _ *CallSite, args []box.Value) (box.Value, error) {
	return s.fn(ctx, args)
}
