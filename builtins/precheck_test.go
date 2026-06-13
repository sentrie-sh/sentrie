// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package builtins

import (
	"context"
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
