---
name: Derive feature 88
overview: "Implement sentrie-sh/sentrie#88 and website#34. Open questions are resolved: Elvis (null+undefined), PureBuiltins whitelist beside registration, derive shadows builtins by resolution order, DetachedChildContext + deriveCallable, purity walker scope stack for let, policy derive only in body phase, grammar in lockstep with parser. Spike gates Phases B–F with three green areas (FQN callee, Elvis, typed param parse) before merge."
todos:
  - id: spike-fqn-callee
    content: "Spike 1: FQN callee vs division — lock `a / b` as division and `com/example/ns/name(x)` as qualified derive call; document rule; tests green before FQN AST/parser work."
    status: pending
  - id: spike-elvis
    content: "Spike 2: Elvis `?:` — null+undefined trigger RHS; five eval tests (null, undefined, 0, \"\", false); green before wiring Elvis into optional-param paths."
    status: pending
  - id: spike-typed-params
    content: "Spike 3: Typed lambda param parse — AST round-trip for (a: number, b?: string), (a, b?: string); parse error for (a?, b: number). Green before Phases B–F merge."
    status: pending
  - id: lexer-ast-parser
    content: "Post-spike: KeywordDerive, Lambda AST extensions, DeriveStatement/ExportDeriveStatement, lambda/derive parsing, export dispatch; grammar.peg + grammar.ebnf same commit/PR as parser."
    status: pending
  - id: index-validate
    content: "Post-spike: Index derives, collisions (no builtin prohibition), Resolve/Verify export, derive DAG, purity walker (PureBuiltins + TS + scope stack for let + callable return rules)."
    status: pending
  - id: runtime
    content: "Post-spike: getTarget derive→builtin→TS, DetachedChildContext, deriveCallable (not lambdaCallable), arity/padding/typed checks, Elvis eval."
    status: pending
  - id: tests
    content: "Post-spike: Full matrix (policy derive before fact parse error, internal let in derive purity, etc.)."
    status: pending
  - id: website-docs
    content: "Website #34 — Derives, lambdas, collections, getting-started, TypeScript boundary, nav."
    status: pending
isProject: false
---

# Plan: `derive` and typed lambdas (sentrie-sh/sentrie#88) + docs (sentrie-sh/website#34)

## Sources of truth

- Feature spec: [sentrie-sh/sentrie#88](https://github.com/sentrie-sh/sentrie/issues/88).
- Documentation scope: [sentrie-sh/website#34](https://github.com/sentrie-sh/website/issues/34).

## Locked decisions (apply before spike; spike tests reflect these)

### 1. Elvis `?:` semantics

- **`a ?: b`** evaluates to **`b`** when **`a` is `null` or `undefined`**; otherwise **`a`**.
- Rationale: facts/shape fields often use `null`; omitted optional params use `undefined`; authors should not need to distinguish for defaults.
- **Not** triggers: `0`, `""`, `false` — LHS short-circuits (RHS not evaluated for those cases in the usual lazy-Elvis sense; lock in tests).

**Spike tests (must be green before Elvis is used for optional-param defaults elsewhere):**

| Case | Expected |
|------|----------|
| `null ?: rhs` | `rhs` |
| `undefined ?: rhs` | `rhs` |
| `0 ?: rhs` | `0` |
| `"" ?: rhs` | `""` |
| `false ?: rhs` | `false` |

### 2. Builtins in derives — explicit whitelist (`PureBuiltins`)

- Maintain an explicit **`PureBuiltins` set** (or `pure` flag per builtin registration) **alongside** builtin registration in [`runtime/builtins.go`](sentrie/runtime/builtins.go) or a **companion file colocated with registration** — not inside the purity walker.
- The **authoritative builtin name set** is the same as runtime registration; **do not duplicate** a second string list that can drift (e.g. derive `PureBuiltins` from the same `Builtins` map keys at init, or attach a `pure bool` to each registration entry — one source of truth).
- The **purity walker consults** this whitelist for derive bodies. Calls to builtins **not** on the whitelist produce a **specific** error, e.g.:

  `builtin 'X' is not permitted inside a derive — it cannot guarantee deterministic output within a single policy execution`

- **`now()` is explicitly whitelisted** (`PureBuiltins` / `pure: true` on its registration). It returns **`createdAt`**, not wall-clock time per call — so it is **deterministic within a single policy execution** by design (same justification as inheriting `createdAt` in **`DetachedChildContext`**, §4). Treat **`now()` as the canonical whitelisted example** in docs and comments — **not** as the canonical excluded builtin. When documenting builtins that fail the check, contrast them with **`now()`** as the intended allowed case. Implementers should add **`now()` as the first** (or first-documented) whitelist entry so the connection stays obvious during reviews.
- Future builtins that cannot meet the determinism-within-one-execution bar stay **off** the whitelist and remain unusable from derives.

### 3. Derive names vs builtins — shadowing allowed

- **Resolution order** (`derive` → builtin → TypeScript) is the contract; **no** prohibition and **no** warning when a derive name collides with a builtin.
- A derive named e.g. `filter` wins in scope over the builtin; new builtins remain non-breaking for existing policies that shadow them.
- **Do not** add derive-vs-builtin checks to [`index/namespace.go`](sentrie/index/namespace.go) `checkNameAvailable` (or equivalent): builtins are not a naming conflict surface for derives.
- **Out of scope for this work:** CLI `sentrie fix` / lint to flag redundant shadowing (post-v1).

### 4. Runtime purity enforcement — `DetachedChildContext` + `deriveCallable` (final spec)

Implementation lives in [`runtime/exec_ctx.go`](sentrie/runtime/exec_ctx.go) (and callable construction beside today’s [`runtime/callable.go`](sentrie/runtime/callable.go) / eval paths). **`lambdaCallable` must not be reused for derives** as-is: it captures the parent EC by reference (late-bound), exposing full ambient context. Anonymous lambdas keep existing `lambdaCallable` behavior unchanged.

#### `DetachedChildContext`

```go
func (ec *ExecutionContext) DetachedChildContext() *ExecutionContext {
    ec.rwmu.RLock()
    defer ec.rwmu.RUnlock()

    ec2 := NewExecutionContext(ec.policy, ec.executor)
    ec2.refStack = slices.Clone(ec.refStack)
    ec2.createdAt = ec.createdAt
    return ec2
}
```

**Field-by-field rationale**

| Field | Value | Reason |
|-------|-------|--------|
| `parent` | `nil` | Isolation: no bubbling to caller facts, lets, locals, or modules. |
| `policy` | inherited | Typed param / return validation needs policy + shape context; `SetLocal` may consult `ec.policy.Rules` for locals routing. |
| `executor` | inherited (via `NewExecutionContext`) | Eval engine — not ambient policy data. |
| `refStack` | `slices.Clone(ec.refStack)` | `NewExecutionContext` starts `refStack` as zero len/cap; `copy` into that dst would copy nothing. `slices.Clone` allocates the right length — same pattern as `AttachedChildContext`; only correct option. |
| `createdAt` | inherited | `now()` is tied to `createdAt`, not wall clock per call (see **`now()` on `PureBuiltins`**, §2); a fresh `time.Now()` in `NewExecutionContext` would make `now()` inside a derive disagree with `now()` in the enclosing rule — silent correctness bug. |
| `facts` | `nil` (from `NewExecutionContext`) | Derives must not read caller facts; with `parent == nil` there is no bubble-up path anyway. |
| `locals` | fresh empty map | Params are seeded by `deriveCallable` after construction (`SetLocal(..., force=true)`). |
| `lets` | fresh empty map | Body `let` bindings stay inside the derive evaluation. |
| `modules` | fresh empty map | No module bindings in derives. |

**`SetLocal` nil-policy guard:** do **not** add one — with `policy` inherited, `ec.policy.Rules[name]` in `SetLocal` remains safe.

#### `deriveCallable` — param seeding after construction

`DetachedChildContext` returns a context with **empty** locals. **`deriveCallable`** constructs it from the **caller** EC, then seeds each lambda param with `SetLocal(name, args[i], true)` so routing bypasses fact/let/rule paths (none exist on the detached context). **Optional-parameter padding** (missing args → `box.Undefined()`) is done **in the caller** before building `args`, **not** inside `deriveCallable`.

Illustrative **invoke** flow (adapt to repo `Callable` / `evalBlock` signatures):

```text
on invoke(callerEC, args):
  ec := callerEC.DetachedChildContext()
  defer ec.Dispose()
  for i, name := range lambda.Params:
    ec.SetLocal(name, args[i], true)   // force=true
  return evalBlock(..., ec, ..., lambda.Body)
```

#### Cross-derive calls inside a derive body — **define-site closure** (locked)

**Decision:** **`deriveCallable` closes over the derive-resolution surface visible when that derive is constructed** (namespace + policy derives in scope at index / bind time — a **registry snapshot**), not over whatever the **caller** rule happens to have in scope at runtime.

- **`DetachedChildContext` does not** take a derive registry as an extra constructor argument; it stays the lean EC described in the table above.
- **Rejected for this work:** **inject at construction** (passing the registry only into `DetachedChildContext`) — valid alternative, but extra plumbing and weaker author predictability.

**Why define-site:** a derive’s behaviour is fixed by what was visible **when it was defined**, not by derives added to the calling policy later; matches author mental model and avoids leaking caller EC through late-bound scope. Implementing callee resolution for `other_derive(...)` inside the body uses the **captured registry** + normal `DetachedChildContext` per inner invoke (fresh detached EC, args only) — still **no** caller facts/modules.

**Nested inner derive:** each inner `deriveCallable` invocation again uses `DetachedChildContext()` from the **current** detached EC only to clone policy/executor/refStack/createdAt; resolution for the inner target still goes through **that** callable’s own define-site closure if it is a derive, or builtins/whitelist as applicable.

#### `PushRefStack` / refStack identity for derives (**FQN required**)

- Cycle detection uses `PushRefStack(uniqueID string)` (see [`runtime/exec_ctx.go`](sentrie/runtime/exec_ctx.go)). For **derives**, `uniqueID` must be the derive’s **fully qualified name** (namespace path + derive segment, globally unambiguous — same form as index / exported FQN resolution), **never** the bare derive name alone.
- **Why:** two namespaces can each define `derive check`; if the ref stack used only `check`, a call from one namespace’s derive into the other’s `check` could be mistaken for self-recursion and raise a **false** cycle error.
- **Where:** inner **`deriveCallable`** invoke path: **push** callee FQN before evaluating the callee body, **pop** after return; same for **cross-namespace** calls to an exported derive. Cloned `refStack` in `DetachedChildContext` preserves the chain across nested detached contexts.

### 5. Purity walker — scope stack including `let`

- The walker maintains a **local scope stack** (not only param names): initialise with derive params; **push** identifiers introduced by **`let`** inside the derive body as the walker descends; only flag identifiers that **escape** the accumulated scope.
- **Test:** derive with internal `let` (e.g. `doubled = x * 2`, `yield doubled + 1`) must **not** emit a false positive purity violation.

### 6. Policy `derive` placement — explicit parser constraint

- **`derive` is valid only in the policy body phase** — after metadata, `fact`, and `use` (same phase class as `let` / `rule` / `shape` body items today via [`index/policy.go`](sentrie/index/policy.go) phases).
- **`derive` before `fact`** (or in metadata / `use` phase) → **parse error** with messaging **consistent** with existing misplaced-statement errors (e.g. same style as `latePolicyHeaderErr` / similar).
- **Parser test:** `derive` placed before `fact` declarations → error (prevents confusing runtime-only failure).

### 7. Grammar files — lockstep with parser

- Update [`grammar/grammar.peg`](sentrie/grammar/grammar.peg) and [`grammar/grammar.ebnf`](sentrie/grammar/grammar.ebnf) in the **same commit** as parser changes for each construct (minimum: **same PR**). Do not let grammar drift from the hand-written parser.
- Before opening the PR, confirm CI (if any) validates grammar vs parser expectations for the **new derive and typed-param productions**.

---

## Spike (gates Phases B–F — nothing merges until all three areas green)

The spike produces **green tests** for **exactly** these three areas. **Phases B through F do not merge** until all three are locked.

### Spike area 1 — FQN callee vs binary division

- **Minimum cases:**
  - `a / b` → **division** (two identifiers, no call).
  - `com/example/ns/name(x)` → **qualified derive call** (disambiguation rule chosen and **documented in plan or code comment** next to the disambiguation implementation).
- Lock with tests **before** broader FQN AST nodes or full parser integration proceeds (implementation approach follows the locked rule).

### Spike area 2 — Elvis

- Implement enough of lexer/parser/eval for `?:` to run the **five cases** in section 1 above. All green **before** Elvis is wired into optional-param / derive examples broadly.

### Spike area 3 — Typed param parsing

- Round-trip through parser with correct AST fields:
  - `(a: number, b?: string)` → typed required + typed optional.
  - `(a, b?: string)` → untyped required + typed optional.
  - `(a?, b: number)` → **parse error** (optional before required).

---

## Current codebase anchors (unchanged summary)

- Lambdas: [`ast/lambda.go`](sentrie/ast/lambda.go), [`parser/lambda.go`](sentrie/parser/lambda.go), [`parser/block.go`](sentrie/parser/block.go).
- Calls: [`runtime/eval_call.go`](sentrie/runtime/eval_call.go) — builtins then `alias.fn`; **insert derives first** per resolution order.
- Callables: [`runtime/callable.go`](sentrie/runtime/callable.go) — extend for arity/optional padding on **derive** path only as per `deriveCallable`.
- Index: [`index/index.go`](sentrie/index/index.go), [`index/program.go`](sentrie/index/program.go), [`index/namespace.go`](sentrie/index/namespace.go), [`index/policy.go`](sentrie/index/policy.go).
- FQN parsing for declarations/types: [`parser/fqn.go`](sentrie/parser/fqn.go); value-position `/` is still division until spike rule lands.
- Elvis: absent today — spike adds it.
- Validate: [`index/validate.go`](sentrie/index/validate.go) — add derive DAG; purity uses **PureBuiltins** + call-site resolution + **let** scope stack.

---

## Implementation phases (Sentrie repo) — post-spike unless noted

### Phase A — Lexer and tokens

- `KeywordDerive` + `"derive"` keyword map.
- **`?:`** as dedicated token (recommended: `TokenElvis` in lexer) to avoid ternary ambiguity in [`parser/ternary.go`](sentrie/parser/ternary.go).
- Lambda param `?` remains inside param-list parsing only.

### Phase B — AST

- Extend [`ast/lambda.go`](sentrie/ast/lambda.go): `ParamTypes`, `ParamOpts`, `ReturnType`; `String()` for diagnostics.
- `DeriveStatement`, `ExportDeriveStatement`; FQN/Elvis expression nodes per spike outcome.

### Phase C — Parser

- Top-level derive + `export derive` dispatch (refactor `export` entry: `shape` | `derive`; policy keeps `export decision`).
- Policy derive **only** in body phase (section 6); namespace derive placement per grammar.
- Lambda signature + return type parsing (spike-validated shapes).
- **Same commit/PR:** [`grammar/grammar.peg`](sentrie/grammar/grammar.peg) + [`grammar/grammar.ebnf`](sentrie/grammar/grammar.ebnf).

### Phase D — Index / static analysis

- Namespace / policy derive maps; `checkNameAvailable` includes derives vs policies/shapes/namespaces — **not** vs builtins (section 3).
- `AddProgram` ordering, `ResolveDerive` / export verify, derive **cycle** detection.
- **Runtime bind metadata:** each indexed derive yields enough data for **`deriveCallable` construction** to capture the **define-site** visible derive map (and FQN exports as needed) for inner derive resolution — no reliance on caller rule scope at runtime.
- **Purity walker:** scope stack + **let**; identifier rules; **CallExpression** resolution for TS vs derive vs **PureBuiltins** (section 2); no callable returns.

### Phase E — Runtime

- **`getTarget` order:** derives → builtins → TypeScript ([`runtime/eval_call.go`](sentrie/runtime/eval_call.go)).
- **`DetachedChildContext`** + **`deriveCallable`** — full field table, `slices.Clone` rationale, param seeding, caller-side optional padding, **define-site derive registry closure**, and **`PushRefStack(uniqueID)` using the derive FQN** (never bare name; §4). Anonymous lambdas unchanged.
- Typed param / return validation; optional padding; **Elvis** eval per section 1.
- Collection builtins: unchanged if arity from `deriveCallable` matches HOF expectations; add tests as needed.

### Phase F — Tests (beyond spike)

- Parser: misplaced `derive`, export derive in policy, duplicate params, etc.
- Index: cycles, exports, collisions (no builtin collision tests).
- Purity: facts, TS, import decision, yield lambda, recursion; **let** inside derive (section 5).
- Runtime: shadow builtin by derive name, detached EC (attempted fact read fails or undefined as designed), etc.
- Runtime: **`PushRefStack` FQN** — two namespaces each define `derive check`; one derive body calls the other via **fully qualified** name; execution completes **without** a false infinite-recursion / cycle error (bare-name ref IDs would incorrectly collide).

### Phase G — Optional / defer

- Derive memoization (#88 optional) — not blocking.

---

## Documentation (sentrie-sh/website#34)

Unchanged intent: Derives section (`#derives`), lambdas in [`src/content/docs/reference/index.md`](website/src/content/docs/reference/index.md), collections, getting-started, TypeScript boundary — plus document **Elvis** semantics (null + undefined), **builtin whitelist** rationale at high level if user-facing, and **shadowing** resolution order.

---

## Cross-repo sequencing

- Land sentrie implementation + tests before or with website examples.

---

## Delivery: commits, push, and pull requests

When **executing** this epic (after the spike gates), use the following in addition to each repo’s `AGENTS.md`.

### Sentrie (`sentrie/`)

- **Source of process rules:** [`AGENTS.md`](sentrie/AGENTS.md) — commit policy, commit message style, PR description format (`PR_DESCRIPTION_<branch_name>.md`), license headers on touched **source** files, `sandbox.go` reset after tests, and scope of what belongs in a PR description (`git log` / `git diff` vs base).

- **Commits:** prefer **small, focused commits** — one logical change per commit when practical; combine only when inseparable (same fix, same slice, or mechanical rename). Short imperative subjects; **no** Conventional Commit prefixes; **no** Cursor-style signatures in messages.

- **Explicit authorization:** AGENTS.md says “Do not commit unless explicitly told” and “NEVER commit any `*.md` files.” For this derive epic, **commits are explicitly expected** for **non-markdown** work (`.go`, `.peg`, `.ebnf`, etc.) on the feature branch. **Do not commit** `*.md` in the Sentrie repo (including `PR_DESCRIPTION_*.md`).

- **PR descriptions:** still **generate** `PR_DESCRIPTION_<branch_name>.md` at the Sentrie repo root per [`AGENTS.md`](sentrie/AGENTS.md) (title as first `#` line, `Closes #88` on the Sentrie PR, sections for Summary / Changes by area / Review notes / Testing / Dependencies). Use that file to **paste** into the forge UI when opening the PR, or with `gh pr create --body-file` **without** `git add`/`commit` of the markdown file.

- **Push:** push the feature branch to `origin` after meaningful batches of commits (or at agreed milestones).

- **Pull requests:** open the Sentrie PR with the generated title/body; open a **separate** website PR for docs when that work is ready, following [website `AGENTS.md`](website/AGENTS.md) for that repo’s PR description file and commit rules.

### Website (`website/`)

- Same pattern: small commits, follow [website `AGENTS.md`](website/AGENTS.md); separate PR description file per website rules; PR for documentation aligned with [#34](https://github.com/sentrie-sh/website/issues/34).
