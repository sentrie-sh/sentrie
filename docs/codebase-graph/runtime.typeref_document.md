---
id: runtime.typeref_document
type: Function / Endpoint
language: Go
file_path: runtime/typeref_document.go
tags: type-validation, constraints, schemaless
---

# Node: runtime.validateAgainstDocumentTypeRef

## 1. Architectural Role & Intent
Validates a value against the `document` type — the schemaless container intended for arbitrary external payloads such as a Kubernetes manifest or a cloud resource description. It checks only that the value is a map, delegating any real structural expectation to constraints from the document checker table.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `runtime.typeref_document` | `DEPENDS_ON` | [[constraints]] | `constraints.DocumentContraintCheckers` — the table that gives `document` whatever teeth it has. |
| `runtime.typeref_document` | `CALLS` | [[runtime.eval]] | Evaluates constraint argument expressions. |
| `runtime.typeref_document` | `CALLS` | [[runtime.err_typedef]] | Constraint error constructors. |
| [[runtime.typeref]] | `CALLS` | [[runtime.typeref_document]] | Dispatched for `*ast.DocumentTypeRef`. |

## 3. Interface Contracts & Public Surface

- **Signature:** `validateAgainstDocumentTypeRef(ctx, ec, exec Executor, p *index.Policy, v box.Value, typeRef *ast.DocumentTypeRef, pos tokens.Range) -> error`
  - **Behavior:** Requires `v.DictValue()` — the source comment states plainly that it "just validate[s] that it's a map" — then runs document constraints.
  - **Side Effects:** Constraint argument evaluation.
  - **Exceptions:** `value %v is not a document at %s - expected document`; `ErrUnknownConstraint`; `ErrConstraintFailed`.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless.
- **Performance/Scale Notes:** O(constraints); document size is irrelevant to the base check.
- **Dependencies Risk:**
  - **A document is a dict with a different constraint table.** The two validators are structurally the same; the type distinction exists to select `DocumentContraintCheckers` over `DictContraintCheckers`. Anything meaningful about `document` lives in [[constraints]], not here.
  - **Only maps qualify.** A YAML or JSON document whose root is a list fails, so multi-document or array-rooted payloads must be wrapped before they can be declared as `document`.
  - **A document-typed fact carries no guarantee about its contents**, so every field access against it is an unchecked lookup that yields `Undefined` on a miss. Policies over documents should pair the type with `is defined` guards or explicit constraints.
  - Shares the `exec.(*executorImpl)` assertion described in [[runtime.typeref]].
