// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"time"

	"github.com/binaek/cling"
	"github.com/sentrie-sh/sentrie/runtime"
)

// evalTimeoutFlag is the `--eval-timeout` flag shared by exec and serve (#142).
func evalTimeoutFlag() cling.CmdFlag {
	return cling.NewStringCmdInput("eval-timeout").
		WithDefault(runtime.DefaultEvalTimeout.String()).
		WithDescription("Deadline for evaluating one decision, as a Go duration (for example 500ms, 10s); 0 disables it").
		AsFlag()
}

// parseEvalTimeout validates an `--eval-timeout` value.
func parseEvalTimeout(v string) (time.Duration, error) {
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, fmt.Errorf("invalid --eval-timeout %q: %w", v, err)
	}
	if d < 0 {
		return 0, fmt.Errorf("invalid --eval-timeout %q: must not be negative", v)
	}
	return d, nil
}
