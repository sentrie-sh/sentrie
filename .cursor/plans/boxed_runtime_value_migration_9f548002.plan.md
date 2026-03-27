---
name: Boxed Runtime Value Migration
overview: "Implement issue #64 by introducing `runtime.Value` as the evaluator’s internal value carrier, migrating runtime internals away from pervasive `any`, and preserving API/JSON/JS compatibility via explicit boundary conversions."
todos:
  - id: add-runtime-value
    content: Implement `runtime.Value` primitives and conversion tests.
    status: pending
  - id: migrate-decision-context
    content: Migrate `Decision`, attachments, and execution context storage to boxed values.
    status: pending
  - id: convert-eval-core
    content: Convert `eval(...)` and core expression evaluators to return/use `Value`.
    status: pending
  - id: convert-higher-order-eval
    content: Migrate quantifier/reduce/call paths and memoization argument flow.
    status: pending
  - id: update-executor-import-boundaries
    content: Box/unbox at executor/import boundaries while keeping public contracts stable.
    status: pending
  - id: stabilize-tests-compat
    content: Update runtime/API tests and verify JSON/trace/CLI compatibility.
    status: pending
  - id: retire-legacy-values-go
    content: Remove `runtime/values.go` after migrating required helpers and `Undefined` semantics into `runtime.Value`.
    status: pending
isProject: false
---

# Implement `runtime.Value` Across Evaluator Core

If merged, this plan will replace pervasive runtime-internal `any` usage with a boxed `runtime.Value` model while keeping external API and JSON behavior stable.

## Scope and goals

- Introduce a boxed runtime value algebra in [runtime/value.go](runtime/value.go) with explicit kinds (`undefined`, `null`, `bool`, `number`, `string`, `trinary`, `list`, `map`, `object`) and conversion helpers (`FromAny`, `Any`).
- Move evaluator internals to `runtime.Value` return/transport types, including facts/locals, expression evaluation, decisions, attachments, and imports.
- Keep `any` only at system boundaries (fact ingress, trace payloads, JSON/API output, JS interop).

## Issue #64 reference samples

- Use the sample code in [Issue #64](https://github.com/sentrie-sh/sentrie/issues/64) as architectural reference, not as a strict patch script.
- Treat the issue snippets as guidance for target APIs and semantics in these files:
  - `runtime/value.go` (boxed value kinds, constructors, accessors, `FromAny`, `Any`, JSON behavior)
  - `runtime/decision.go` (`Decision.Value` and attachments boxed as `Value`)
  - execution context shape (`facts`/`lets` boxed)
  - evaluator handlers (`eval*` functions returning `Value`)
  - executor/import glue (`FromAny` on ingress, `Any()` on boundary egress)
- Preserve current repo-specific behavior where needed (errors, trace payload shape, module integration details), even if exact code in the issue sample differs.
- During implementation review, compare each migrated area against issue intent and document deliberate deviations.

## Phased implementation plan

### 1) Add boxed value primitives and tests

- Gating rule: start with tests for [runtime/value.go](runtime/value.go) and do not proceed to later phases until that file reaches 100% line coverage.
- Create [runtime/value.go](runtime/value.go) implementing:
  - `ValueKind` enum + `String()`
  - `Value` struct and constructors: `Undefined`, `Null`, `Bool`, `Number`, `String`, `Trinary`, `List`, `Map`, `Object`
  - Accessors: `Kind`, `IsUndefined`, `IsNull`, `IsValid`, typed `*Value()` getters
  - Bridging: `FromAny(any) Value` and `(Value).Any() any`
  - JSON/String behavior: `MarshalJSON`, `String`
- Add tests in [runtime/value_test.go](runtime/value_test.go): scalar round-trips, `undefined`/`null` distinctions, recursive list/map conversions, object escape-hatch behavior, and JSON output parity.
- Include explicit assertions for all value kinds, boundary conversions, and negative/default branches so [runtime/value.go](runtime/value.go) reaches 100% line coverage before phase 2 starts.
- Migrate legacy helpers currently in [runtime/values.go](runtime/values.go) (`IsUndefined`, `Undefined`, numeric conversion helpers) into `runtime.Value` APIs with equivalent or stricter semantics.

### 2) Migrate decision and execution context storage

- Update [runtime/decision.go](runtime/decision.go):
  - `Decision.Value` from `any` -> `Value`
  - `DecisionAttachments` from `map[string]any` -> `map[string]Value`
  - `DecisionOf` to accept boxed values
  - `MarshalJSON` to serialize through `Value.Any()` for compatibility
- Update [runtime/exec_ctx.go](runtime/exec_ctx.go):
  - fact/local storage and getter/setter APIs to use `Value`
  - keep type-validation/fact-injection control flow intact

### 3) Convert evaluator signatures and expression handlers

- Change root signature in [runtime/eval.go](runtime/eval.go):
  - `eval(...) (any, *trace.Node, error)` -> `eval(...) (Value, *trace.Node, error)`
- 3a. Migrate core expression handlers to boxed flow and get green tests before touching higher-order evaluators:
  - literals and collections: [runtime/eval.go](runtime/eval.go)
  - identifiers/locals/facts: [runtime/eval_ident.go](runtime/eval_ident.go)
  - unary/infix/cast/ternary: [runtime/eval_unary.go](runtime/eval_unary.go), [runtime/eval_infix.go](runtime/eval_infix.go), [runtime/eval_cast.go](runtime/eval_cast.go), [runtime/eval_ternary.go](runtime/eval_ternary.go)
  - block/let scoping: [runtime/eval_block.go](runtime/eval_block.go)
  - field/index access: [runtime/eval_field_access.go](runtime/eval_field_access.go), [runtime/eval_index_access.go](runtime/eval_index_access.go)
- 3b. Freeze core semantics and verify parity on literals/operators/access before proceeding to higher-order evaluators.
- Continue setting trace results via `n.SetResult(v.Any())` to preserve trace JSON shape.
- Define and document evaluator treatment of `undefined` vs `null` at each operator/access boundary before progressing to higher-order evaluators.

### 4) Migrate higher-order evaluator paths and memoized calls

- Update quantifier/reduce/call paths to boxed signatures and container handling:
  - [runtime/eval_quantifier.go](runtime/eval_quantifier.go)
  - [runtime/eval_reduce.go](runtime/eval_reduce.go)
  - [runtime/eval_call.go](runtime/eval_call.go)
- Keep builtin/module call boundaries variadic `...any` where required, but box/unbox at call sites.
- Add an explicit memoization compatibility checkpoint:
  - capture pre-migration hash behavior for representative inputs (`undefined`, `null`, booleans, numeric edge values, nested lists, maps with differing key order)
  - compare post-migration behavior and decide compatibility policy (preserve exact hash keys vs accept key changes with documented cache behavior)
  - add regression tests that lock the chosen behavior in [runtime/eval_call_test.go](runtime/eval_call_test.go) (or nearest call-path test file)

### 5) Convert executor/import glue and boundary boxing

- In [runtime/executor.go](runtime/executor.go):
  - box incoming `facts map[string]any` exactly once (`FromAny`) during injection
  - keep public `Executor` interface unchanged initially for compatibility
  - ensure attachments collected as `map[string]Value`
  - keep attachments boxed internally and unbox only at external serialization boundaries
- In [runtime/imports.go](runtime/imports.go):
  - keep imported facts wire format as `map[string]any` at the external call boundary
  - convert internal evaluator values to `Any()` only when invoking nested exec
  - rebox JS/module return values through one funnel before they re-enter evaluator internals

### 6) Align validators and utility helpers with boxed collections

- Refactor value/type helper logic and typeref validators:
  - [runtime/typeref.go](runtime/typeref.go)
  - [runtime/typeref_list.go](runtime/typeref_list.go)
  - [runtime/typeref_map.go](runtime/typeref_map.go)
  - [runtime/typeref_document.go](runtime/typeref_document.go)
- Invariant: validator inputs must be normalized once through a single funnel (`FromAny`/`Any` bridge), and validator implementations must not reintroduce scattered type-switch logic.
- Strategy: keep validator entrypoints compatible, but normalize through `FromAny`/`Any` in one place to avoid duplicated type-switches.
- Remove [runtime/values.go](runtime/values.go) once all call sites are migrated to `runtime.Value` APIs and no legacy `any` helpers remain.

### 7) Verify boundary compatibility (API, CLI, JS)

- Validate no API contract break for decision endpoints in [api/handle_decision.go](api/handle_decision.go).
- Ensure CLI output still renders attachments recursively in [cmd/exec.go](cmd/exec.go), adapting formatting helpers if they currently assume only `[]any`/`map[string]any`.
- Keep JS interop boundary conversion explicit in [runtime/modules.go](runtime/modules.go).

### 8) Test and stabilization pass

- Update/add runtime tests that currently assert `[]any`/`map[string]any` internals:
  - [runtime/builtins_test.go](runtime/builtins_test.go)
  - [runtime/eval_distinct_test.go](runtime/eval_distinct_test.go)
  - typeref tests under [runtime](runtime)
- Run targeted runtime + API tests first, then full suite.
- Confirm serialized `Decision`/trace payloads remain backward-compatible.
- Add a benchmark gate before this phase is complete:
  - allocations/op on representative pure-runtime rule sets
  - bytes/op on the same rule sets
  - end-to-end exec latency on a medium policy pack
  - one JS-heavy path and one pure-runtime path for comparison
  - record before/after results and require no unexpected regression outside agreed tolerance

## Key migration invariants

- Evaluator hot path should no longer transport raw `any` values internally.
- `undefined` and `null` remain distinct runtime states in the boxed model; coercion between them must be explicit and intentional.
- Numeric runtime semantics remain one `number` kind backed by `float64`.
- Internal lists must be `[]Value` and internal maps must be `map[string]Value` throughout evaluator/runtime internals.
- External API shape remains stable by unboxing at boundaries.
- `object` kind is retained as an escape hatch but not used as default collection/scalar representation.
- Do not optimize `ValueObject` early; treat it as a compatibility escape hatch first, and only optimize after boxed core semantics are stable.
- Numeric compatibility must stay explicit for equality and modulo (`%`) semantics under `float64`-backed `number`.

## Review focus

- Correctness of `Value` equality/coercion semantics in infix/cast operations.
- Correctness of recursive list/map equality semantics after boxed migration.
- Correctness of `undefined` propagation and comparison semantics across unary/infix/access/cast paths.
- Boundary discipline (`FromAny` only on ingress, `Any()` only on egress).
- No trace/API output regressions.
- No unexpected cache-key churn in memoized calls after boxed migration.
- Complete removal of legacy `runtime/values.go` without semantic drift in undefined handling or numeric coercion behavior.
- Enforcement of phase gate: [runtime/value.go](runtime/value.go) remains at 100% file coverage when phase 1 is complete.

