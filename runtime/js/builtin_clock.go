// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package js

import (
	"errors"

	"github.com/dop251/goja"
	"github.com/sentrie-sh/sentrie/constants"
)

// errNoExecutionStart reports a VM that was invoked without the execution
// start time injected. That is a programming error in the caller, never a
// reason to fall back to the wall clock.
var errNoExecutionStart = errors.New("execution start time is not set on this VM")

// executionStartUnix returns the evaluation's pinned start time in Unix
// seconds, read from the ExecutionStartTimeUnixKey global that
// runtime.ModuleBinding.Call sets before every invocation.
func executionStartUnix(vm *goja.Runtime) (int64, error) {
	v := vm.Get(constants.ExecutionStartTimeUnixKey)
	if v == nil || goja.IsUndefined(v) || goja.IsNull(v) {
		return 0, errNoExecutionStart
	}
	return v.ToInteger(), nil
}
