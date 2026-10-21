// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package js

import (
	"runtime"
	"sync"
	"time"
	"weak"

	"github.com/dop251/goja"
)

// executionStarts holds each VM's pinned execution start, in Unix seconds,
// keyed by a weak pointer to the VM. It lives on the Go side so module code
// cannot read or overwrite it (#154); the entry is removed when the VM is
// garbage collected.
var executionStarts sync.Map // weak.Pointer[goja.Runtime] -> int64

// SetExecutionStart pins what `@sentrie/time.now()` returns inside vm.
func SetExecutionStart(vm *goja.Runtime, start time.Time) {
	key := weak.Make(vm)
	if _, loaded := executionStarts.Swap(key, start.UTC().Unix()); !loaded {
		runtime.AddCleanup(vm, func(k weak.Pointer[goja.Runtime]) { executionStarts.Delete(k) }, key)
	}
}

// executionStartUnix returns the start pinned for vm, if any.
func executionStartUnix(vm *goja.Runtime) (int64, bool) {
	v, ok := executionStarts.Load(weak.Make(vm))
	if !ok {
		return 0, false
	}
	return v.(int64), true
}
