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
    - **Pure computation** - `math`, `string`, `collection` (the largest at ~450 lines), `semver`, `json`, `encoding` (which carries base64, hex and URL encoding), `url`.
    - **Cryptographic** - `hash`, `crypto`, `jwt` (verification and claim extraction).
    - **Pattern and network** - `regex`, `net` (CIDR containment, intersection, validity - **arithmetic only, no sockets**).
    - **Non-deterministic** - `uuid` (fresh values per call), `time` (clock-dependent).
    - **TypeScript** - the `js` module, embedded from `ts_src/_sentrie_js.ts` rather than Go-implemented.
  - **Side Effects:** None escape the VM; no filesystem, network, or process access.
  - **Exceptions:** Every failure is a JavaScript exception raised through `throwErr(vm, err)` in `builtin_errors.go`, which panics with `vm.NewGoError(err)`. The conventions, applied to every module: argument-count and argument-type errors throw; predicates (`isValid`, the `net` `is*` checks, `cidrIsValid`, `jwt.verify`) answer `false` for input they cannot parse; a genuine "no result" is `null` (`net.parseIP`, `regex.find`, `semver.prerelease`/`metadata`); every other failure throws.

- **Signature:** `jwt.decode(token, secret, options?) -> object` / `jwt.verify(token, secret, algorithm = "HS256", options?) -> boolean`
  - **Behavior:** HS256/384/512 only. The signature is compared with `hmac.Equal` on decoded bytes. `exp` and `nbf` are validated by default against the execution start (Unix seconds, via `executionStartUnix` in `builtin_clock.go`), RFC 7519 style: valid while `now < exp + leeway` and once `now + leeway >= nbf`. `decode` requires a key: an absent one (`undefined`/`null`, which is what an unset env var or a missing fact produces) throws rather than skipping verification. `options` is `{ leeway?: seconds, checkTime?: boolean }`, with `leeway` capped at 86400 so a millisecond value is an error; `checkTime: false` is the inspect-only opt-out, and unknown keys throw. `verify` requires the header `alg` to equal `algorithm`. `getPayload`/`getHeader` never verify.
  - **Side Effects:** None.
  - **Exceptions:** `decode` throws on any failure, including expiry. `verify` answers `false` for a malformed, mismatched, badly signed, expired or not-yet-valid token, and throws for `none` or an unsupported `algorithm`, an empty or non-string secret, bad options, or a VM with no execution start time.

- **Signature:** `@sentrie/string` - `lower`, `upper`, `trim`, `trimStart`, `trimEnd`, `startsWith`, `endsWith`, `includes`, `split`, `join`, `replace`, `replaceAll`, `equalFold`, `format`
  - **Behavior:** Go `strings` over UTF-8 code points: `split(s, "")` yields code points, trims strip Unicode White_Space, case mapping and folding are Go's simple one-to-one mappings. `replace` replaces the first occurrence only. `format` is a closed verb set (`%s %q %d %f %.Nf %v %%`, no flags or widths, precision 0 to 20) and requires the argument count to equal the verbs that consume one. `join` takes a JS array or a goja-wrapped Go slice of at most `stringMaxListLength` (2^20) elements; negative zero prints as `0` under every numeric verb. The contract is documented in `ts_src/sentrie/string.d.ts`.
  - **Side Effects:** None.
  - **Exceptions:** **Thrown**, never returned: exact argument counts, string-typed arguments with no coercion, unsupported `format` verbs or mismatched verb arguments.

- **Signature:** `@sentrie/math` - `abs`, `floor`, `ceil`, `round`, `trunc`, `min`, `max`, `pow`, `sqrt`, `clamp`
  - **Behavior:** No non-finite value in or out: arguments must be finite numbers, a NaN or infinite result throws, and negative zero is returned as 0. `round` is half away from zero (unlike JS `Math.round`). `min`/`max` take one or more numbers or a single non-empty array of at most `mathMaxArrayLength` (2^20) elements. `clamp` requires `lo <= hi`. The contract is documented in `ts_src/sentrie/math.d.ts`.
  - **Side Effects:** None.
  - **Exceptions:** **Thrown**, never returned: exact argument counts, non-number or non-finite arguments, `sqrt` of a negative, non-finite `pow` results, `clamp` with `lo > hi`.

- **Signature:** `time.now() -> number`
  - **Behavior:** Reads the `ExecutionStartTimeUnixKey` VM global and returns it, so JavaScript observes the **same pinned instant** as the policy evaluator. Falls back to `time.Now().Unix()` when the global is absent.
  - **Side Effects:** None.
  - **Exceptions:** Argument-count error if given any argument.

## 4. Operational Context & Gotchas
- **Statefulness:** Providers are stateless factories, but they are re-invoked on **every** `require` of a builtin because [[runtime.js.alias_runtime]] checks the provider branch before the cache. Two requires of the same builtin in one VM produce two distinct objects.
- **Performance/Scale Notes:** Native Go, so orders of magnitude faster than equivalent interpreted JavaScript - which is the main reason to prefer a builtin over a hand-rolled implementation in a pack. Object construction on each require is small but not free.
- **Dependencies Risk:**
  - **Failures throw; never return an error object.** In goja a value returned from a Go function is a normal result, so `return vm.NewGoError(...)` hands JavaScript a truthy Error object and a malformed CIDR reads as a match. That was the behaviour until [#111](https://github.com/sentrie-sh/sentrie/issues/111). Every site now uses `return throwErr(vm, err)`, and `TestBuiltinsNeverReturnErrorObjects` walks the AST of every `runtime/js` source file and fails if a `NewGoError`/`NewTypeError` call has any parent other than `panic(...)`, or if `ToValue` wraps a Go error. `TestErrorObjectGuardCatchesEverySpelling` keeps the guard itself honest. An uncaught throw fails the calling rule through [[runtime.modules]].
  - **Predicates answer `false` for garbage.** `net.isPrivate("junk")` and `net.isPublic("junk")` are both `false`, so a negated predicate (`!isPrivate(ip)`) is `true` for unparseable input. Positive checks are the fail-closed form.
  - **`uuid` and the `time` fallback are non-deterministic**, which is why module calls are forbidden inside derives (see [[runtime.eval_call]]) - a derive must be reproducible within an execution.
  - **The `time.now()` fallback is silent.** If the execution-start global is missing, JavaScript quietly gets wall-clock time and diverges from what the evaluator's own `now()` reports, with no signal.
  - **JWT time checks trust the execution-start global.** Module code can still overwrite `__executionStartTimeUnix` before calling `jwt.verify`, which would move the clock the checks use; making internal globals unreachable is [#154](https://github.com/sentrie-sh/sentrie/issues/154).
  - **`net` deliberately has no I/O.** It is CIDR arithmetic over `net.ParseCIDR`, not a network client. Preserving that is what keeps the sandbox claim honest.
  - **Array arguments are length-checked before allocation.** A JS array's `length` is caller-controlled (`a.length = 2**32 - 1` is free), so `join`, `min` and `max` reject lengths over 2^20 and grow their buffers while validating elements. Sizing a buffer straight from `length` would let one line of module code trigger an unrecoverable Go out-of-memory fatal error.
  - **`@sentrie/string` needs an explicit alias.** `string` is a keyword, so [[parser.use]] accepts it as a path segment but rejects it as the default alias; `use { lower } from @sentrie/string as str` is the working form.
  - **Argument counts are checked exactly**, so an extra argument is an error rather than being ignored - stricter than JavaScript convention and a likely source of surprise for module authors.
  - **`crypto` and `jwt` are security-critical surfaces.** Their failures throw like every other builtin's (#111).
  - **`crypto.sha256` returns the lowercase hex digest**, the same encoding as `hash.sha256`, and throws on a wrong argument count (#138). Before #138 it returned the input bytes with the empty-string digest appended.
