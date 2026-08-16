---
id: parser.namespace
type: Function / Endpoint
language: Go
file_path: parser/namespace.go
tags: declaration, namespace, front-end, program-structure
---

# Node: parser.parseNamespaceStatement (Namespace Declaration)

## 1. Architectural Role & Intent
Parses `namespace <fqn>`, the mandatory first declaration of every Sentrie file. It is a thin production — consume the keyword, delegate the path to [[parser.fqn]], stitch the span — but it produces the node that [[index.package]] uses as the root key for every policy, shape, and derive in the file, making it the anchor of the entire symbol hierarchy.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `parser.namespace` | `CALLS` | [[parser.fqn]] | Delegates the slash-separated path to `parseFQN`. |
| `parser.namespace` | `CALLS` | [[parser.parser]] | Uses `head()`, `expect(KeywordNamespace)`. |
| `parser.namespace` | `CALLS` | [[ast]] | Emits `ast.NewNamespaceStatement`. |
| [[parser.statement]] | `CALLS` | [[parser.namespace]] | Registered for `tokens.KeywordNamespace` in the **top-level** table only. |
| [[parser.parse]] | `DEPENDS_ON` | [[parser.namespace]] | Type-asserts the first non-comment statement to `*ast.NamespaceStatement`. |
| [[index.namespace]] | `DEPENDS_ON` | [[parser.namespace]] | Consumes the produced node to build the namespace tree. |

## 3. Interface Contracts & Public Surface

- **Signature:** `parseNamespaceStatement(ctx: context.Context, p: *Parser) -> ast.Statement`
  - **Behavior:** Captures the keyword's range, consumes `namespace`, parses the FQN, and extends the span to the FQN's end so the statement covers `namespace a/b/c` in full.
  - **Side Effects:** Consumes tokens; emits `PARSE_NS` / `PARSE_NS_DONE` debug logs.
  - **Exceptions:** Returns `nil` (with the error already recorded by `expect`/`parseFQN`) if the keyword is missing or the FQN fails to parse.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless function over the parser's token window.
- **Performance/Scale Notes:** Two debug log calls per invocation; once per file, so irrelevant.
- **Dependencies Risk:**
  - **Position in the file is enforced elsewhere.** This production happily parses a `namespace` declaration anywhere; the "must be first" and "only one" rules live in [[parser.parse]]. Reading this file alone understates the constraint.
  - **The FQN is syntax only.** No check that the namespace path is unique, non-empty beyond one segment, or consistent with the file's directory. Those are [[index.package]]'s job.
