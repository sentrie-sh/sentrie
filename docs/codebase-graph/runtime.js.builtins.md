---
id: runtime.js.builtins
type: System / Package
language: Go
file_path: runtime/js/builtin_*.go
tags: standard-library, ffi, determinism, error-handling
---

# Node: js.builtins (Go-Backed Standard Library)

## 1. Architectural Role & Intent
Sixteen Go-implemented modules exposed to JavaScript under the `@sentrie/*` namespace. They exist so that common policy operations - hashing, JWT verification, CIDR arithmetic, semver comparison, regex, encoding, time - run at native speed and with vetted implementations, rather than being reimplemented in interpreted JavaScript by each pack author. Each is a `ModuleProvider` that fabricates a `module.exports` object directly in the VM, with no program compilation or execution involved.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `runtime.js.builtins` | `IMPORTS` | `ext.dop251.goja` | Every function is a `func(goja.FunctionCall) goja.Value` set on a VM object. |
| `runtime.js.builtins` | `LAYERED_ON` | [[constants]] | `ExecutionStartTimeUnixKey` is read by `time.now()` and by `jwt`'s `exp`/`nbf` checks to pin the clock. |
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
    - **Pure computation** - `math`, `string`, `collection` (the largest at ~450 lines), `semver`, `json`, `encoding`, `base64`, `url`.
    - **Cryptographic** - `hash`, `crypto`, `jwt` (HMAC, RSA and ECDSA verification, `exp`/`nbf` checks and claim extraction).
    - **Pattern and network** - `regex`, `net` (CIDR containment, intersection, validity - **arithmetic only, no sockets**).
    - **Non-deterministic** - `uuid` (fresh values per call). `time` is clock-dependent but pinned: `now()` is the execution start.
    - **TypeScript** - the `js` module, embedded from `ts_src/_sentrie_js.ts` rather than Go-implemented.
  - **Side Effects:** None escape the VM; no filesystem, network, or process access.
  - **Exceptions:** Every failure is a JavaScript exception raised through `throwErr(vm, err)` in `builtin_errors.go`, which panics with `vm.NewGoError(err)`. The conventions, applied to every module: argument-count and argument-type errors throw; predicates (`isValid`, the `net` `is*` checks, `cidrIsValid`, `jwt.verify`) answer `false` for input they cannot parse; a genuine "no result" is `null` (`net.parseIP`, `regex.find`, `semver.prerelease`/`metadata`); every other failure throws.

- **Signature:** `jwt.decode(token, secret?, options?) -> object` / `jwt.verify(token, secret, algorithm = "HS256", options?) -> boolean`
  - **Behavior:** HS256/384/512 with a shared secret; RS256/384/512 (PKCS #1 v1.5, keys of at least 2048 bits) and ES256/384/512 (fixed-width R || S) with a public key given as PEM (`PUBLIC KEY`, `RSA PUBLIC KEY`, `CERTIFICATE`) or a JWK object or JSON string, parsed in `builtin_jwt_keys.go` with the standard library only. `parseJWTKey` requires the key's form to match the algorithm's family, so a PEM or JWK is never accepted as an HMAC secret (the algorithm-confusion forgery). HMAC signatures are compared with `hmac.Equal` on decoded bytes. `exp` and `nbf` are validated by default against the execution start (Unix seconds, via `executionStartUnix` in `builtin_clock.go`), RFC 7519 style: valid while `now < exp + leeway` and once `now + leeway >= nbf`. `options` is `{ leeway?: seconds, checkTime?: boolean }`; `checkTime: false` is the inspect-only opt-out, and unknown keys throw. `verify` requires the header `alg` to equal `algorithm`. `getPayload`/`getHeader` never verify.
  - **Side Effects:** None.
  - **Exceptions:** `decode` throws on any failure, including expiry. `verify` answers `false` for a malformed, mismatched, badly signed, expired or not-yet-valid token, and throws for `none` or an unsupported `algorithm`, a key that does not suit the algorithm, bad options, or a VM with no execution start time.

- **Signature:** `time.now() -> number`
  - **Behavior:** Returns the `ExecutionStartTimeUnixKey` VM global through `executionStartUnix`, so JavaScript observes the **same pinned instant**, in the same unit (Unix seconds), as the policy evaluator's builtin `now()`.
  - **Side Effects:** None.
  - **Exceptions:** Argument-count error if given any argument; `execution start time is not set on this VM` when the global is missing. There is no wall-clock fallback (#153).

## 4. Operational Context & Gotchas
- **Statefulness:** Providers are stateless factories, but they are re-invoked on **every** `require` of a builtin because [[runtime.js.alias_runtime]] checks the provider branch before the cache. Two requires of the same builtin in one VM produce two distinct objects.
- **Performance/Scale Notes:** Native Go, so orders of magnitude faster than equivalent interpreted JavaScript - which is the main reason to prefer a builtin over a hand-rolled implementation in a pack. Object construction on each require is small but not free.
- **Dependencies Risk:**
  - **Failures throw; never return an error object.** In goja a value returned from a Go function is a normal result, so `return vm.NewGoError(...)` hands JavaScript a truthy Error object and a malformed CIDR reads as a match. That was the behaviour until [#111](https://github.com/sentrie-sh/sentrie/issues/111). Every site now uses `return throwErr(vm, err)`, and `TestBuiltinsNeverReturnErrorObjects` fails the build if `return <vm>.NewGoError(` or `NewTypeError(` reappears anywhere in `runtime/js`. An uncaught throw fails the calling rule through [[runtime.modules]].
  - **Predicates answer `false` for garbage.** `net.isPrivate("junk")` and `net.isPublic("junk")` are both `false`, so a negated predicate (`!isPrivate(ip)`) is `true` for unparseable input. Positive checks are the fail-closed form.
  - **`uuid` is non-deterministic**, which is why module calls are forbidden inside derives (see [[runtime.eval_call]]) - a derive must be reproducible within an execution.
  - **JWT time checks trust the execution-start global.** Module code can still overwrite `__executionStartTimeUnix` before calling `jwt.verify`, which would move the clock the checks use; making internal globals unreachable is [#154](https://github.com/sentrie-sh/sentrie/issues/154).
  - **`net` deliberately has no I/O.** It is CIDR arithmetic over `net.ParseCIDR`, not a network client. Preserving that is what keeps the sandbox claim honest.
  - **Argument counts are checked exactly**, so an extra argument is an error rather than being ignored - stricter than JavaScript convention and a likely source of surprise for module authors.
  - **`crypto` and `jwt` are security-critical surfaces.** Their failures throw like every other builtin's (#111).
