// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"
	"fmt"
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
		s.ErrorIs(err, errModuleClosed)
	}
	_, err = exec.ExecRule(ctx, "cache/test", "p", "r", map[string]any{})
	s.Require().Error(err)
	s.ErrorIs(err, ErrExecutorClosed)
}
