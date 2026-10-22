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

// evalMaxTimeoutFlag is serve's `--eval-max-timeout` flag (#243): the longest
// deadline a request may ask for with the Sentrie-Timeout header. Unset, it
// equals `--eval-timeout`, so a request can only shorten its deadline.
func evalMaxTimeoutFlag() cling.CmdFlag {
	return cling.NewStringCmdInput("eval-max-timeout").
		WithDefault("").
		WithDescription("Longest deadline a request may ask for with the Sentrie-Timeout header, as a Go duration (for example 30s); unset, it equals --eval-timeout").
		AsFlag()
}

// parseEvalMaxTimeout validates an `--eval-max-timeout` value against the
// already-parsed `--eval-timeout`. It returns 0 when the flag is unset. A
// ceiling below a bounded default is an error; with an unbounded default
// (`--eval-timeout 0`) the ceiling also becomes the default.
func parseEvalMaxTimeout(v string, evalTimeout time.Duration) (time.Duration, error) {
	if v == "" {
		return 0, nil
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, fmt.Errorf("invalid --eval-max-timeout %q: %w", v, err)
	}
	if d <= 0 {
		return 0, fmt.Errorf("invalid --eval-max-timeout %q: must be positive; leave it unset for no ceiling", v)
	}
	if evalTimeout > 0 && d < evalTimeout {
		return 0, fmt.Errorf("invalid --eval-max-timeout %s: must not be below --eval-timeout %s", d, evalTimeout)
	}
	return d, nil
}
