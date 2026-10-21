// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"
	"fmt"
	"time"
)

// statefulModule accumulates state at module level: the reproduction from #105,
// with `record` renamed to `track` because `record` is a Sentrie keyword.
const statefulModule = `
let seen: string[] = [];
export function track(x: string): number {
  seen.push(x);
  return seen.length;
}

let counter = 0;
export function next(): number {
  counter += 1;
  return counter;
}
`

func (s *RuntimeTestSuite) decisionNumber(out *ExecutorOutput) float64 {
	s.Require().NotNil(out)
	s.Require().NotNil(out.Decision)
	n, ok := out.Decision.Value.NumberValue()
	s.Require().True(ok, "decision value %v is not a number", out.Decision.Value)
	return n
}

// TestModuleStateDoesNotLeakAcrossEvaluations covers #105: module-level state
// written by one evaluation is not visible to a later one.
func (s *RuntimeTestSuite) TestModuleStateDoesNotLeakAcrossEvaluations() {
	ctx := s.T().Context()
	exec := s.newTempPackExecutor(map[string]string{
		"state.ts": statefulModule,
		"policy.sentrie": `namespace isolation/test

policy p {
  use { track } from "./state" as st

  rule first = { yield st.track("a") }
  rule second = { yield st.track("b") }

  export decision of first
  export decision of second
}
`,
	})

	for range 5 {
		for _, rule := range []string{"first", "second"} {
			out, err := exec.ExecRule(ctx, "isolation/test", "p", rule, map[string]any{})
			s.Require().NoError(err)
			s.Equal(float64(1), s.decisionNumber(out), "rule %s saw state from an earlier evaluation", rule)
		}
	}
}

// TestModuleStateDoesNotLeakBetweenConcurrentRules is #105's update: ExecPolicy
// runs every exported rule concurrently, so a module-level counter used to give a
// different answer on every run depending on which pooled VM each rule drew.
func (s *RuntimeTestSuite) TestModuleStateDoesNotLeakBetweenConcurrentRules() {
	ctx := s.T().Context()
	policy := "namespace isolation/test\n\npolicy p {\n  use { next } from \"./state\" as st\n"
	for i := range 16 {
		policy += fmt.Sprintf("  rule r%d = { yield st.next() }\n  export decision of r%d\n", i, i)
	}
	policy += "}\n"
	exec := s.newTempPackExecutor(map[string]string{
		"state.ts":       statefulModule,
		"policy.sentrie": policy,
	})

	for range 10 {
		outputs, err := exec.ExecPolicy(ctx, "isolation/test", "p", map[string]any{})
		s.Require().NoError(err)
		s.Require().Len(outputs, 16)
		for _, out := range outputs {
			s.Equal(float64(1), s.decisionNumber(out), "rule %s saw another rule's counter", out.RuleName)
		}
	}
}

// TestModuleStateDoesNotLeakBetweenCallsInOneRule pins the per-acquisition
// contract: two calls in one rule body each see a freshly loaded module.
func (s *RuntimeTestSuite) TestModuleStateDoesNotLeakBetweenCallsInOneRule() {
	ctx := s.T().Context()
	exec := s.newTempPackExecutor(map[string]string{
		"state.ts": statefulModule,
		"policy.sentrie": `namespace isolation/test

policy p {
  use { track } from "./state" as st

  rule twice = {
    let a = st.track("a")
    let b = st.track("b")
    yield a + b
  }

  export decision of twice
}
`,
	})

	out, err := exec.ExecRule(ctx, "isolation/test", "p", "twice", map[string]any{})
	s.Require().NoError(err)
	s.Equal(float64(2), s.decisionNumber(out))
}

// TestFreshRuntimesReuseTheCompiledProgram measures the cost side of #105's
// option 2: every call gets a fresh goja.Runtime, but the module is transpiled
// and compiled exactly once.
func (s *RuntimeTestSuite) TestFreshRuntimesReuseTheCompiledProgram() {
	ctx := s.T().Context()
	policy := "namespace isolation/test\n\npolicy p {\n  use { next } from \"./state\" as st\n"
	for i := range 4 {
		policy += fmt.Sprintf("  rule r%d = { yield st.next() }\n  export decision of r%d\n", i, i)
	}
	policy += "}\n"
	exec := s.newTempPackExecutor(map[string]string{
		"state.ts":       statefulModule,
		"policy.sentrie": policy,
	})

	for range 5 {
		_, err := exec.ExecPolicy(ctx, "isolation/test", "p", map[string]any{})
		s.Require().NoError(err)
	}
	s.Equal(int64(1), exec.jsRegistry.CompiledPrograms())
}

// TestCancelledModuleCallDoesNotAffectTheNextCall interrupts a looping call and
// checks that the next call on the same module runs normally: no interrupt state
// survives into another acquisition.
func (s *RuntimeTestSuite) TestCancelledModuleCallDoesNotAffectTheNextCall() {
	exec := s.newTempPackExecutor(map[string]string{
		"spin.ts": `
export function spin(): number { while (true) {} }
export function one(): number { return 1; }
`,
		"policy.sentrie": `namespace isolation/test

policy p {
  use { spin, one } from "./spin" as m
  rule spinning = { yield m.spin() }
  rule fine = { yield m.one() }
  export decision of spinning
  export decision of fine
}
`,
	})

	for range 3 {
		ctx, cancel := context.WithTimeout(s.T().Context(), 50*time.Millisecond)
		_, err := exec.ExecRule(ctx, "isolation/test", "p", "spinning", map[string]any{})
		cancel()
		s.Require().Error(err)
		s.Contains(err.Error(), context.DeadlineExceeded.Error())

		out, err := exec.ExecRule(s.T().Context(), "isolation/test", "p", "fine", map[string]any{})
		s.Require().NoError(err)
		s.Equal(float64(1), s.decisionNumber(out))
	}
}
