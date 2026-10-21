// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"
	"errors"
	"time"

	"github.com/dop251/goja"
	"github.com/sentrie-sh/sentrie/runtime/js"
)

var limitsPack = map[string]string{
	"bad.ts": `
export function spin(): number { while (true) {} }
export function deep(n: number): number { return deep(n + 1) + 1; }
`,
	"policy.sentrie": `namespace limits/test

policy p {
  use { spin, deep } from "./bad" as m
  rule loops = { yield m.spin() }
  rule recurses = { yield m.deep(0) }
  export decision of loops
  export decision of recurses
}
`,
}

// runBounded runs f and fails the test if it does not return within limit, so
// a regression shows up as a failure rather than a hung test binary.
func (s *RuntimeTestSuite) runBounded(limit time.Duration, f func() error) error {
	done := make(chan error, 1)
	go func() { done <- f() }()
	select {
	case err := <-done:
		return err
	case <-time.After(limit):
		s.FailNow("evaluation did not return", "still running after %s", limit)
		return nil
	}
}

// TestInfiniteLoopModuleHitsTheEvalDeadline is #142's first regression test.
func (s *RuntimeTestSuite) TestInfiniteLoopModuleHitsTheEvalDeadline() {
	exec := s.newTempPackExecutor(limitsPack, WithEvalTimeout(100*time.Millisecond))
	defer func() { s.NoError(exec.Close()) }()

	start := time.Now()
	err := s.runBounded(5*time.Second, func() error {
		_, err := exec.ExecRule(s.T().Context(), "limits/test", "p", "loops", map[string]any{})
		return err
	})
	s.Require().Error(err)
	s.ErrorIs(err, context.DeadlineExceeded)
	s.Contains(err.Error(), "evaluation exceeded its 100ms deadline")
	s.Less(time.Since(start), 2*time.Second)
}

// TestDeepRecursionModuleHitsTheStackLimit is #142's second regression test.
// The deadline is generous so the failure must come from the call-depth limit.
func (s *RuntimeTestSuite) TestDeepRecursionModuleHitsTheStackLimit() {
	exec := s.newTempPackExecutor(limitsPack, WithEvalTimeout(5*time.Second))
	defer func() { s.NoError(exec.Close()) }()

	err := s.runBounded(5*time.Second, func() error {
		_, err := exec.ExecRule(s.T().Context(), "limits/test", "p", "recurses", map[string]any{})
		return err
	})
	s.Require().Error(err)
	var overflow *goja.StackOverflowError
	s.True(errors.As(err, &overflow), "want a stack overflow, got %v", err)
	s.NotErrorIs(err, context.DeadlineExceeded)
}

// TestExecPolicyIsBoundedAsAWhole checks that the deadline covers every rule
// ExecPolicy runs concurrently.
func (s *RuntimeTestSuite) TestExecPolicyIsBoundedAsAWhole() {
	exec := s.newTempPackExecutor(limitsPack, WithEvalTimeout(100*time.Millisecond))
	defer func() { s.NoError(exec.Close()) }()

	err := s.runBounded(5*time.Second, func() error {
		_, err := exec.ExecPolicy(s.T().Context(), "limits/test", "p", map[string]any{})
		return err
	})
	s.Require().Error(err)
	s.ErrorIs(err, context.DeadlineExceeded)
}

// TestCallerDeadlineIsNotRelabelled keeps the error untouched when the caller's
// own, shorter deadline is what ended the evaluation.
func (s *RuntimeTestSuite) TestCallerDeadlineIsNotRelabelled() {
	exec := s.newTempPackExecutor(limitsPack)
	defer func() { s.NoError(exec.Close()) }()
	s.Equal(DefaultEvalTimeout, exec.evalTimeout)

	ctx, cancel := context.WithTimeout(s.T().Context(), 50*time.Millisecond)
	defer cancel()
	err := s.runBounded(5*time.Second, func() error {
		_, err := exec.ExecRule(ctx, "limits/test", "p", "loops", map[string]any{})
		return err
	})
	s.Require().ErrorIs(err, context.DeadlineExceeded)
	s.NotContains(err.Error(), "evaluation exceeded")
}

func (s *RuntimeTestSuite) TestEveryRuntimeHasTheCallStackLimit() {
	rt := js.NewAliasRuntime(js.NewRegistry("."), ".").VM
	_, err := rt.RunString(`(function f(n) { return f(n + 1); })(0)`)
	var overflow *goja.StackOverflowError
	s.Require().ErrorAs(err, &overflow)
}
