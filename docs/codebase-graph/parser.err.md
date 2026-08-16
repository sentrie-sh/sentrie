---
id: parser.err
type: Module / File
language: Go
file_path: parser/err.go
tags: error-handling, sentinel, dead-code
---

# Node: parser.err (Parse Error Sentinel)

## 1. Architectural Role & Intent
`parser/err.go` declares a single sentinel, `ErrParse`, intended as the wrappable identity for parse failures. In practice the parser does not use it: [[parser.parser]] builds every diagnostic with `fmt.Errorf` and combines them via `errors.Join`, so no returned error wraps this sentinel. It is documented here specifically to stop future code from writing `errors.Is(err, parser.ErrParse)` and getting a silently-always-false result.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `parser.err` | `IMPORTS` | `std.errors` | Sole dependency. |
| [[parser.err]] | `DEPENDS_ON` | [[parser.parser]] | The sentinel is declared for use by the parser's `errorf` path, which never actually wraps it. |

## 3. Interface Contracts & Public Surface

- **Signature:** `ErrParse` — `errors.New("parse error")`
  - **Behavior:** An exported sentinel with no producers. Matching against it never succeeds.
  - **Side Effects:** None.
  - **Exceptions:** N/A.

## 4. Operational Context & Gotchas
- **Statefulness:** Package-level immutable value.
- **Performance/Scale Notes:** None.
- **Two non-edges worth stating.** [[parser.parser]] produces the errors that actually reach callers — joined `fmt.Errorf` values, **not** wrappers of this sentinel. And parse errors do **not** flow through [[xerr]]; unlike index and runtime diagnostics they are plain joined errors with no structured span payload.
- **Dependencies Risk:** **Do not classify parse failures with `errors.Is`.** Errors returned by `ParseProgram` are `errors.Join` trees of formatted messages, each prefixed `parsing error at <file>:<line>:<column>:`. Callers wanting to distinguish parse failures from I/O or validation failures must do so by call site (they came out of [[parser.parse]]) or by wrapping at the boundary — [[loader]] takes the call-site approach. Either wire this sentinel through `errorf` or delete it; leaving it exported invites a false-negative check.
