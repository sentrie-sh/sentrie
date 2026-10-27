## Binaek Sarkar (binaek89@gmail.com)

### 14:53 - Start the deterministic JavaScript layer
- Top layer of the #168 stack, cut from `168-2-capabilities`. Adds optional deterministic `Date.now` and `Math.random` shims in pooled JavaScript runtimes, seeded from the execution context and enabled by a capabilities flag.
- _Model: claude-opus-5-5_

### 14:57 - Deterministic Date and Math.random shims
- New `deterministic_js` flag in the capabilities file (field `Deterministic`, method `DeterministicJS()`, default false). This puts the switch on the operator's grant, as #168 asks, with no new CLI flag.
- `SetupStdLib` installs the shims when the registry's grant enables them. The registry already carries the grant from layer 2, so `runtime/executor.go` (reworked by #102, #105 and #191) needed no change.
- `Date` is replaced by a small JavaScript wrapper. `Date.now`, `new Date()` and `Date()` read a Go-backed pinned clock; every other form goes to the real `Date`, and the prototype is shared, so `instanceof` and `constructor` still hold. Shimming only `Date.now` would have left `new Date()` as a wall-clock leak. `Math.random` is a PCG stream from `math/rand/v2`.
- Per-VM state sits in a non-writable, non-enumerable, non-configurable global holding an opaque Go pointer. A side map keyed by `*goja.Runtime` would have kept pooled VMs alive. JavaScript can neither replace nor enumerate the global (tested).
- `js.BeginCall(vm, start, moduleKey, fn)` replaces the one-line `ExecutionStartTimeUnixKey` set in `ModuleBinding.Call`. It pins the clock to `ec.CreatedAt()` in milliseconds and reseeds from `(UnixNano, fnv64a(moduleKey, fn))`. Seeding per call keeps the result a pure function of execution and call target, independent of which pooled VM serves it or what ran before. The cost is that two calls of the same function in one execution see the same sequence; documented.
- Before the first `BeginCall` (module top-level code at load) the clock reads 0. That is deterministic, and pools outlive executions anyway.
- `constants` is no longer imported by `runtime` (only `runtime/js` sets the key), so the `constants` graph edge moved to `runtime.js`.
- Rewrote the legacy headers of `runtime/modules.go` and `runtime/js/stdlib.go` on touch.
- Recording nondeterministic values into decision logs is still out of scope (#171 does not exist yet).
- `make test`, `make lint` and `go test -race ./runtime/js/ ./capabilities/` pass.
- _Model: claude-opus-5-5_

### 22:35 - Review fixes (pr-206)
- Critical: the `Date` wrapper called `Function.prototype.bind.apply` and `Array.prototype.slice/concat` at call time, so pack code that replaced those intrinsics received the real `Date` and with it the wall clock. The wrapper now captures `Date`, `Reflect.construct` and `Object.defineProperty` at install time (before any pack code runs) and builds every form with `construct(RealDate, args, new.target)`, with no lookups on shared prototypes. A regression test replaces `bind`, `apply`, `concat`, `slice` and `Reflect.construct` before constructing, and asserts that no real `Date` leaks. Kept it in JavaScript with captured intrinsics instead of a Go constructor: same guarantee, much less code.
- Hit and fixed a hoisting trap on the way: naming the inner function `Date` made `var RealDate = Date` capture the wrapper itself (infinite recursion). The inner function is `SentrieDate`, and `name` is defined as `"Date"`.
- Forwarding `new.target` makes `class X extends Date` instances `instanceof X` again. `Date.name` is `"Date"` and `Date.length` matches the real one (7).
- Shim state moved off the global scope into a Go-side map weakly keyed by `*goja.Runtime`, with `runtime.AddCleanup` deleting entries when a VM is collected. The values never reference the VM, so pooled VMs are not kept alive. This lines up with #215's no-internals-on-globals direction; when #215 lands, `BeginCall` should call its `SetExecutionStart` instead of setting `ExecutionStartTimeUnixKey`.
- Docs: `@sentrie/uuid` stays nondeterministic with the flag on (the earlier wording overstated it); the clock is described against `@sentrie/time.now` rather than the `now` builtin, whose unit #204 changes; the seed uses nanoseconds, so a replay must carry that precision.
- The end-to-end test now compares `floor(js.now() / 1000)` with `@sentrie/time.now` instead of the `now` builtin, so #204's unit change cannot break it.
- New tests: top-level module code reads the clock as 0; the shims add no globals.
- Left the field name `Deterministic`: `DeterministicJS` is the method, and Go does not allow a field and a method with the same name.
- The FORMAT.md line "to forbid them, leave them out of builtins and modules" is corrected in layer 2, not here, to avoid two edits of one line across the stack.
- `make test`, `make lint` and `go test -race ./runtime/js/` pass.
- _Model: claude-opus-5-5_

### 11:34 - Restack onto main after #193 and #195 merged
- Rebased onto the restacked 168-2. One conflict, in the `ModuleBinding.Call` entry of `docs/codebase-graph/runtime.modules.md`: combined this layer's `js.BeginCall` description with #193's "looks the export up through this binding's projection".
- _Model: claude-opus-5-5_
