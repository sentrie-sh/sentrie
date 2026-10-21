## Binaek Sarkar (binaek89@gmail.com)

### 21:55 - Pick up #224
- Branch cut from `154-js-boundary-values`. `runtime/js_values.go` exists only from #154 on.
- Opened as a draft PR before implementation, per the issue workflow.
- _Model: claude-opus-5-5_

### 22:02 - Probe how goja reports the opaque built-ins
- Probed goja (cf18d89f3cf6) with a throwaway `sandbox.go` and a throwaway suite method, both removed:
  - `WeakMap`, `WeakSet`, `DataView`: class `Object`, export type `map[string]any`, so the converter returns `{}`.
  - `ArrayBuffer`: class `Object`, export type `goja.ArrayBuffer` (a struct), so it takes the wrapped-Go path and `structs.Map` turns it into `{}`.
  - Typed arrays: class `Object`, export type `[]uint8`, `[]float64` and so on, so they pass through the wrapped-Go path as raw Go slices, which `box.FromAny` turns into an opaque `Document`. The #154 table already says typed arrays are an error, so this was a second silent gap.
  - `WeakRef`, `FinalizationRegistry`, `SharedArrayBuffer`, `Atomics`: not defined at all in goja, so a module cannot return one.
- Decision: detect by brand check against the VM's own intrinsics rather than by prototype identity. `WeakMap.prototype.has` / `WeakSet.prototype.has` throw a TypeError on any receiver without the internal slot, so subclasses and objects whose prototype was swapped are still caught, and plain objects whose prototype is `WeakMap.prototype` (no slot) are not. `ArrayBuffer.isView` covers every typed array and `DataView` in one non-throwing call. `ArrayBuffer` itself is identified by its export type, which only goja's array buffer object produces. All three are native and run no module code.
- Decision: reject typed arrays, `ArrayBuffer` and `DataView` rather than convert them to lists. That keeps the documented #154 table, and `Array.from(x)` is the explicit, lossless route.
- _Model: claude-opus-5-5_

### 22:03 - Reject WeakMap, WeakSet and binary data at the JS boundary
- `runtime/js_values.go`: in the class `Object` branch, every export type other than `Map`'s and `Set`'s now goes through `opaqueBuiltin` first. It checks the `goja.ArrayBuffer` export type, then `ArrayBuffer.isView` (named from `Symbol.toStringTag`, so the error says `Uint8Array` or `DataView`), then the `WeakMap` / `WeakSet` `has` brand checks. The intrinsics are resolved once per conversion, lazily, and a missing global just skips its check.
- A `*goja.InterruptedError` from a brand-check call is returned rather than read as "no slot", so a call whose context ends mid-conversion does not quietly convert a WeakMap to `{}`.
- Checked that a Go `[]byte` from a builtin (`crypto` returns one) is untouched: it is class `Array` in goja, and `ArrayBuffer.isView` is false for it. Covered by a test.
- Tests on `RuntimeTestSuite`: `TestJSToSentrieRejectsOpaqueBuiltins` (paths, subclasses, swapped prototypes, slot-less `Object.create(WeakMap.prototype)`, `Array.from`, absent `WeakRef` / `FinalizationRegistry`, deleted globals) and `TestModuleReturningWeakMapIsAnError` (the issue's reproduction end to end). Both fail with the converter change stashed.
- Coordinator heads-up: the #154 author is fixing getter / Proxy-trap panics, returned Proxies and unbounded sparse arrays in the same file on `154-js-boundary-values`. Kept this change to the opaque built-in detection and did not touch getters, Proxy or array length, so the two merge cleanly. A returned Proxy still becomes `{}` here; that is the #154 fix's to own.
- Graph: `runtime.modules` conversion list and the class-`Object` gotcha updated.
- The TypeScript module guide is not in this repo, so its table update goes in the PR description for the docs follow-up.
- _Model: claude-opus-5-5_
