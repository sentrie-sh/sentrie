// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sentrie-sh/sentrie/trinary"
)

var cachePack = map[string]string{
	"helpers.ts": `
export function yes(): boolean { return true; }
export function no(): boolean { return false; }
`,
	"policy.sentrie": `namespace cache/test

policy p {
  use { yes } from "./helpers" as h
  use { no } from "./helpers" as h2
  rule r = { yield h.yes() and not h2.no() }
  export decision of r
}
`,
}

// TestModuleLoaderRunsOnceAcrossExecRule is #191's regression test: repeated
// ExecRule calls for the same `use` reuse the loaded module instead of
// rebuilding it on every evaluation.
func (s *RuntimeTestSuite) TestModuleLoaderRunsOnceAcrossExecRule() {
	ctx := s.T().Context()
	exec := s.newTempPackExecutor(cachePack)
	defer func() { s.NoError(exec.Close()) }()

	for range 10 {
		out, err := exec.ExecRule(ctx, "cache/test", "p", "r", map[string]any{})
		s.Require().NoError(err)
		s.Equal(trinary.True, out.Decision.State)
	}
	s.Equal(int64(1), exec.moduleLoads.Load(), "two uses of one module across ten evaluations must load it once")
	s.Equal(int64(1), exec.jsRegistry.CompiledPrograms())
}

// TestConcurrentFirstUseLoadsTheModuleOnce exercises the singleflight path: all
// rules of a fresh executor bind the module at the same time.
func (s *RuntimeTestSuite) TestConcurrentFirstUseLoadsTheModuleOnce() {
	ctx := s.T().Context()
	policy := "namespace cache/test\n\npolicy p {\n  use { yes } from \"./helpers\" as h\n"
	for i := range 16 {
		policy += fmt.Sprintf("  rule r%d = { yield h.yes() }\n  export decision of r%d\n", i, i)
	}
	policy += "}\n"
	exec := s.newTempPackExecutor(map[string]string{
		"helpers.ts":     cachePack["helpers.ts"],
		"policy.sentrie": policy,
	})
	defer func() { s.NoError(exec.Close()) }()

	outputs, err := exec.ExecPolicy(ctx, "cache/test", "p", map[string]any{})
	s.Require().NoError(err)
	s.Len(outputs, 16)
	s.Equal(int64(1), exec.moduleLoads.Load())
}

// TestFailedModuleLoadIsRetried checks that a load which failed because its
// caller's context ended is not cached for later callers.
func (s *RuntimeTestSuite) TestFailedModuleLoadIsRetried() {
	exec := s.newTempPackExecutor(cachePack)
	defer func() { s.NoError(exec.Close()) }()

	cancelled, cancel := context.WithCancel(s.T().Context())
	cancel()
	_, err := exec.ExecRule(cancelled, "cache/test", "p", "r", map[string]any{})
	s.Require().Error(err)
	s.ErrorIs(err, context.Canceled)

	out, err := exec.ExecRule(s.T().Context(), "cache/test", "p", "r", map[string]any{})
	s.Require().NoError(err)
	s.Equal(trinary.True, out.Decision.State)
	s.Equal(int64(2), exec.moduleLoads.Load(), "the cancelled load must not be cached")
}

// TestCloseReleasesModulesAndRejectsLaterLoads covers the executor lifecycle:
// Close drops every loaded module, is idempotent, and later evaluations that
// need a module fail with ErrExecutorClosed.
func (s *RuntimeTestSuite) TestCloseReleasesModulesAndRejectsLaterLoads() {
	ctx := s.T().Context()
	exec := s.newTempPackExecutor(cachePack)

	_, err := exec.ExecRule(ctx, "cache/test", "p", "r", map[string]any{})
	s.Require().NoError(err)
	exec.modulesMu.Lock()
	var loaded []*jsModule
	for _, l := range exec.modules {
		loaded = append(loaded, l.module)
	}
	exec.modulesMu.Unlock()
	s.Require().Len(loaded, 1)

	s.Require().NoError(exec.Close())
	s.Require().NoError(exec.Close(), "Close is idempotent")

	for _, m := range loaded {
		_, _, err := m.acquire(ctx, time.Now())
		s.ErrorIs(err, ErrExecutorClosed)
	}
	_, err = exec.ExecRule(ctx, "cache/test", "p", "r", map[string]any{})
	s.Require().Error(err)
	s.ErrorIs(err, ErrExecutorClosed)
}

// slowLoadPack's module spends about 60ms in its top level, so tests can act
// while a load is in flight.
var slowLoadPack = map[string]string{
	"m.ts": `const t0 = Date.now(); while (Date.now() - t0 < 60) {}
export function f(): number { return 1; }`,
	"policy.sentrie": `namespace cache/test

policy slow {
  use { f } from "./m" as m
  rule r = { yield m.f() == 1 }
  export decision of r
}
`,
}

// TestCloseDuringLoad covers both Close branches for a load in flight: Close
// skips it, and the loader, seeing the executor closed, closes its own module.
// Every concurrent first user gets a result or ErrExecutorClosed.
func (s *RuntimeTestSuite) TestCloseDuringLoad() {
	exec := s.newTempPackExecutor(slowLoadPack)
	var wg sync.WaitGroup
	errs := make([]error, 8)
	for i := range errs {
		wg.Go(func() {
			_, errs[i] = exec.ExecRule(s.T().Context(), "cache/test", "slow", "r", map[string]any{})
		})
	}
	time.Sleep(20 * time.Millisecond)
	s.Require().NoError(exec.Close())
	wg.Wait()
	for _, err := range errs {
		if err != nil {
			s.ErrorIs(err, ErrExecutorClosed)
		}
	}
}

// TestWaiterRetriesAfterLoaderIsCancelled: a caller that joined a load which
// failed only because the loader's context ended retries with its own context.
func (s *RuntimeTestSuite) TestWaiterRetriesAfterLoaderIsCancelled() {
	exec := s.newTempPackExecutor(slowLoadPack)
	defer func() { s.NoError(exec.Close()) }()

	short, cancel := context.WithTimeout(s.T().Context(), 20*time.Millisecond)
	defer cancel()
	var wg sync.WaitGroup
	var loaderErr, waiterErr error
	wg.Go(func() { _, loaderErr = exec.ExecRule(short, "cache/test", "slow", "r", map[string]any{}) })
	time.Sleep(5 * time.Millisecond)
	wg.Go(func() { _, waiterErr = exec.ExecRule(s.T().Context(), "cache/test", "slow", "r", map[string]any{}) })
	wg.Wait()

	s.ErrorIs(loaderErr, context.DeadlineExceeded)
	s.NoError(waiterErr)
	s.Equal(int64(2), exec.moduleLoads.Load())
}

// TestCloseDoesNotWaitForABusyExportCheck covers the #205 review finding:
// Close must not wait on a module's probe lock (held while an export check
// runs module code) while it holds the executor's module lock.
func (s *RuntimeTestSuite) TestCloseDoesNotWaitForABusyExportCheck() {
	exec := s.newTempPackExecutor(cachePack)
	_, err := exec.ExecRule(s.T().Context(), "cache/test", "p", "r", map[string]any{})
	s.Require().NoError(err)
	exec.modulesMu.Lock()
	var m *jsModule
	for _, l := range exec.modules {
		m = l.module
	}
	exec.modulesMu.Unlock()

	m.probeMu.Lock() // stands in for an export getter that never returns
	defer m.probeMu.Unlock()
	closed := make(chan error, 1)
	go func() { closed <- exec.Close() }()
	select {
	case err := <-closed:
		s.NoError(err)
	case <-time.After(2 * time.Second):
		s.FailNow("Close blocked on a busy export check")
	}
	s.ErrorIs(m.checkExports([]string{"yes"}), ErrExecutorClosed)
}

// TestCallOnABindingAfterCloseIsExecutorClosed: a binding made before Close
// fails at call time with an error matching ErrExecutorClosed.
func (s *RuntimeTestSuite) TestCallOnABindingAfterCloseIsExecutorClosed() {
	exec := s.newTempPackExecutor(cachePack)
	p := exec.index.Namespaces["cache/test"].Policies["p"]
	ec := NewExecutionContext(p, exec)
	s.Require().NoError(exec.bindUses(s.T().Context(), ec, p))
	binding, ok := ec.Module("h")
	s.Require().True(ok)
	s.Require().NoError(exec.Close())

	_, err := binding.Call(s.T().Context(), ec, "yes")
	s.ErrorIs(err, ErrExecutorClosed)
}
