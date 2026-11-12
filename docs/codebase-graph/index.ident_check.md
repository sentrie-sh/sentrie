---
id: index.ident_check
type: Function / Endpoint
language: Go
file_path: index/ident_check.go
tags: static-analysis, name-resolution, diagnostics
---

# Node: index.checkIdentifiers (Static Identifier Resolution)

## 1. Architectural Role & Intent
Reports every identifier, evaluated as a value, that names nothing at its site, such as `yield undefined_thing` ([#155](https://github.com/sentrie-sh/sentrie/issues/155)). Before it, such a pack passed `sentrie validate` and failed only when the expression was evaluated, with `identifier not found`. It walks the same expression sites as the other static checks through `forEachScopedExpr` ([[index.builtin_check]]) and asks the scope whether the name is a value there (`namesValue`, [[index.is_resolve]]), which mirrors the runtime's `evalIdent` ([[runtime.eval_ident]]).

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `index.ident_check` | `CALLS` | [[index.builtin_check]] | `forEachScopedExpr` supplies each expression with its lexical scope. |
| `index.ident_check` | `CALLS` | [[index.is_resolve]] | `namesValue` decides whether a name is defined; `joinSiteErrors` orders the report. |
| `index.ident_check` | `CALLS` | [[index.expr_walk]] | `forEachExprChild` passes the not-evaluated mark of a callee to its children. |
| `index.ident_check` | `DEPENDS_ON` | [[builtins]] | `builtins.Table` for the builtin hint. |
| `index.ident_check` | `READS_FROM` | [[index.policy]] | `Policy.Uses` for the module alias hints. |
| `index.ident_check` | `DEPENDS_ON` | [[xerr]] | Errors wrap `ErrIndex`. |
| [[index.validate]] | `CALLS` | [[index.ident_check]] | Runs right after [[index.self_reference]]. |

## 3. Interface Contracts & Public Surface

- **Signature:** `(*Index).checkIdentifiers(ctx) -> error`
  - **Behavior:** Defined means a lambda parameter or block let in scope, pack data, a fact (by alias), a policy let, a rule, or a visible derive (policy, then namespace). The callee of a call is skipped whole: `getTarget` never evaluates it, an identifier callee is [[index.call_check]]'s, and `alias.fn` names a `use` module alias. Derive bodies are skipped: [[index.derive_purity]] already rejects every name they cannot read. A name bound by any lambda parameter or block let anywhere in the same policy is never reported (see below).
  - **Side Effects:** None.
  - **Exceptions:** Joined, in source order, one per identifier at its span: `identifier '<name>' not found at <span>: it is not a parameter, let, fact, data, rule or derive in scope: index error`. The middle names the fix when the name is a `use` alias (`is a module alias, not a value`), a name imported from a module (`must be called as alias.name(...)`) or a builtin (`is a builtin, not a value`). `validation cancelled` on context cancellation.

## 4. Operational Context & Gotchas
- **The runtime scopes rules and lets dynamically.** `evalIdent` evaluates a rule or let in the context that references it, so its body also sees the parameters and block lets of that site: `let y = x + 1` read inside `(x) => { yield y }` reads the parameter, and a rule read from inside `{ let z = 1 ... }` sees `z`. Such a name is not lexically in scope but can resolve, so the check suppresses every name any lambda parameter or block let in the policy binds. That weakens it for those names; it never reports one that can resolve. Imported decisions start a fresh root context, and derive bodies a detached one, so neither leaks.
- **Builtins and module aliases are not values.** `count` or `time` alone is an error at runtime too; only a call reaches them.
- **Fact defaults run before lets are bound.** A fact default that reads a let fails at runtime, but the let is defined, so it is not reported here.
- **Shape and type names** are [[index.shape_refs]]' and [[index.is_resolve]]'s; this check sees only expressions.
