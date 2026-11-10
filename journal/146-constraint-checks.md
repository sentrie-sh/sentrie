## Binaek Sarkar (binaek89@gmail.com)

### 2026-11-10 15:42 - Pick up #146
- `@min(-5)` fails to parse, `@in` cannot be written, `@url` rejects ports and recompiles its regex per call, `@length` counts bytes (constraints/string.go:29), and the parse error is garbled (`%!s(int=3)`).
- Picked up in batch 1 of the v0.0.5 milestone (#307/#308, #149, #143, #146, #141, #159): one pickup push and a draft PR per branch, plus a paired `sentrie-sh/website` branch of the same name cut from `next`, before implementation.
- Decisions recorded with the user on 2026-11-10, which the reviewer judges the change against:
  - `@in` and `@not_in` take variadic arguments only (`@in(1, 2)`), not a list (the user chose this over accepting both).
  - On a string, `@in("a", "b")` means the value equals one of the listed strings.
  - Any keyword is accepted as a constraint name after `@`; the checker table decides whether it exists.
  - `@url` parses with net/url and needs a scheme and a host; ports, IPs and localhost are fine; with no arguments only http and https pass, and variadic scheme arguments widen it (`@url("ftp", "s3")`).
  - `@length` counts Unicode code points.
  - Constraint arguments accept unary minus on numeric literals only.
- Docs PR: sentrie-sh/website#80 (branch `146-constraint-checks` on `next`). Sentrie PR: #320.
- _Model: claude-opus-5-5, subagents: 1x sonnet (design questions)_

### 2026-11-10 15:56 - Implement constraint fixes
- Re-reproduced every symptom on the branch with a one-rule pack per case: `@min(-5)` parse error with `{%!s(int=4) ...}`, `@in(1, 2, 3)` wrong arity, `string @in(...)` not found, `@url()` false for `http://localhost:8080/x`, `"é"` passing `@length(2)`, `-3.5` passing `@multiple_of(2)`.
- Keyword constraint names after `@` were already accepted on main (#285, `advanceNameOrKeyword` in `parser/typeref.go`), so the parser needed no change there and does not touch the paths #159 changes.
- Negative arguments: `parseConstraintNegativeLiteral` in `parser/literal.go` folds `-` and the number token into one negative `IntegerLiteral` / `FloatLiteral`, so the index folding, runtime and formatter need no change. `-x`, `-"a"`, `-(1)` get "a minus in a constraint argument must be followed by a number literal". The non-literal and map-key errors print the token `Range` (the garbled `%!s(int=…)` came from `Range.From`).
- `@in` / `@not_in`: variadic on numbers (`NumArgs: -1`) and added to the string table (equality with a listed string). The old list form `@in([1, 2])` is an invalid argument, `in constraint takes each value as its own argument, as in @in(1, 2), not a list`, reported by `sentrie validate` and at runtime; it no longer compares against the list.
- `@url`: `net/url`, scheme and non-empty hostname required; http/https by default; each argument adds a scheme, case-insensitively, so `@url("ftp")` accepts http, https and ftp ("widen" read as adding to the default set). Arguments are checked against the RFC 3986 scheme syntax, so `@url("ftp://")` is an invalid argument. This needed a new arity, `NumArgs: -2` (zero or more), in `constraints` and `ast.validateConstraint`; `make gen` regenerated the table.
- Also from the issue text: `@length` / `@minlength` / `@maxlength` count runes; `@multiple_of` compares the float remainder's magnitude; the email regexp is compiled once at package level.
- Existing tests that encoded the old list form were rewritten to the variadic form (constraints, index, runtime, cmd, lang_test fixtures 0050-in-constraint and 0050-exponent-literals); parse-only parser tests keep the list form since it still parses. `make test`, `make lint`, `make docs-graph` pass.
- _Model: claude-opus-5-5_

### 2026-11-10 16:36 - @url scheme arguments replace the default
- Decision from the user (2026-11-10, supersedes "widen" in the pickup entry): scheme arguments replace the default set. `@url()` accepts http and https; `@url("ftp")` only ftp; `@url("https")` only https; `@url("http", "https", "ftp")` those three. Case-insensitive comparison and the invalid-scheme-argument error are unchanged.
- Changed the checker in constraints/string.go, added tests in constraints/string_test.go, updated docs/codebase-graph/constraints.md and the PR description (open question removed).
- _Model: claude-sonnet-5-5_

### 2026-11-10 16:37 - Model deviation in round 2
- The round-2 `@url` replace-the-default change (a1e0096) was implemented by a sonnet agent after the opus implementer reached its context cap. Constraint checking is an opus area under `.agentrules/parallelize-with-subagents.md`, and a capped agent should be replaced on the same model, so this was a deviation. The commit stays; the opus reviewer checks it with extra care, and any further fix round on this branch goes to opus.
- _Model: claude-opus-5-5 (orchestrator)_

### 2026-11-10 16:48 - Regression tests for the constraint checks
- Added tests in constraints, parser, index, runtime and cmd plus lang_test fixture 0057-constraint-checks (validate); commit a396223. Covers variadic and list-form in/not_in (number and string), url schemes/hosts/invalid scheme argument, rune lengths, negative arguments, multiple_of with negatives, parse error position, keyword constraint names.
- Bite proof against main's production files (literal.go; constraints contraint/number/string; ast typeref and gen) done per file group; every new test fails except where noted in the report.
- Found: map literals in constraint arguments never parse (parser/literal.go uses canExpect for the colon, which does not consume it); present on main, not fixed here.
- Gate: make test, make lint, make docs-graph, gofmt, race on touched packages all pass.
- _Model: claude-sonnet-5-5_
