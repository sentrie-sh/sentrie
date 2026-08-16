---
id: index.program
type: Class
language: Go
file_path: index/program.go
tags: file-record, provenance, data-model
---

# Node: index.Program (Per-File Record)

## 1. Architectural Role & Intent
A flat per-source-file record that partitions one `ast.Program`'s top-level statements by kind: the namespace declaration, policies, shapes, and shape exports. It exists so the index can answer file-oriented questions - what did this file contribute? - without re-walking the AST, and it retains the original `ast.Program` for provenance.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `index.program` | `DEPENDS_ON` | [[ast]] | Holds references to the original program and its statements; performs no copying. |
| [[index.index]] | `CALLS` | [[index.program]] | `AddProgram` calls `createProgram` and stores the result under the file path. |

## 3. Interface Contracts & Public Surface

- **Signature:** `Program` struct - `{ Reference *ast.Program, Namespace *ast.NamespaceStatement, Policies []*ast.PolicyStatement, Shapes []*ast.ShapeStatement, ShapeExports []*ast.ShapeExportStatement }`
  - **Behavior:** `Reference` is the whole parsed program; the other fields are filtered views into its statement list. Note the field is named `Reference` while the *key* under which the record is stored is `ast.Program.Reference`, the file path - two different meanings of the same word.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `createProgram(astProgram *ast.Program) -> *Program`
  - **Behavior:** Single pass over all statements, appending each recognised kind to its slice and assigning the namespace. Unrecognised kinds - including `DeriveStatement` and `ExportDeriveStatement` - are silently ignored.
  - **Side Effects:** Allocates the slices.
  - **Exceptions:** None; it cannot fail.

## 4. Operational Context & Gotchas
- **Statefulness:** Immutable after construction.
- **Performance/Scale Notes:** One linear pass per file. Negligible.
- **Dependencies Risk:**
  - **Incomplete by design, but silently so.** Derives and derive exports are real top-level statements that this record does not capture, so `Program` is *not* a faithful inventory of a file's contributions. Anything asking "what does this file declare?" will under-report derives.
  - **Unlike `AddProgram`, this walks from index 0** and picks up the namespace correctly regardless of leading comments - so the two file walks in the package disagree about where statements start.
  - **No validation.** A file with two namespace statements would leave only the last one here; the rejection happens in [[parser.parse]], not in this record.
