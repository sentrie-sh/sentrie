// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package js

import (
	"fmt"
	"hash/fnv"
	"math/rand/v2"
	"runtime"
	"sync"
	"time"
	"weak"

	"github.com/dop251/goja"
	"github.com/sentrie-sh/sentrie/constants"
)

// determinismByVM holds each shimmed VM's state on the Go side, so nothing of
// the runtime's internals is visible on the JavaScript global scope. Keys are
// weak, and a cleanup removes the entry when the VM is collected; the values
// never reference the VM, so the map does not keep pooled VMs alive.
var (
	determinismMu   sync.Mutex
	determinismByVM = map[weak.Pointer[goja.Runtime]]*determinism{}
)

func registerDeterminism(vm *goja.Runtime, st *determinism) {
	key := weak.Make(vm)
	determinismMu.Lock()
	determinismByVM[key] = st
	determinismMu.Unlock()
	runtime.AddCleanup(vm, func(k weak.Pointer[goja.Runtime]) {
		determinismMu.Lock()
		delete(determinismByVM, k)
		determinismMu.Unlock()
	}, key)
}

func determinismFor(vm *goja.Runtime) *determinism {
	determinismMu.Lock()
	defer determinismMu.Unlock()
	return determinismByVM[weak.Make(vm)]
}

// determinism is the per-VM state behind the deterministic shims. A pooled VM
// runs one call at a time, so BeginCall and the shims never race.
type determinism struct {
	nowMs int64
	pcg   *rand.PCG
	rng   *rand.Rand
}

// deterministicDateSource wraps the global Date so that Date.now() and a
// zero-argument `new Date()` or `Date()` read the pinned execution clock.
// Date.parse, Date.UTC, the prototype and every other constructor form are the
// real Date's.
//
// The real Date and Reflect.construct are captured when the wrapper is
// installed, before any pack code runs, and the wrapper looks up nothing on
// shared prototypes afterwards (no bind, apply, slice or concat), so pack code
// that replaces those intrinsics cannot get hold of the real Date. Forwarding
// new.target keeps `class X extends Date` instances instanceof X.
const deterministicDateSource = `(function (nowFn) {
  "use strict";
  var RealDate = Date;
  var construct = Reflect.construct;
  var defineProperty = Object.defineProperty;
  function SentrieDate() {
    if (new.target === undefined) {
      return construct(RealDate, [nowFn()]).toString();
    }
    if (arguments.length === 0) {
      return construct(RealDate, [nowFn()], new.target);
    }
    return construct(RealDate, arguments, new.target);
  }
  defineProperty(SentrieDate, "name", { value: "Date", configurable: true });
  defineProperty(SentrieDate, "length", { value: RealDate.length, configurable: true });
  defineProperty(SentrieDate, "prototype", { value: RealDate.prototype });
  SentrieDate.now = nowFn;
  SentrieDate.parse = RealDate.parse;
  SentrieDate.UTC = RealDate.UTC;
  defineProperty(RealDate.prototype, "constructor", { value: SentrieDate, writable: true, configurable: true });
  globalThis.Date = SentrieDate;
})`

var deterministicDateProgram = goja.MustCompile("sentrie:deterministic-date", deterministicDateSource, true)

// installDeterministicShims replaces Date.now, zero-argument Date construction
// and Math.random in vm with versions driven by BeginCall: the clock is the
// execution start, and Math.random is a PCG stream seeded from the execution
// start and the called module function. Same execution, same function, same
// sequence; a later execution gets a different one.
//
// Before the first BeginCall (module top-level code at load time) the clock
// reads 0 and the stream is seeded with zeros, which is still deterministic.
func installDeterministicShims(vm *goja.Runtime) error {
	st := &determinism{pcg: rand.NewPCG(0, 0)}
	st.rng = rand.New(st.pcg)

	nowFn := func(goja.FunctionCall) goja.Value { return vm.ToValue(st.nowMs) }
	randomFn := func(goja.FunctionCall) goja.Value { return vm.ToValue(st.rng.Float64()) }

	wrapper, err := vm.RunProgram(deterministicDateProgram)
	if err != nil {
		return fmt.Errorf("deterministic shims: %w", err)
	}
	install, ok := goja.AssertFunction(wrapper)
	if !ok {
		return fmt.Errorf("deterministic shims: Date wrapper is not a function")
	}
	if _, err := install(goja.Undefined(), vm.ToValue(nowFn)); err != nil {
		return fmt.Errorf("deterministic shims: %w", err)
	}

	if err := vm.Get("Math").ToObject(vm).Set("random", randomFn); err != nil {
		return fmt.Errorf("deterministic shims: %w", err)
	}

	registerDeterminism(vm, st)
	return nil
}

// BeginCall prepares a pooled VM for one call of moduleKey's export fn in the
// execution that started at start. It always pins the clock that
// `@sentrie/time.now` reads. When the deterministic shims are installed it
// also pins Date.now to start and reseeds Math.random.
func BeginCall(vm *goja.Runtime, start time.Time, moduleKey, fn string) error {
	if err := vm.Set(constants.ExecutionStartTimeUnixKey, start.UTC().Unix()); err != nil {
		return err
	}
	st := determinismFor(vm)
	if st == nil {
		return nil
	}
	st.nowMs = start.UnixMilli()
	st.pcg.Seed(uint64(start.UnixNano()), callSeed(moduleKey, fn))
	return nil
}

func callSeed(moduleKey, fn string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(moduleKey))
	_, _ = h.Write([]byte{0})
	_, _ = h.Write([]byte(fn))
	return h.Sum64()
}
