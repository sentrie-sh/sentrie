---
id: lang_test.lang_test
type: Module / File
language: Go
file_path: lang_test/lang_test.go
tags: acceptance-corpus, fixtures, regression-testing
---

# Node: lang_test.TestFixturesParseIndexAndValidate (Language Acceptance Corpus)

## 1. Architectural Role & Intent
Walks every `.sentrie` fixture in `lang_test/` and asserts the full front half of the compiler pipeline succeeds: [[parser.parse]] produces a program, [[index.index]] ingests it, and [[index.validate]] accepts the semantic model. Each fixture runs under its own `t.Run` subtest so failures name the file directly. This replaces the silent `continue` path that previously lived in [[index.builtin_check]]'s repo-pack walk.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `lang_test.lang_test` | `CALLS` | [[parser.parse]] | `ParseProgram` on each fixture source. |
| `lang_test.lang_test` | `CALLS` | [[index.index]] | `CreateIndex` and `AddProgram` per fixture. |
| `lang_test.lang_test` | `CALLS` | [[index.validate]] | Full validation after ingest. |
| `lang_test.lang_test` | `READS_FROM` | [[infra.filesystem]] | Reads sibling `.sentrie` files from the `lang_test/` directory. |

## 3. Interface Contracts & Public Surface

- **Signature:** `TestFixturesParseIndexAndValidate(t *testing.T)`
  - **Behavior:** `ReadDir` on the package directory, skips subdirectories, runs one subtest per `*.sentrie` file. Each subtest calls `require.NoError` on parse, index, and validate with the fixture path in the message.
  - **Side Effects:** None beyond test execution.
  - **Exceptions:** Test failure when any stage returns an error.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless; each subtest builds a fresh index.
- **Performance/Scale Notes:** Forty-three fixtures run in parallel subtests; cost is dominated by parse and validation, not I/O.
- **Dependencies Risk:**
  - **Fixtures must be whole programs.** Policies need at least one exported rule for [[index.policy]] ingest to succeed; parse-only snippets fail index with `does not export any rules`.
  - **Corpus rot is visible again.** Before this harness, [[index.builtin_check]]'s repo walk used `continue` on parse and index errors, so broken fixtures reported green. Any fixture edit that breaks parse or validation fails CI immediately.
