# 172-1-sentrie-fmt

## Binaek Sarkar (binaek89@gmail.com)

### 2026-11-01 07:02 - Pick up issue #172
- Picked up as part of a wave-2 batch of two issues (174 first slice, 172), each on its own branch and worktree off `main` at 3301d9a.
- Problem: There is no canonical formatter for .sentrie files, so style varies, diffs are noisy and docs examples drift.
- Decision: `sentrie fmt [paths...]` rewrites files in place, with `--check` (non-zero exit when anything would change) and `--diff`. It prints from the AST, so the parser keeps leading and standalone comments as well as trailing ones. A golden test over `lang_test/` and `example_pack/` checks that formatting is idempotent and that the output parses to an equivalent AST. If the change grows past about 800 lines it is split into a stack (comment retention and printer, then the command).
- Work runs through the implement, test and CI, and review loop with separate subagents; the parent session owns origin.

### 2026-11-01 07:34 - Keep every comment and the pipeline source form in the AST
- `CommentStatement` gains `Trailing` (set from the lexer's `TrailingComment` kind) so a printer can put a comment back at the end of the line it followed.
- Policy bodies no longer discard comments after a statement; the next loop iteration parses them as comment statements, which the index already skips (`policyStmtComment`).
- Block bodies also accept a trailing comment as a statement (`{ -- note` used to be a parse error), and shape bodies keep their comments in a new `Cmplx.Comments` list instead of dropping them. A comment as the first item of a shape body used to be a parse error and now parses.
- Decision: the pipeline is desugared at parse time, so a printer could only print `f(x)` for `x |> f()`. `CallExpression` gains `Piped` and `PipeArgs` holding the source form; evaluation keeps reading `Arguments` only. Rewriting every pipeline into nested calls would make the formatter unusable.
- Decision: `TrinaryTypeRef.Boolean` keeps the `boolean` spelling, which the example pack uses; without it every `boolean` would be rewritten to `trinary`.
- ast, parser, lexer, index, runtime, lang_test and loader tests pass unchanged.
- _Model: claude-opus-5-5, role: implementer subagent_

### 2026-11-01 07:47 - Add the format package and golden tests; stop before the command
- New package `format`: `Source` parses, prints, re-parses its own output and requires `Equal` to hold, so a printer bug returns an error instead of rewriting a policy. `Program` prints an AST; `Equal` is a reflective comparison that skips `tokens.Range` fields and treats nil and empty slices alike.
- Style taken from example_pack (two-space indent, `name: Type`, blank line after each top-level declaration) and written down in the package doc comment.
- Decision: line breaks inside lists, maps, call arguments, pipelines, binary expressions and `with`/`attach` clauses follow the source, decided from positions, so output is idempotent. Fully layout-independent output would need a line-width algorithm and would reflow every file.
- Decision: heredocs are recognised from the source bytes at the literal's offset (the token does not say), so they stay heredocs with their tag, and a quoted string with `\n` stays quoted.
- Parser quirks the printer has to respect: the right side of `/` binds at index precedence (`ns/fn(x)`, `a / b.c` is `(a / b).c`), ternaries group to the left, Elvis reads its right side at comparison precedence, and `(a ?` is taken as a typed lambda, so a group that opens with a ternary on a bare name prints `((a) ? ...)`. A trailing comment after an operator expression prints after parentheses, or the parser would attach it to the last operand.
- Comma handling for an item ending in a trailing comment: commas are optional in lists, maps and arguments, so the comma is dropped; when the next item starts with `-`, `+`, `!`, `(`, `[` or `not` it would continue the item, so the comma goes at the start of the next line (checked by rendering the next item on its own).
- Found while testing (pre-existing, not changed): a group such as `(d ? 1 : 2)` in source fails to parse because of the typed-lambda lookahead, and `(x, y?) => ...` is a parse error because the typed path needs `:` or `?` after the first name.
- Size stop rule: ast/parser/format hand-written lines are about 1380 added (format alone about 1310) before step 3, past the 800-line threshold. Per the brief, steps 1, 2 and 4 are finished as a tested layer and the `sentrie fmt` command (step 3) is left for layer 2.
- Tests: format unit and golden tests, parser comment-retention tests; ast, parser, lexer, index, runtime, cmd, lang_test, loader and docsgraph pass; lint 0 issues; docs-graph passes.
- _Model: claude-opus-5-5, role: implementer subagent_

### 2026-11-01 08:20 - Test and CI pass for layer 1
- Added `format/layout_test.go` (comment placement at file end, between policies, comment-only files, export shape; round-trip and stability of ternary chains, Elvis chains, `/` precedence, `not (x in y)`, lambdas in pipelines, heredoc arguments, nullable typed lets, import `with` expressions), two parser tests (consecutive and file-end comments, comments do not change statement counts), `runtime/comment_eval_test.go` (a policy with comments in every position evaluates the same as the comment-stripped policy) and lang_test fixture `0059-comments-everywhere.sentrie` (which the golden corpus also formats). Ran gofmt over `format/format_test.go`, which had a formatting issue.
- Bite proof: with the ast and parser files restored from main the parser, format, lang_test and runtime tests fail (compile errors on `Trailing` and `Comments`); with only the parser files restored the parser tests, the format tests and the 0059 fixture fail.
- Gate: make test, lint (0 issues), gofmt, docs-graph, race on ast, parser, format, index and lang_test all pass.
- Found while testing (all pre-existing on main, not changed): a comment inside a derive body fails validation (`derive body may only contain let declarations`); a trailing comment after a pipeline stage (`x |> f() -- c`) or after a derive lambda is a parse error; `c ? d ? 1 : 2 : e ? 3 : 4` panics the parser (nil condition in `parseTernaryExpression`). New in the formatter: a trailing comment after a comma inside call arguments (`f(1, -- c`) is printed on its own line, so it becomes a leading comment of the next argument (stable, but the placement changes).
- _Model: claude-sonnet-5-5, role: test and CI subagent_

### 2026-11-01 12:08 - Pause before review
- The test and CI gate passed on this layer (tests, lint, graph check, race on ast, parser, format, index and lang_test).
- The review was stopped before it reported, to stay inside the week's usage budget. Nothing on this branch is committed beyond the pickup journal; the queued commits in the parent's queue are not run.
- Open formatter issue for the review loop: a trailing comment after a comma in call arguments (`f(1, -- one` then `2)`) moves onto its own line as a leading comment of the next argument. The program stays equal, but the placement changes.
- Next: rerun the review from scratch, fix its findings and the comment placement above, commit and push this layer, then pick up layer 2 (`sentrie fmt` command) on `172-2-fmt-command` cut from this branch.
- _Model: claude-opus-5-5, role: parent session_

### 2026-11-02 18:15 - Fix the layer 1 review findings
- A comment after a comma (or right after `[`/`(`) is attached by the parser to the next element as a preceding comment. `seq` now prints it back on the previous item's line (`1, -- one`) and the element on the next line, instead of moving the comment onto its own line. Covers call arguments and lists (the 0059 fixture); a comment after a comma in a map is a parse error, so maps never hit it.
- Comment text is now printed from the source bytes at the comment's span, because the lexer trims the content and `Equal` cannot see inner whitespace: `--    indented` keeps its indentation, trailing whitespace is dropped, and `--text` gets one space. This applies to statement, shape, trailing and preceding comments, and falls back to the trimmed content when there is no source.
- Corrected an earlier analysis: a preceding comment does not absorb less than its context. The parser wraps the comment around whatever the context reads next, so `a + (-- c` then `b - 1)` needs its parentheses. `needsParens` now treats a preceding comment as needing parentheses unless it is the last operand, and then exactly when the expression it wraps does. Before this fix, Source rejected that input with "formatted output changes the program".
- `Equal` also compares the merged order of shape fields and comments (`shapeOrder`), since fields are a map and comments a separate slice.
- Smaller fixes: the style list says that a heredoc item drops its comma and that both comma rules depend on the parser's optional commas; the `Program` doc lists every use of `src`; the golden corpus walks directories recursively; nil expressions, statements and types return "cannot print a missing ..." instead of panicking.
- New tests in `format/review_test.go`, one or more per fix; the list-items expectation in `format_test.go` changed to the new comma placement. format, parser, lang_test, runtime, ast and index tests pass; lint 0 issues; docs-graph passes. format.md and the PR description are updated.
- _Model: claude-opus-5-5, role: implementer subagent_
