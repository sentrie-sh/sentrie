---
id: runtime.eval_ident
type: Function / Endpoint
language: Go
file_path: runtime/eval_ident.go
tags: name-resolution, lazy-evaluation, memoization, purity, recursion-detection
---

# Node: runtime.evalIdent (Identifier Resolution)

## 1. Architectural Role & Intent
Resolves a bare identifier through a fixed precedence chain — local, fact, let, rule, derive — and is where Sentrie's **lazy evaluation** lives: a `let` is not evaluated when declared but when first read, then cached as a local so subsequent reads are free. Referencing a rule by name triggers a full nested rule execution, which makes identifier resolution one of the most expensive operations in the evaluator.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `runtime.eval_ident` | `READS_FROM` | [[runtime.exec_ctx]] | `GetLocal`, `GetFact`, `GetLet`, `SetLocal`, `PushRefStack`/`PopRefStack`, `evalDerive`. |
| `runtime.eval_ident` | `CALLS` | [[runtime.eval]] | Evaluates a let's initializer expression on first read. |
| `runtime.eval_ident` | `CALLS` | [[runtime.executor]] | A rule reference calls `exec.execRule` — a complete nested rule evaluation. |
| `runtime.eval_ident` | `CALLS` | [[runtime.typeref]] | Validates a let's value against its declared type. |
| `runtime.eval_ident` | `CALLS` | [[runtime.eval_call]] | Reuses `lookupDeriveByIdentifier` for the derive-as-value case. |
| `runtime.eval_ident` | `CALLS` | [[runtime.callable]] | Boxes a derive as a `deriveCallable` for higher-order use. |
| `runtime.eval_ident` | `READS_FROM` | [[index]] | Reads `Policy.Facts` (for the derive error message) and `Policy.Rules`. |
| [[runtime.eval]] | `CALLS` | [[runtime.eval_ident]] | All `ast.Identifier` nodes dispatch here. |

## 3. Interface Contracts & Public Surface

- **Signature:** `evalIdent(ctx, ec, exec, p, i *ast.Identifier) -> (box.Value, *trace.Node, error)`
  - **Behavior:** Resolves in strict order:
    1. **Local** — an already-evaluated value, including lambda and derive parameters.
    2. **Derive fact guard** — inside a derive, a name matching a policy fact is refused with a targeted message rather than falling through.
    3. **Fact** — resolved at the root context.
    4. **Let** — pushed onto the recursion stack, evaluated, type-validated if annotated, then cached via `SetLocal`.
    5. **Rule** — outside a derive only; executes the rule and caches its **decision value**.
    6. **Derive** — boxed as a callable so it can be passed to higher-order builtins.
  - **Side Effects:** Evaluates let initializers; executes rules; writes locals; mutates the recursion stack.
  - **Exceptions:** `facts are not available inside a derive (%q)`; `xerr.ErrInfiniteRecursion` from the let guard; `invalid value for let declaration %s`; `identifier not found: %s`.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless itself, but it **mutates the execution context** as a cache — the first read of a let or rule changes the context for every later read.
- **Performance/Scale Notes:**
  - Lets and rules are memoized per context via `SetLocal`, so repeated references are cheap after the first. But because `SetLocal` without `force` walks up to the context that *declares* the name, the caching level depends on where the declaration lives.
  - A rule reference is a **full `execRule`**, including that rule's own fact-type validation and trace subtree. A rule referenced from several other rules is re-executed once per referencing context, not once per execution.
- **Dependencies Risk:**
  - **Rule references cache only `decision.Value`, discarding `decision.State`.** A rule that yields `Unknown` with a non-nil value is indistinguishable, when referenced by name, from one that yielded that value definitively. Trinary information is lost at the identifier boundary.
  - **Recursion detection covers lets but not rules here.** The let branch pushes the identifier onto the reference stack; the rule branch relies on `execRule` pushing the rule's FQN. The two use **different key spaces** — a bare name versus a fully-qualified name — so a let and a rule with the same name occupy separate stack entries.
  - **Purity is enforced by two separate mechanisms.** The fact guard produces a good message, but the real protection is structural: a detached derive context has no facts map and no parent, so `GetFact` finds nothing. The rule branch is gated explicitly on `evalDerive == nil`.
  - **The fact guard reads `p.Facts` with the identifier as key**, and that map is **alias-keyed** — so the friendly error only appears for the alias, not the declared name.
  - **`identifier not found` is the catch-all.** Because resolution has six stages, this message gives no indication of which stage was expected to match, making typos in derive names and typos in fact aliases indistinguishable.
