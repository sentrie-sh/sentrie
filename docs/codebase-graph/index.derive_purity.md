---
id: index.derive_purity
type: Function / Endpoint
language: Go
file_path: index/derive_purity.go
tags: purity, sandboxing, scoping, static-analysis, security
---

# Node: index.validateDerivePurity (Purity Enforcement)

## 1. Architectural Role & Intent
The static gate that makes `derive` genuinely pure. It walks every derive body with an explicit lexical scope and rejects anything that could read policy state or reach the outside world: facts, rules, TypeScript module calls, non-derive-safe builtins, unknown identifiers, and yielding a lambda. Because derives are shared across policies and evaluated without a policy context, this check - not the runtime - is what guarantees a derive cannot smuggle in ambient state.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `index.derive_purity` | `DEPENDS_ON` | [[ast]] | Exhaustive type-switches over every expression form; unknown forms are a hard error. |
| `index.derive_purity` | `DEPENDS_ON` | [[builtins]] | `builtins.IsDeriveSafe(name)` is the allowlist for calls to native functions. |
| `index.derive_purity` | `CALLS` | [[index.derive_expr_walk]] | Delegates generic child traversal to `forEachDeriveExprChild`. |
| `index.derive_purity` | `READS_FROM` | [[index.derive]] | Consults `DefineShort` / `DefineFQN` snapshots and `VisibleFromDeriveCaller`. |
| `index.derive_purity` | `READS_FROM` | [[index.policy]] | Checks `Policy.Facts` and `Policy.Rules` purely to produce a *better error message*. |
| [[index.validate]] | `CALLS` | [[index.derive_purity]] | Runs after derive cycle detection. |

## 3. Interface Contracts & Public Surface

- **Signature:** `(*Index).validateDerivePurity() -> error`
  - **Behavior:** Iterates `DerivesByFQN` and validates each, prefixing failures with `derive <fqn>:`.
  - **Side Effects:** None.
  - **Exceptions:** The first failure encountered; iteration order over a map means **which** failure surfaces first is nondeterministic across runs.

- **Signature:** `validateDerivePure(idx, d) -> error`
  - **Behavior:** Seeds the scope with the lambda's parameter names and walks the body block.
  - **Side Effects:** None.
  - **Exceptions:** Propagated from the walk.

- **Signature:** `walkDeriveBlock(idx, d, b, scope) -> error`
  - **Behavior:** A derive body may contain **only `let` declarations before `yield`**. Each let's initializer is checked against the scope *before* the binding is added, so a let cannot reference itself. The scope is cloned per binding, giving proper sequential shadowing.
  - **Side Effects:** None; scopes are copied.
  - **Exceptions:** `derive body may only contain let declarations before yield (got %T)`.

- **Signature:** `yieldHasNoLambda(e) -> error` / `scanLambdasOutsideCalls(e, underCall bool) -> error`
  - **Behavior:** Rejects a lambda appearing in a yielded position unless it is an **argument to a call** - so `filter(xs, x => …)` is fine but returning a function is not. This keeps derives first-order at their boundary.
  - **Side Effects:** None.
  - **Exceptions:** `derive cannot yield a lambda value`; `derive purity: unsupported expression in yield scan %T`.

- **Signature:** `walkDeriveExpr` / `walkDeriveExprSeen(idx, d, e, scope, seen) -> error`
  - **Behavior:** The main recursive check, carrying a `seen` set of expression pointers to detect a cyclic AST.
  - **Side Effects:** None.
  - **Exceptions:** `derive purity: cyclic expression graph`; `derive purity: unsupported expression %T`.

- **Signature:** `checkDeriveIdentifier(d, name, scope) -> error`
  - **Behavior:** Accepts only names in lexical scope. Produces **targeted diagnostics**: a visible derive used bare says "must be called as name(...)"; a fact says "facts are not available inside a derive"; a rule says "rules cannot be referenced inside a derive"; anything else gets the general message.
  - **Side Effects:** None.
  - **Exceptions:** One of the four messages above.

- **Signature:** `checkDeriveCall(idx, d, c, scope, seen) -> error`
  - **Behavior:** Rejects field-access callees outright (that is the TypeScript module call form). Resolves slash-qualified callees of **three or more segments** through the snapshots then the global registry, enforcing visibility. Otherwise accepts in-scope callables, visible short-name derives, and derive-safe builtins. Everything else fails.
  - **Side Effects:** None.
  - **Exceptions:** `TypeScript module calls are not permitted inside a derive`; `unknown derive %q`; visibility errors; `call is not permitted inside a derive (only visible derives and pure builtins)`.

- **Signature:** `isAllowedDeriveCallbackArg(builtin string, argIdx int) -> bool`
  - **Behavior:** A hardcoded allowlist letting a single-parameter derive be passed as a callback in argument position 1 of `filter`, `collect`, `any`, `all`, `first`, and `distinct`.
  - **Side Effects:** None.
  - **Exceptions:** None.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless. Scope maps are cloned rather than mutated, so no state leaks between branches.
- **Performance/Scale Notes:** `cloneScope` copies on every let and every lambda, making deeply nested bodies quadratic in binding count. The `seen` map allocates once per top-level expression walk.
- **Dependencies Risk:**
  - **This is a security boundary, not a style check.** It is the only thing preventing a derive from calling into a TypeScript module or reading facts. Weakening a case here widens the sandbox.
  - **Exhaustive switches are a maintenance hazard.** Three separate switches - here, in `scanLambdasOutsideCalls`, and in [[index.derive_expr_walk]] - must all learn about any new `ast.Expression` type. Miss one and you get a spurious "unsupported expression" rejection of valid code.
  - **The callback allowlist is hardcoded by name and index.** Adding a higher-order pure builtin to [[builtins]] without updating `isAllowedDeriveCallbackArg` means derives cannot be passed to it, even though the builtin is derive-safe.
  - **Slash-qualified resolution requires three segments.** A two-segment path falls through to the generic identifier path and reports a less helpful error.
  - **Error selection is nondeterministic.** Because `DerivesByFQN` is a map, a pack with purity errors in several derives may report a different one on each run - confusing when iterating on fixes.
