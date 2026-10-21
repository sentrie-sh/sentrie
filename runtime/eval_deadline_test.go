// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"
	"testing"
	"time"

	"github.com/sentrie-sh/sentrie/trinary"
)

// deadlinePack has only module-free rules, so nothing in it reaches the goja
// interrupt that #142 installs around JS calls (#225).
var deadlinePack = map[string]string{
	"policy.sentrie": `namespace deadline/test

derive scan = (acc: list[number], item: number): list[number] => {
  yield ((item + 0.5) in acc) ? [] : acc
}

policy p {
  fact xs: list[number]

  -- n^2 lambda applications: any() scans the whole list for every element.
  rule pairs = { yield count(filter(xs, (x: number): trinary => { yield any(xs, (y: number): trinary => { yield y < 0 }) })) == 0 }
  -- n derive calls, each an O(n) scan with no lambda inside.
  rule derived = { yield count(reduce(xs, xs, scan)) > 0 }
  rule trivial = { yield true }

  export decision of pairs
  export decision of derived
  export decision of trivial
}
`,
}

func numberList(n int) []any {
	xs := make([]any, n)
	for i := range xs {
		xs[i] = i
	}
	return xs
}

// BenchmarkCollectionLambdaRule measures the evaluator's hottest loop, a
// lambda applied per element by a collection builtin, so the cost of the
// per-iteration context check is visible (#225).
func BenchmarkCollectionLambdaRule(b *testing.B) {
	exec := buildTempPackExecutor(b, deadlinePack)
	defer func() { _ = exec.Close() }()
	facts := map[string]any{"xs": numberList(100)}
	b.ResetTimer()
	for b.Loop() {
		if _, err := exec.ExecRule(b.Context(), "deadline/test", "p", "pairs", facts); err != nil {
			b.Fatal(err)
		}
	}
}

// TestModuleFreeCollectionLoopHitsTheEvalDeadline is #225's regression test:
// a policy with no module calls iterating a large list stops at the #142
// deadline instead of running to completion. Unbounded, 5000^2 lambda
// applications take tens of seconds.
func (s *RuntimeTestSuite) TestModuleFreeCollectionLoopHitsTheEvalDeadline() {
	exec := s.newTempPackExecutor(deadlinePack, WithEvalTimeout(100*time.Millisecond))
	defer func() { s.NoError(exec.Close()) }()
	facts := map[string]any{"xs": numberList(5000)}

	start := time.Now()
	err := s.runBounded(10*time.Second, func() error {
		_, err := exec.ExecRule(s.T().Context(), "deadline/test", "p", "pairs", facts)
		return err
	})
	s.Require().Error(err)
	s.ErrorIs(err, context.DeadlineExceeded)
	s.Contains(err.Error(), "evaluation exceeded its 100ms deadline")
	s.Less(time.Since(start), 2*time.Second)
}

// TestDeriveCallbackLoopHitsTheEvalDeadline covers derive entry on its own:
// reduce calls a derive per element and the derive body has no lambda, so
// only the derive entry check can stop it.
func (s *RuntimeTestSuite) TestDeriveCallbackLoopHitsTheEvalDeadline() {
	exec := s.newTempPackExecutor(deadlinePack, WithEvalTimeout(100*time.Millisecond))
	defer func() { s.NoError(exec.Close()) }()
	facts := map[string]any{"xs": numberList(20000)}

	start := time.Now()
	err := s.runBounded(10*time.Second, func() error {
		_, err := exec.ExecRule(s.T().Context(), "deadline/test", "p", "derived", facts)
		return err
	})
	s.Require().Error(err)
	s.ErrorIs(err, context.DeadlineExceeded)
	s.Contains(err.Error(), "evaluation exceeded its 100ms deadline")
	s.Less(time.Since(start), 2*time.Second)
}

// TestEndedContextStopsAModuleFreeRule covers rule entry: a rule whose
// context has already ended does not evaluate, for ExecRule and ExecPolicy,
// while the same rule succeeds under a live context.
func (s *RuntimeTestSuite) TestEndedContextStopsAModuleFreeRule() {
	exec := s.newTempPackExecutor(deadlinePack)
	defer func() { s.NoError(exec.Close()) }()
	facts := map[string]any{"xs": numberList(3)}

	out, err := exec.ExecRule(s.T().Context(), "deadline/test", "p", "trivial", facts)
	s.Require().NoError(err)
	s.Equal(trinary.True, out.Decision.State)

	cancelled, cancel := context.WithCancel(s.T().Context())
	cancel()
	_, err = exec.ExecRule(cancelled, "deadline/test", "p", "trivial", facts)
	s.ErrorIs(err, context.Canceled)
	_, err = exec.ExecPolicy(cancelled, "deadline/test", "p", facts)
	s.ErrorIs(err, context.Canceled)
}
