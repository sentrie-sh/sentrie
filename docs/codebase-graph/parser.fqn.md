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
| `parser.fqn` | `CALLS` | [[parser.parser]] | Uses `advanceExpected(Ident)` for the first segment, `canExpect(TokenDiv)`, and `advanceNameOrKeyword` for every later segment. |
| `parser.fqn` | `CALLS` | [[ast]] | Emits `ast.NewFQN` with the accumulated parts and a span covering the whole path. |
| [[parser.namespace]] | `CALLS` | [[parser.fqn]] | Namespace path. |
| [[parser.shape]] | `CALLS` | [[parser.fqn]] | The `with <fqn>` base-shape reference. |
| [[parser.import]] | `CALLS` | [[parser.fqn]] | The `from <fqn>` source of an imported decision. |
| [[parser.typeref]] | `CALLS` | [[parser.fqn]] | Named shape type references. |

## 3. Interface Contracts & Public Surface

- **Signature:** `parseFQN(ctx: context.Context, p: *Parser) -> *ast.FQN`
  - **Behavior:** Requires an identifier first segment, then greedily consumes `/ <segment>` pairs, where a later segment is an identifier **or a keyword** (`namespace constraints/in`, `text/string/Item`, #297). The span runs from the first segment to the last.
  - **Side Effects:** Consumes tokens; emits `PARSE_FQN` / `PARSE_FQN_DONE` debug logs.
  - **Exceptions:** Returns `nil` when the first token is a keyword (`expected Ident, got keyword '<kw>': the first segment of a name cannot be a keyword`) or otherwise not an identifier (`expected Ident, got <kind>`), or when a `/` is not followed by an identifier or keyword.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless.
- **Performance/Scale Notes:** Two debug log calls per FQN; FQNs are frequent in type-heavy policies, so this is a visible contributor to debug-log volume.
- **Dependencies Risk:**
  - **`/` is also the division operator.** The path separator and `TokenDiv` are the same token kind, so FQN parsing is only unambiguous because it runs in declaration and type positions (including the type after `as` in a cast), where after a name a `/` can only continue the path. Slash-qualified names in expression position (a derive callee `lib/in/double(2)`, the right of `is`) are not read here: they are division chains flattened by `ast.FlattenSlashIdentChain`. [[parser.infix]] reads a keyword after `/` as an identifier segment when it has no prefix handler (`isNameOnlyKeyword` in [[parser.parser]]), so those agree with this production. The exception is the six keywords that start an expression (`not`, `true`, `false`, `unknown`, `null`, `transform`): after `/` they keep that meaning (`a/not b`, `a/true` are division), so a namespace with one of them as a later segment can be declared, imported from and used as a type, but its derives and shapes cannot be named by a slash chain in an expression.
  - **The first segment stays an identifier.** [[parser.typeref]] dispatches on the first token, so a type keyword there (`string`) must keep its type meaning; requiring an identifier first segment everywhere keeps every namespace nameable from a type position.
  - **Returns a pointer to a local.** `ast.FQN` is a value type; this production returns `&fqn`, so callers that dereference (as [[parser.namespace]] does) get a copy while callers that store the pointer (as shape `with` does) share it. Be deliberate about which form you hold.
  - **No validation.** Segment naming, depth, and existence are entirely unchecked here - resolution happens in [[index.resolve]].
