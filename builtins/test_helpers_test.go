// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package builtins

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/sentrie-sh/sentrie/box"
)

type testEnv struct {
	startTime time.Time
}

func (e *testEnv) Call(_ context.Context, fn box.Value, args []box.Value) (box.Value, error) {
	ref, ok := fn.CallableRef()
	if !ok {
		return box.Undefined(), fmt.Errorf("expected callable, got %s", fn.Kind())
	}
	sc, ok := ref.(stubCallable)
	if !ok {
		return box.Undefined(), fmt.Errorf("internal error: callable payload is %T", ref)
	}
	if sc.fn == nil {
		return box.Undefined(), nil
	}
	return sc.fn(args)
}

func (e *testEnv) CallableArity(fn box.Value) (int, error) {
	ref, ok := fn.CallableRef()
	if !ok {
		return 0, fmt.Errorf("expected callable, got %s", fn.Kind())
	}
	sc, ok := ref.(stubCallable)
	if !ok {
		return 0, fmt.Errorf("internal error: callable payload is %T", ref)
	}
	return sc.arity, nil
}

func (e *testEnv) ExecutionStart() time.Time {
	if e.startTime.IsZero() {
		return time.Unix(1_700_000_000, 0)
	}
	return e.startTime
}

type stubCallable struct {
	arity int
	fn    func(args []box.Value) (box.Value, error)
}

func boxArgs(parts ...any) []box.Value {
	out := make([]box.Value, len(parts))
	for i, p := range parts {
		out[i] = box.FromAny(p)
	}
	return out
}

func invoke(t *testing.T, name string, env Env, args ...box.Value) (box.Value, error) {
	t.Helper()
	decl := Table[name]
	if decl == nil {
		t.Fatalf("unknown builtin %q", name)
	}
	handled, v, err := decl.Precheck(env, args)
	if handled || err != nil {
		return v, err
	}
	return decl.Impl(t.Context(), env, args...)
}

func noopEnv() *testEnv {
	return &testEnv{}
}
