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
