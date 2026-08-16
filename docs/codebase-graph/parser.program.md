---
id: parser.program
type: Class
language: Go
file_path: parser/program.go
tags: dead-code, vestigial, data-model
---

# Node: parser.Program (Vestigial Program Struct)

## 1. Architectural Role & Intent
`parser/program.go` declares a `Program` struct holding a statement slice. It is **not** the type produced by parsing — [[parser.parse]] returns `ast.Program` — and nothing in the repository constructs, returns, or consumes this type. It is recorded here so that agents searching for "the parser's program type" resolve the ambiguity immediately rather than mistaking it for the real output contract.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `parser.program` | `DEPENDS_ON` | [[ast]] | Declares a `[]ast.Statement` field; this is its only linkage. |
| [[parser.parse]] | `CALLS` | [[ast]] | **Does not** use this type — the real output is `ast.Program`, which additionally carries `Reference`. |

## 3. Interface Contracts & Public Surface

- **Signature:** `Program` struct — `{ Statements []ast.Statement }`
  - **Behavior:** Inert. No constructor, no methods, no references.
  - **Side Effects:** None.
  - **Exceptions:** None.

## 4. Operational Context & Gotchas
- **Statefulness:** N/A — never instantiated.
- **Performance/Scale Notes:** None.
- **Dependencies Risk:** The only hazard is **name collision confusion**: `parser.Program` and `ast.Program` are distinct types with near-identical shapes, and the exported-but-unused one is the wrong one. Anything that appears to accept a "parser program" is either mistaken or dead. The authoritative parse result is `ast.Program`, which carries the source `Reference` this struct lacks. Safe to delete; kept in the graph so the ambiguity is documented rather than rediscovered.
