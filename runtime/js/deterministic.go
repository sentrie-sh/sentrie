// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package js

import (
	"fmt"
	"hash/fnv"
	"math/rand/v2"
	"time"

	"github.com/dop251/goja"
)

// determinismStateKey names the non-writable, non-enumerable global that
// carries a VM's shim state. JavaScript sees an opaque object with no fields.
const determinismStateKey = "__sentrie_determinism"

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
const deterministicDateSource = `(function (nowFn) {
  var RealDate = Date;
  function SentrieDate() {
    if (!new.target) {
      return new RealDate(nowFn()).toString();
    }
    if (arguments.length === 0) {
      return new RealDate(nowFn());
    }
    var args = Array.prototype.slice.call(arguments);
    return new (Function.prototype.bind.apply(RealDate, [null].concat(args)))();
  }
  SentrieDate.prototype = RealDate.prototype;
  SentrieDate.now = nowFn;
  SentrieDate.parse = RealDate.parse;
  SentrieDate.UTC = RealDate.UTC;
  Object.defineProperty(RealDate.prototype, "constructor", { value: SentrieDate, writable: true, configurable: true });
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

	return vm.GlobalObject().DefineDataProperty(determinismStateKey, vm.ToValue(st), goja.FLAG_FALSE, goja.FLAG_FALSE, goja.FLAG_FALSE)
}

// BeginCall prepares a VM for one call of moduleKey's export fn in the
// execution that started at start. It always pins the clock that
// `@sentrie/time.now` reads (SetExecutionStart). When the deterministic shims
// are installed it also pins Date.now to start and reseeds Math.random.
func BeginCall(vm *goja.Runtime, start time.Time, moduleKey, fn string) error {
	SetExecutionStart(vm, start)
	v := vm.Get(determinismStateKey)
	if v == nil {
		return nil
	}
	st, ok := v.Export().(*determinism)
	if !ok {
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
