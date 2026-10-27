// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package js

import (
	"errors"

	"github.com/dop251/goja"
)

// errNoExecutionStart reports a VM that was invoked without the execution
// start time pinned. That is a programming error in the caller, never a
// reason to fall back to the wall clock.
var errNoExecutionStart = errors.New("execution start time is not set on this VM")

// executionStartSeconds returns the evaluation's pinned start time in Unix
// seconds, as SetExecutionStart recorded it before the invocation.
func executionStartSeconds(vm *goja.Runtime) (int64, error) {
	ts, ok := executionStartUnix(vm)
	if !ok {
		return 0, errNoExecutionStart
	}
	return ts, nil
}
