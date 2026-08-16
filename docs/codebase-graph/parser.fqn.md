---
id: parser.fqn
type: Function / Endpoint
language: Go
file_path: parser/fqn.go
tags: identifier-path, namespace, shared-production
---

# Node: parser.parseFQN (Fully-Qualified Name Production)

## 1. Architectural Role & Intent
Parses a slash-separated identifier path (`a/b/c`) into an `ast.FQN`. It is the shared building block behind namespace declarations, shape references, `with` clauses on complex shapes, and cross-namespace decision imports - every place Sentrie names something outside the current scope resolves through this one production.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `parser.fqn` | `CALLS` | [[parser.parser]] | Uses `advanceExpected(Ident)` and `canExpect(TokenDiv)`. |
| `parser.fqn` | `CALLS` | [[ast]] | Emits `ast.NewFQN` with the accumulated parts and a span covering the whole path. |
| [[parser.namespace]] | `CALLS` | [[parser.fqn]] | Namespace path. |
| [[parser.shape]] | `CALLS` | [[parser.fqn]] | The `with <fqn>` base-shape reference. |
| [[parser.import]] | `CALLS` | [[parser.fqn]] | The `from <fqn>` source of an imported decision. |
| [[parser.typeref]] | `CALLS` | [[parser.fqn]] | Named shape type references. |

## 3. Interface Contracts & Public Surface

- **Signature:** `parseFQN(ctx: context.Context, p: *Parser) -> *ast.FQN`
  - **Behavior:** Requires at least one identifier, then greedily consumes `/ <ident>` pairs. The span runs from the first identifier to the last.
  - **Side Effects:** Consumes tokens; emits `PARSE_FQN` / `PARSE_FQN_DONE` debug logs.
  - **Exceptions:** Returns `nil` when the first token is not an identifier, or when a `/` is not followed by one - the error is recorded by `advanceExpected`.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless.
- **Performance/Scale Notes:** Two debug log calls per FQN; FQNs are frequent in type-heavy policies, so this is a visible contributor to debug-log volume.
- **Dependencies Risk:**
  - **`/` is also the division operator.** The path separator and `TokenDiv` are the same token kind, so FQN parsing is only unambiguous because it runs in declaration positions where an expression cannot appear. Any future attempt to accept an FQN in expression position will collide with arithmetic.
  - **Returns a pointer to a local.** `ast.FQN` is a value type; this production returns `&fqn`, so callers that dereference (as [[parser.namespace]] does) get a copy while callers that store the pointer (as shape `with` does) share it. Be deliberate about which form you hold.
  - **No validation.** Segment naming, depth, and existence are entirely unchecked here - resolution happens in [[index.resolve]].
