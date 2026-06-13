# Adding and changing builtins

This document is the authoritative guide for contributors adding or modifying Sentrie
built-in functions. Builtins live entirely in the [`builtins/`](.) package as **one
struct literal per function** in `Table`. There is no `init()` registration, no second
purity list, and no hand-rolled arity checks in implementations — `Precheck` is the
single enforcement point at dispatch.

**Read this file end-to-end before opening a PR.** Skipping a step (especially
`Result.Kinds`, error strings, or the strict/lenient profile) is the most common source
of review churn.

---

## Architecture (30-second overview)

```
policy call "filter(xs, f)"
        │
        ▼
runtime/eval_call.go          lookup builtins.Table["filter"]
        │                     derive guard: decl.DeriveSafe
        ▼
decl.Precheck(site, args)     arity, kind, callable-arity
        │                     undefined/null args pass through
        ▼
decl.Impl(ctx, site, args...)  real logic only
        │
        ▼
runtime.CallSite              implements builtins.Env
                              (Call, CallableArity, ExecutionStart)
```

- **`builtins/`** is a **leaf package**: it may import `box`, stdlib, and `xerr` (for
  `error()` only). It must **never** import `runtime`, `index`, or `ast`.
- **`index/`** and **`runtime/`** both import `builtins` and read `Table` / `IsDeriveSafe`.
- **TypeScript / JS modules** under `runtime/js/` are a separate registry — do not touch
  them when adding Go builtins.

---

## File layout

| File | Contents |
|------|----------|
| [`builtins.go`](builtins.go) | `Env`, `Fn`, `Decl`, `Sig`, `ParamSig`, `Precheck`, `Table`, `IsDeriveSafe`, `hofSig` helper |
| [`scalar.go`](scalar.go) | Scalar builtins (`count`, `merge`, `flatten`, …) and private helpers |
| [`collection.go`](collection.go) | Higher-order builtins (`filter`, `reduce`, …), `iterArgs`, `reduceArgs` |
| `*_test.go` | Tests; use `invoke()` + `noopEnv()` (see [Testing](#testing)) |

Pick **scalar** vs **collection** by whether the builtin invokes callables via `env.Call`.
When in doubt, put helpers next to the impl they serve.

---

## Step-by-step: add a new builtin

### 1. Choose a name and file

- Name is the map key in `Table` and the identifier used in policies (e.g. `filter`).
- Add the `Decl` variable in `scalar.go` or `collection.go` (e.g. `declMyBuiltin`).
- Register it in `Table` in [`builtins.go`](builtins.go). **Map key must equal `Decl.Name`.**

### 2. Write the `Decl` literal

Every builtin is exactly one `&Decl{...}`:

```go
var declExample = &Decl{
    Name:        "example",
    Description: "One sentence for docs and reviewers.",
    DeriveSafe:  true, // see [DeriveSafe](#derivesafe)
    Sig: Sig{
        Params: []ParamSig{ /* see below */ },
        Variadic: nil,              // or &ParamSig{...} for variadic tail (e.g. error)
        TooFewError:  "example requires 2 arguments",
        TooManyError: "example requires 2 arguments", // often same as TooFewError
        Result: ParamSig{
            Name:  "result",
            Kinds: kindList, // REQUIRED — see [Result kinds](#result-kinds)
        },
    },
    Impl: implExample,
}
```

`TestTableWellFormed` enforces: non-empty `Description`, non-nil `Impl`, `DeriveSafe`,
map key == `Name`, optional params trailing-only, callable arities only on callable kinds,
variadic params not optional, and **`Result.Kinds`** (update `expectedResultKinds` in
[`table_test.go`](table_test.go)).

### 3. Define `Sig.Params` (and optional `Variadic`)

Each `ParamSig`:

| Field | Purpose |
|-------|---------|
| `Name` | Documentation / future static checker |
| `Kinds` | Allowed `box.ValueKind`s; `nil` = any defined kind |
| `Optional` | Trailing-only optional param (`flatten` depth, `distinct` keyFn) |
| `CallableArities` | e.g. `{1,2}` — only valid when `Kinds == {ValueCallable}` |
| `OnMismatch` | Default `MismatchError`; use `MismatchUndefined` only with strong justification (today: `count` only) |
| `KindError` | **Verbatim** error string when a *defined* arg has wrong kind |
| `CallableArityError` | Verbatim string when callable arity ∉ `CallableArities` |

**Error strings are normative.** Copy them from an existing builtin or from grep of current
code; do not invent a generic template like `"example: argument 1 must be list"`. Golden
tests assert exact text.

### 4. Implement `Fn` — signature is mandatory

```go
func implExample(ctx context.Context, env Env, args ...box.Value) (box.Value, error) {
    // ...
}
```

Rules:

1. **First parameter must be `ctx context.Context`.** Runtime passes the evaluation
   context from `eval_call.go`. Do not omit or reorder parameters.
2. **Second parameter is `env Env`.** Use it for `env.Call`, `env.CallableArity`, and
   `env.ExecutionStart` — never import `runtime`.
3. **Third is variadic `args ...box.Value`.** `Precheck` has already run; do not
   re-check arity/kind/callable-arity unless [strict profile](#strict-vs-lenient-profiles)
   requires Impl-side checks for undefined/null.

**Higher-order builtins** must forward `ctx` on every `env.Call`:

```go
res, err := env.Call(ctx, fn, iterArgs(arity, item, idx))
```

Use `env.CallableArity(fn)` once per call (not per iteration if arity is loop-invariant).

### 5. Decide strict vs lenient undefined handling

See [Strict vs lenient profiles](#strict-vs-lenient-profiles). This determines what you
delete from `Impl` vs what you keep after moving checks to `Precheck`.

### 6. Set `DeriveSafe`

| Value | Meaning |
|-------|---------|
| `true` | Deterministic **within a single policy execution** (pinned `ExecutionStart` for time). Required for use inside `derive` bodies. |
| `false` | Not allowed in derives; runtime and `index` purity walker reject it. |

If you widen `Env`, reconsider `DeriveSafe` for every builtin — widening `Env` is a
deliberate review event.

### 7. Update tests

1. **`table_test.go` → `expectedResultKinds`** — add your builtin name and result kinds.
2. **Behavior tests** in `scalar_test.go` or `collection_test.go` using `invoke(t, name, env, args...)`.
3. **`TestGoldenBehavior`** — add rows for undefined/sentinel behavior if non-obvious.
4. **Arity/kind/callable errors** — table-driven cases like `TestBuiltinsCollection_ArityAndTypeErrors`.
5. Run `go test -race ./builtins/...` and `golangci-lint run ./builtins/...`.

You do **not** need to change `runtime/eval_call.go` for a new name — dispatch is table-driven.

### 8. Index / static metadata consumers

- `index/derive_purity.go` uses `builtins.IsDeriveSafe(name)` — no change needed if
  `DeriveSafe: true`.
- Future validate-time checking (Issue B) reads `Table[name].Sig` — populate `Result.Kinds`
  and param `Kinds` correctly now.

---

## `Precheck` behavior (what Impl can assume)

Dispatch in `runtime/eval_call.go`:

```go
if handled, v, err := decl.Precheck(site, args); handled || err != nil {
    return v, err
}
return decl.Impl(ctx, site, args...)
```

`Precheck` algorithm:

1. **Arity** — `min` = non-optional params; `max` = `len(Params)` or unbounded if `Variadic != nil`.
2. **Kind** — for each arg: if `undefined` or `null`, **skip** (pass through to `Impl`).
   If `Kinds` is set and kind ∉ `Kinds`: apply `OnMismatch` or `KindError`.
   Callable params with wrong kind: `"expected callable, got <kind>"` when `KindError` is empty.
3. **Callable arity** — if `CallableArities` set: `env.CallableArity` + `CallableArityError`.

**Delete from `Impl`:** hand-rolled `len(args)`, list-kind checks, and callable-arity
checks for **lenient** builtins once `Precheck` covers them.

**Keep in `Impl`:** value-level rules `Precheck` cannot express (e.g. `flatten` depth ≥ 0,
`normalise_list` nesting depth, `error` format-string validation).

---

## Strict vs lenient profiles

| Profile | Examples | Undefined/null on collection param | Impl kind checks |
|---------|----------|-----------------------------------|------------------|
| **Lenient** | `filter`, `any`, `count`, `flatten` | Sentinel result (`[]`, `false`, `undefined`, …) | Remove duplicate checks for *defined* values; keep sentinel logic |
| **Strict** | `merge`, `distinct` | Still errors (via `DictValue`/`ListValue` in Impl) | **Retain** `DictValue`/`ListValue` in Impl; `Precheck` adds early path for defined non-kind |

For **strict** builtins, `Precheck` skips kind check on undefined → `Impl` must still
reject undefined with the same message as before (e.g. `"first argument is not a dict"`).

Golden tests: `merge(undefined, dict)`, `merge(bad, bad)` (arg-0 error first), `distinct(undefined)`.

---

## Result kinds

Populate `Sig.Result.Kinds` for every entry (validate-time propagation):

| Result kinds | Builtins |
|--------------|----------|
| `{bool}` | `all`, `any` |
| `{list}` | `as_list`, `filter`, `collect`, `flatten`, `flatten_deep`, `normalise_list`, `distinct` |
| `{number}` | `count`, `now` |
| `{dict}` | `merge` |
| `nil` (any) | `error`, `first`, `reduce` |

---

## `Env` capabilities

```go
type Env interface {
    Call(ctx context.Context, fn box.Value, args []box.Value) (box.Value, error)
    CallableArity(fn box.Value) (int, error)
    ExecutionStart() time.Time
}
```

| Method | Use |
|--------|-----|
| `Call` | Invoke predicate/transform/reducer callables (HOFs) |
| `CallableArity` | Usually only needed inside `Precheck`; HOFs call once per invocation |
| `ExecutionStart` | `now()` — `box.Number(float64(env.ExecutionStart().UnixMilli()))` |

At runtime, `Env` is `*runtime.CallSite`. In tests, use `noopEnv()` or a custom fake
(see [`test_helpers_test.go`](test_helpers_test.go)).

---

## Allowed imports

| Allowed | Notes |
|---------|-------|
| `github.com/sentrie-sh/sentrie/box` | Values and kinds |
| `github.com/sentrie-sh/sentrie/xerr` | `error()` builtin only (`ErrInjected`) |
| `context`, `fmt`, `slices`, `time`, … | stdlib |

| **Forbidden** | Reason |
|---------------|--------|
| `runtime`, `index`, `ast` | Import cycle; use `Env` instead |

---

## Testing

### `invoke()` — mirrors production dispatch

```go
func invoke(t *testing.T, name string, env Env, args ...box.Value) (box.Value, error) {
    decl := Table[name]
    handled, v, err := decl.Precheck(env, args)
    if handled || err != nil {
        return v, err
    }
    return decl.Impl(t.Context(), env, args...)
}
```

Always pass a real `context.Context` via `t.Context()` in tests (same as runtime).

### Checklist for a new builtin

- [ ] `Table` key == `Name`
- [ ] `TestTableWellFormed` passes (incl. `expectedResultKinds`)
- [ ] Arity too few / too many → exact `TooFewError` / `TooManyError`
- [ ] Wrong kind → exact `KindError` or `MismatchUndefined` behavior
- [ ] Wrong callable arity → exact `CallableArityError`
- [ ] Undefined/null sentinel behavior (if any) in `TestGoldenBehavior` or dedicated test
- [ ] HOF: `ctx` forwarded to every `env.Call`
- [ ] `go test -race ./builtins/...`

### Injecting a test builtin into `Table`

Rare (e.g. memoization integration tests in `runtime/`). Use a full `Decl` with permissive
`Sig` (e.g. variadic `any`) so `Precheck` does not block; restore `Table` in `defer`.
Must flow through real `eval_call` dispatch, not bypass it.

---

## Changing an existing builtin

1. Edit **only** its `Decl` literal and `impl*` function in `builtins/`.
2. If arity/kind/callable rules change, update `Sig` and **grep-verify** error strings.
3. If behavior changes, update golden tests — byte-for-byte for errors, value-for-value for results.
4. If `DeriveSafe` flips to `false`, expect derive purity and runtime derive-guard failures in tests.
5. Do **not** add a second registration site or `init()` hook.

---

## Common mistakes

| Mistake | Consequence |
|---------|-------------|
| Generic invented error strings | CI / golden test failure |
| Deleting `DictValue`/`ListValue` from strict `merge`/`distinct` | `undefined` stops erroring |
| Forgetting `expectedResultKinds` entry | `TestTableWellFormed` fails |
| `CallableArities` on non-callable param | `TestTableWellFormed` fails |
| Importing `runtime` for `CallSite` | Build failure (cycle) |
| HOF without `env.Call(ctx, ...)` | Callable runs without cancellation / wrong frame |
| `Impl` without `ctx context.Context` first | Does not satisfy `Fn`; won't compile |
| Using `MismatchUndefined` casually | Hides errors; only `count` uses it today |

---

## Reference: current builtins (15)

| Name | File | Arity | Notes |
|------|------|-------|-------|
| `all`, `any` | collection | 2 | HOF; bool result |
| `as_list` | scalar | 1 | |
| `collect` | collection | 2 | HOF; list result |
| `count` | scalar | 1 | `MismatchUndefined` for wrong kind |
| `distinct` | collection | 1–2 | strict; optional keyFn |
| `error` | scalar | 1+ | variadic; never returns |
| `filter`, `first` | collection | 2 | HOF |
| `flatten`, `flatten_deep` | scalar | 1–2 / 1 | depth optional |
| `merge` | scalar | 2 | strict dict |
| `normalise_list` | scalar | 1 | |
| `now` | scalar | 0 | `ExecutionStart().UnixMilli()` |
| `reduce` | collection | 3 | reducer arity 2 or 3 |

---

## Related code (outside `builtins/`)

| Location | Role |
|----------|------|
| [`runtime/eval_call.go`](../runtime/eval_call.go) | Builtin dispatch: `Precheck` → `Impl` |
| [`runtime/builtin_call.go`](../runtime/builtin_call.go) | `CallSite` implements `Env` |
| [`index/derive_purity.go`](../index/derive_purity.go) | `builtins.IsDeriveSafe` |
| [`index/builtins_registry_test.go`](../index/builtins_registry_test.go) | Proves index can import `Table` |

Do not modify dispatch unless fixing a bug; new builtins are table-only.
