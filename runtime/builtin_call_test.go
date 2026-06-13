// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
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
