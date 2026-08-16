---
id: parser.import
type: Module / File
language: Go
file_path: parser/import.go
tags: cross-namespace, decision-import, rule-body, parameter-binding
---

# Node: parser.import (Cross-Namespace Decision Import)

## 1. Architectural Role & Intent
Parses `import decision <rule> from <fqn> (with <ident> as <expr>)*`, the alternative rule body that delegates a decision to a rule in another namespace. The `with` clauses bind values into the imported policy's facts, making this the composition mechanism by which one policy reuses another's logic without duplicating it. It is the consumer side of what [[parser.export_rule]] declares.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `parser.import` | `CALLS` | [[parser.fqn]] | Parses the source namespace path after `from`. |
| `parser.import` | `CALLS` | [[parser.expression]] | Parses each `with … as <expr>` binding at `LOWEST`. |
| `parser.import` | `CALLS` | [[ast]] | Emits `ast.NewImportClause` and `ast.NewWithClause`. |
| [[parser.rule]] | `CALLS` | [[parser.import]] | The only caller: a rule body beginning with `import` routes here instead of to an expression. |
| [[parser.export_rule]] | `DEPENDS_ON` | [[parser.import]] | The imported rule must be exported by the source policy. |
| [[runtime.imports]] | `DEPENDS_ON` | [[parser.import]] | Resolves the target policy, binds the `with` values, and evaluates the remote decision. |
| [[index.resolve]] | `DEPENDS_ON` | [[parser.import]] | Validates that the referenced namespace and rule exist and are visible. |

## 3. Interface Contracts & Public Surface

- **Signature:** `parseImportExpression(ctx, p) -> ast.Expression`
  - **Behavior:** Requires `import`, `decision`, a rule identifier, `from`, and an FQN, then consumes zero or more `with` clauses, extending the span across each. Returns an `ImportClause` that occupies a rule's body slot - note it is an **expression**, not a statement, despite being usable only in that one position.
  - **Side Effects:** Consumes tokens.
  - **Exceptions:** Returns `nil` on a missing keyword, rule identifier, or FQN.

- **Signature:** `parseWithClause(ctx, p) -> *ast.WithClause`
  - **Behavior:** Parses `with <ident> as <expr>`, binding an arbitrary expression to a named fact of the imported policy. The doc comment claims the value is a string; it is in fact any expression.
  - **Side Effects:** Consumes tokens.
  - **Exceptions:** Returns `nil` on a missing `with`, identifier, `as`, or a failed expression.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless.
- **Performance/Scale Notes:** Nothing notable at parse time; the runtime cost is a nested policy evaluation per import.
- **Dependencies Risk:**
  - **A failed `with` clause is swallowed.** The loop appends only when `parseWithClause` returns non-nil, but does **not** return on nil - so a malformed clause is skipped and the import is built without it. Combined with the loop condition (`head is 'with'`), a clause that fails after consuming its `with` token can leave the parser mid-clause and produce a cascade of unrelated errors. This is the weakest error handling in the package.
  - **Nothing is resolved here.** The target namespace, the rule's existence, whether it is exported, and whether the `with` names match the target's facts are all [[index.resolve]]'s concern. A typo in the FQN parses cleanly.
  - **It is an expression in a statement-shaped role.** `ImportClause` satisfies `ast.Expression` but is only ever valid as a rule body, so a generic expression walker may encounter it in a position it does not expect.
  - **The doc comment is stale** - it specifies `withClause ::= 'with' IDENT 'as' IDENT` and a `blockExpr` production that this file does not implement.
