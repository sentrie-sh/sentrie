// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"
	"time"
)

// TestContextEvalTimeoutReplacesTheExecutorDeadline is #243's runtime half: a
// deadline carried by ContextWithEvalTimeout takes the place of the
// executor's, shorter or longer, and is the one the error names.
func (s *RuntimeTestSuite) TestContextEvalTimeoutReplacesTheExecutorDeadline() {
	cases := []struct {
		name     string
		executor time.Duration
		override time.Duration
		applied  string
		atLeast  time.Duration
	}{
		{name: "shorter than the executor's", executor: 5 * time.Second, override: 100 * time.Millisecond, applied: "100ms"},
		{name: "executor deadline disabled", executor: 0, override: 100 * time.Millisecond, applied: "100ms"},
		{name: "longer than the executor's", executor: 50 * time.Millisecond, override: 300 * time.Millisecond, applied: "300ms", atLeast: 300 * time.Millisecond},
		{name: "zero is ignored", executor: 100 * time.Millisecond, override: 0, applied: "100ms"},
		{name: "negative is ignored", executor: 100 * time.Millisecond, override: -time.Second, applied: "100ms"},
	}
	for _, tc := range cases {
		s.Run(tc.name, func() {
			exec := s.newTempPackExecutor(limitsPack, WithEvalTimeout(tc.executor))
			defer func() { s.NoError(exec.Close()) }()

			ctx := ContextWithEvalTimeout(s.T().Context(), tc.override)
			start := time.Now()
			err := s.runBounded(5*time.Second, func() error {
				_, err := exec.ExecRule(ctx, "limits/test", "p", "loops", map[string]any{})
				return err
			})
			elapsed := time.Since(start)
			s.Require().Error(err)
			s.ErrorIs(err, context.DeadlineExceeded)
			s.Contains(err.Error(), "evaluation exceeded its "+tc.applied+" deadline")
			s.GreaterOrEqual(elapsed, tc.atLeast)
			s.Less(elapsed, 2*time.Second)
		})
	}
}
