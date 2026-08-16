---
id: parser.derive
type: Function / Endpoint
language: Go
file_path: parser/derive.go
tags: declaration, derive, lambda, dependency-graph
---

# Node: parser.parseDeriveStatement (Derive Declaration)

## 1. Architectural Role & Intent
Parses `derive <ident> = <lambda>`, the named reusable computation that both namespaces and policies can declare. Its single enforced rule is that the right-hand side **must** be a lambda expression, not an arbitrary expression — that restriction is what lets [[index]] treat derives as callable units with analysable dependencies, order them topologically via [[dag]], and reject cyclic definitions before evaluation.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `parser.derive` | `CALLS` | [[parser.expression]] | Parses the right-hand side at `LOWEST`, then type-asserts it. |
| `parser.derive` | `CALLS` | [[parser.parser]] | Uses `head`, `advance`, `advanceExpected`, `expect(TokenAssign)`, `errorf`. |
| `parser.derive` | `CALLS` | [[ast]] | Emits `ast.NewDeriveStatement(name, lambda, span)`. |
| [[parser.statement]] | `CALLS` | [[parser.derive]] | Registered for `tokens.KeywordDerive` at **top level**. |
| [[parser.policy]] | `CALLS` | [[parser.derive]] | Registered for the same kind in the **policy scope** — the same handler serves both. |
| [[index.derive]] | `DEPENDS_ON` | [[parser.derive]] | Registers the derive as a graph node and extracts its dependencies. |
| [[index.derive_cycle]] | `DEPENDS_ON` | [[parser.derive]] | Detects cyclic definitions among the derives this production declares. |
| [[runtime.derive_invoke]] | `DEPENDS_ON` | [[parser.derive]] | Invokes the derive's lambda during evaluation. |

## 3. Interface Contracts & Public Surface

- **Signature:** `parseDeriveStatement(ctx: context.Context, p: *Parser) -> ast.Statement`
  - **Behavior:** Consumes `derive` **unconditionally** (bare `advance()`; the handler table guarantees the head), then the name, then a required `=`, then any expression — which is immediately type-asserted to `*ast.LambdaExpression`. The span runs from the keyword to the lambda's end.
  - **Side Effects:** Consumes tokens; may record errors.
  - **Exceptions:** Returns `nil` for a missing name, a missing `=`, a failed expression parse, or a non-lambda right-hand side (`derive value must be a lambda expression`).

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless.
- **Performance/Scale Notes:** Nothing notable.
- **Dependencies Risk:**
  - **One handler, two scopes, different export rules.** The same production runs at namespace and policy scope, but only **namespace-level** derives may be exported — [[parser.export_rule]] explicitly rejects `export derive` inside a policy. The asymmetry is invisible from this file.
  - **The lambda requirement is a post-hoc assertion.** The expression is fully parsed before the type check, so `derive x = 1 + 2` produces the "must be a lambda" error positioned at the derive rather than a targeted syntax error, and any comment wrapper around the lambda (see [[parser.expression]]) would fail the assertion outright.
  - **Zero-arity derives are legal.** `() => { yield … }` parses fine; whether a derive may reference facts or must be pure is enforced by [[index.derive_purity]], not here.
