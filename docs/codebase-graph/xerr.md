---
id: xerr
type: System / Package
language: Go
file_path: xerr/
tags: error-taxonomy, diagnostics, sentinel-errors, failure-modes
---

# Node: Xerr (Structured Error Taxonomy)

## 1. Architectural Role & Intent
`xerr` is the centralized error taxonomy for Sentrie, defining typed error structs and sentinel roots for every failure category the engine can raise: static index/validation failures, policy resolution failures, builtin arity and argument-kind violations, module/import resolution failures, and runtime panics. It exists so that callers can classify failures with `errors.Is`/`errors.As` against stable sentinels instead of matching error strings, and so that span-anchored diagnostics carry a `tokens.Range` all the way to the CLI and HTTP surfaces. It is organized by lifecycle phase across `index.go` (static), `policy.go` (resolution), `builtin_validate.go` (builtin contracts), `runtime.go` (execution), and `generic.go` (shared).

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `xerr` | `LAYERED_ON` | [[tokens]] | Builtin-validation and conflict errors embed `tokens.Range` to render source spans. |
| [[ast]] | `LAYERED_ON` | [[xerr]] | AST construction/inspection reports malformed-node conditions. |
| [[index.package]] | `LAYERED_ON` | [[xerr]] | Static validation emits `ErrIndex`-rooted failures for shape, rule, and namespace problems. |
| [[builtins]] | `LAYERED_ON` | [[xerr]] | Builtin implementations raise arity and argument-kind errors carrying call-site spans. |
| [[runtime]] | `LAYERED_ON` | [[xerr]] | Evaluation raises fact-resolution, recursion, module-invocation, and injected errors. |
| [[cmd]] | `CALLS` | [[xerr]] | CLI classifies failures to choose exit disposition and error rendering. |
| [[api]] | `CALLS` | [[xerr]] | HTTP layer maps error categories onto RFC 7807 problem-detail responses. |

## 3. Interface Contracts & Public Surface

- **Signature:** `ErrIndex` (sentinel `error`)
  - **Behavior:** Root sentinel for all static index construction and validation failures. Every static error wraps it, so `errors.Is(err, xerr.ErrIndex)` cleanly separates compile-time from run-time failures.
  - **Side Effects:** None.
  - **Exceptions:** N/A (sentinel).

- **Signature:** Policy header sentinels — `ErrPolicyMetadataContiguous`, `ErrPolicyFactAfterUse`, `ErrPolicyInvalidVersion`, `ErrPolicyEmptyTitle`, `ErrPolicyEmptyTagKey`
  - **Behavior:** Enforce the language's policy-header grammar rules: metadata (`title`/`description`/`version`/`tag`) must form one contiguous block at the top, `fact` statements must precede `use` statements, and `version` must be valid SemVer. All wrap `ErrIndex`.
  - **Side Effects:** None.
  - **Exceptions:** Callers are expected to add location context at the call site via `fmt.Errorf("at %s: %w", span, err)` — the sentinels themselves are location-free.

- **Signature:** `ErrBuiltinCallArity(at: tokens.Range, message: string) -> error` / `ErrBuiltinArgKind(...)` / `ErrBuiltinCallableArity(...)`
  - **Behavior:** Span-anchored builtin contract violations detected at validate time (not run time), backed by `BuiltinCallArityError`, `BuiltinArgKindError`, and `BuiltinCallableArityError`. The third covers higher-order-function callable arity, e.g. a lambda passed to `map` with the wrong parameter count.
  - **Side Effects:** None.
  - **Exceptions:** All three implement `Unwrap()` for chain traversal.

- **Signature:** `ErrPolicyNotFound(fqn)` / `ErrRuleNotFound(fqn)` / `ErrShapeNotFound(name)` / `ErrNamespaceNotFound(name)` / `ErrNotExported(fqn)`
  - **Behavior:** Resolution failures keyed by fully-qualified name. `ErrNotExported` distinguishes "exists but is not part of the public surface" from "does not exist", which matters for the [[api]] decision endpoint.
  - **Side Effects:** None.
  - **Exceptions:** N/A (constructors).

- **Signature:** `ErrRequiredFact(name)` / `ErrUnresolvableFact(name)`
  - **Behavior:** Distinguish a declared `fact` that was never supplied from one that could not be resolved to a value. Drives the `Unknown` decision path rather than a hard deny.
  - **Side Effects:** None.
  - **Exceptions:** N/A.

- **Signature:** `ErrInfiniteRecursion(stack: []string) -> error`
  - **Behavior:** Raised when rule/derive evaluation re-enters itself; carries the offending call stack for diagnosis.
  - **Side Effects:** None.
  - **Exceptions:** N/A.

- **Signature:** `ErrImportResolution(module, fn)` / `ErrModuleInvocation(module, fn)`
  - **Behavior:** Separate "the imported symbol could not be found" from "the module call itself failed", i.e. the TypeScript/JS boundary in [[runtime.js]].
  - **Side Effects:** None.
  - **Exceptions:** N/A.

- **Signature:** `ErrInjected(format: string, args: ...any) -> error`
  - **Behavior:** The user-facing escape hatch — produced when policy source calls the language's `error` function, letting policy authors raise deliberate, described failures.
  - **Side Effects:** None.
  - **Exceptions:** N/A.

- **Signature:** `ErrShapeValidation(msg)` / `ErrInvalidType(got, expected)` / `ErrInvalidInvocation(reason)` / `ErrConflict(what, where, with: tokens.Range)`
  - **Behavior:** Type and shape contract violations. `ErrConflict` is notable for carrying **two** spans (the declaration and the conflicting redeclaration), enabling paired diagnostics.
  - **Side Effects:** None.
  - **Exceptions:** N/A.

- **Signature:** `ErrRuntimePanic` (sentinel, `*RuntimePanic`) / `ErrNotImplemented` (sentinel, `*NotImplementedError`)
  - **Behavior:** Recovered-panic marker and unimplemented-path marker; the former is how the evaluator converts a Go panic into a policy-level failure instead of crashing the process.
  - **Side Effects:** None.
  - **Exceptions:** `NotImplementedError` implements `Unwrap()`.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless. Sentinels are package-level immutable values; constructors return fresh error instances.
- **Performance/Scale Notes:** Error construction uses `fmt.Errorf` wrapping and is therefore allocation-bearing — acceptable on failure paths but should never appear in a successful evaluation loop. Sentinel comparison via `errors.Is` walks the wrap chain, so deeply nested wraps cost proportionally.
- **Dependencies Risk:** No runtime failure domain of its own. The architectural risks are: (1) the span-anchored sentinels intentionally omit location, so **any caller that forgets the `fmt.Errorf("at %s: %w", span, err)` wrap produces a diagnostic with no file/line**, which is the most common source of unhelpful Sentrie errors; (2) both [[cmd]] and [[api]] classify errors by these categories, so introducing a new failure mode without rooting it under an existing sentinel (`ErrIndex` in particular) causes it to fall through to a generic 500 / generic CLI error rather than a precise one.
