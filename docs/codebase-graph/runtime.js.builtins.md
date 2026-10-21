---
id: runtime.js.builtins
type: System / Package
language: Go
file_path: runtime/js/builtin_*.go
tags: standard-library, ffi, determinism, error-handling
---

# Node: js.builtins (Go-Backed Standard Library)

## 1. Architectural Role & Intent
Fourteen Go-implemented modules, plus the TypeScript `js` module, exposed to JavaScript under the `@sentrie/*` namespace. They exist so that common policy operations - hashing, JWT verification, CIDR arithmetic, semver comparison, regex, encoding, time - run at native speed and with vetted implementations, rather than being reimplemented in interpreted JavaScript by each pack author. Each is a `ModuleProvider` that fabricates a `module.exports` object directly in the VM, with no program compilation or execution involved.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `runtime.js.builtins` | `IMPORTS` | `ext.dop251.goja` | Every function is a `func(goja.FunctionCall) goja.Value` set on a VM object. |
| `runtime.js.builtins` | `LAYERED_ON` | [[constants]] | `ExecutionStartTimeUnixKey` is read by `time.now()` to pin the clock. |
| [[runtime.executor]] | `MUTATES` | `runtime.js.builtins` | Registers each provider by name at executor construction. |
| [[runtime.js.registry]] | `CALLS` | `runtime.js.builtins` | Go providers short-circuit compilation entirely in `programFor`. |
| [[runtime.js.alias_runtime]] | `CALLS` | `runtime.js.builtins` | Invokes the provider on every `require` of a builtin. |

## 3. Interface Contracts & Public Surface

- **Signature:** `ModuleProvider = func(vm *goja.Runtime) (*goja.Object, error)`
  - **Behavior:** The uniform contract. Each provider creates an object, sets its functions and constants, and returns it as `module.exports`.
  - **Side Effects:** Allocates VM objects.
  - **Exceptions:** `ErrPermissionDenied` where gated.

- **Signature:** The module set
  - **Behavior:**
    - **Pure computation** - `math`, `string`, `collection` (the largest at ~450 lines), `semver`, `json`, `encoding` (which carries base64, hex and URL encoding), `url`.
    - **Cryptographic** - `hash`, `crypto`, `jwt` (verification and claim extraction).
    - **Pattern and network** - `regex`, `net` (CIDR containment, intersection, validity - **arithmetic only, no sockets**).
    - **Non-deterministic** - `uuid` (fresh values per call), `time` (clock-dependent).
    - **TypeScript** - the `js` module, embedded from `ts_src/_sentrie_js.ts` rather than Go-implemented.
  - **Side Effects:** None escape the VM; no filesystem, network, or process access.
  - **Exceptions:** Every failure is a JavaScript exception raised through `throwErr(vm, err)` in `builtin_errors.go`, which panics with `vm.NewGoError(err)`. The conventions, applied to every module: argument-count and argument-type errors throw; predicates (`isValid`, the `net` `is*` checks, `cidrIsValid`, `jwt.verify`) answer `false` for input they cannot parse; a genuine "no result" is `null` (`net.parseIP`, `regex.find`, `semver.prerelease`/`metadata`); every other failure throws.

- **Signature:** `time.now() -> number`
  - **Behavior:** Reads the `ExecutionStartTimeUnixKey` VM global and returns it, so JavaScript observes the **same pinned instant** as the policy evaluator. Falls back to `time.Now().Unix()` when the global is absent.
  - **Side Effects:** None.
  - **Exceptions:** Argument-count error if given any argument.

## 4. Operational Context & Gotchas
- **Statefulness:** Providers are stateless factories, but they are re-invoked on **every** `require` of a builtin because [[runtime.js.alias_runtime]] checks the provider branch before the cache. Two requires of the same builtin in one VM produce two distinct objects.
- **Performance/Scale Notes:** Native Go, so orders of magnitude faster than equivalent interpreted JavaScript - which is the main reason to prefer a builtin over a hand-rolled implementation in a pack. Object construction on each require is small but not free.
- **Dependencies Risk:**
  - **Failures throw; never return an error object.** In goja a value returned from a Go function is a normal result, so `return vm.NewGoError(...)` hands JavaScript a truthy Error object and a malformed CIDR reads as a match. That was the behaviour until [#111](https://github.com/sentrie-sh/sentrie/issues/111). Every site now uses `return throwErr(vm, err)`, and `TestBuiltinsNeverReturnErrorObjects` fails the build if `return <vm>.NewGoError(` or `NewTypeError(` reappears anywhere in `runtime/js`. An uncaught throw fails the calling rule through [[runtime.modules]].
  - **Predicates answer `false` for garbage.** `net.isPrivate("junk")` and `net.isPublic("junk")` are both `false`, so a negated predicate (`!isPrivate(ip)`) is `true` for unparseable input. Positive checks are the fail-closed form.
  - **`uuid` and the `time` fallback are non-deterministic**, which is why module calls are forbidden inside derives (see [[runtime.eval_call]]) - a derive must be reproducible within an execution.
  - **The `time.now()` fallback is silent.** If the execution-start global is missing, JavaScript quietly gets wall-clock time and diverges from what the evaluator's own `now()` reports, with no signal.
  - **`net` deliberately has no I/O.** It is CIDR arithmetic over `net.ParseCIDR`, not a network client. Preserving that is what keeps the sandbox claim honest.
  - **Argument counts are checked exactly**, so an extra argument is an error rather than being ignored - stricter than JavaScript convention and a likely source of surprise for module authors.
  - **`crypto` and `jwt` are security-critical surfaces.** Their failures throw like every other builtin's (#111).
