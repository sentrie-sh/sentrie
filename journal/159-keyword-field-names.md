## Binaek Sarkar (binaek89@gmail.com)

### 2026-11-10 15:42 - Pick up #159
- Reserved words cannot be field names, so common JSON keys cannot appear after `.` (parser/access.go:32) or as shape fields (parser/shape.go:106).
- Picked up in batch 1 of the v0.0.5 milestone (#307/#308, #149, #143, #146, #141, #159): one pickup push and a draft PR per branch, plus a paired `sentrie-sh/website` branch of the same name cut from `next`, before implementation.
- Decisions recorded with the user on 2026-11-10, which the reviewer judges the change against:
  - Any keyword is accepted as a name after `.`, as a shape field name and as an attach name.
  - Keywords are not accepted as bare dict keys.
- Docs PR: sentrie-sh/website#82 (branch `159-keyword-field-names` on `next`). Sentrie PR: #322.
- _Model: claude-opus-5-5, subagents: 1x sonnet (design questions)_

### 2026-11-10 16:39 - Accept keywords as field, shape field and attachment names
- Reproduced on b846674 with a scratch pack: `shape S { version: string }` failed with `expected Ident, got version`, and `d.title` failed the same way.
- Reused `advanceNameOrKeyword` (parser/parser.go, already serving module paths, `@` constraint names and FQN segments) at the three positions: the name after `.` (parser/access.go), a shape field name (parser/shape.go) and the name after `attach` (parser/export_rule.go). The error label stays `Ident`, so `x.1` still reports `expected Ident, got Int`. Did not touch parser/literal.go, unary.go or primary.go (branch 146 edits them).
- Ambiguities checked, all resolve to the name reading because right after `.`, at the start of a shape field and after `attach` a keyword was always a parse error, so no valid program changes meaning:
  - `x.in`, `x.not`, `x.is`, `x.true.null` are field accesses; `x.is is defined` reads field `is` then tests it; `2 in [x.in]` and `x.contains contains 7` keep the operators outside the access.
  - `alias.contains(...)` parses as a field access then a call, the same path as `alias.fn(...)`.
  - `attach as as 1` names the attachment `as`. `attach as 1` now reports `expected 'as', got Int` instead of `expected Ident, got as`.
- Bare dict keys unchanged: `{title: 1}` is still rejected (`expected 'yield', got title`), as decided.
- `sentrie fmt` round-trips keyword field, shape field and attach names unchanged.
- Grammar: added `FieldName = IDENT / KEYWORD` (PEG) and `fieldName ::= IDENT | KEYWORD` (EBNF), used in field access, the method part of a function call, shape elements and attach clauses.
- Graph: updated parser.access (the "field names are identifiers only" gotcha was now false), parser.shape, parser.export_rule, parser.parser and grammar. `make docs-graph` passes.
- `go test ./...` green with no test changes; no existing test encoded the old behaviour. Regression tests are left to the test agent.
- _Model: claude-opus-5-5_

### 2026-11-10 16:42 - Regression tests for keyword names
- Added parser suite tests (parser/keyword_field_names_test.go), cmd exec tests for evaluation, attachments and fact validation (cmd/exec_keyword_field_names_test.go), a format round-trip test, and lang_test fixture 0060-keyword-field-names.sentrie.
- Bite proof: with main's access.go, shape.go and export_rule.go every new positive test fails (parser, cmd, lang_test fixture, format incl. corpus stability); the three unchanged-behaviour guards pass on both.
- Method call through an alias with a keyword function name is parse-only: no stdlib function is keyword-named.
- _Model: claude-sonnet-5-5_

### 2026-11-10 16:45 - Docs for keyword names (website#82)
- Added "Keywords as field names" to reference/identifiers.md (example run against the branch binary: `reads` gives "T/9", attachments {"default":5,"version":"1.2"}); updated shapes, rules, policies, index, policy-metadata, types-and-values and policy-composition pages. Replaced the old "user.contribute is a parse error" notes.
- Verified against the binary: `d.title`, `obj.true.null`, `d["a-b"]`, `x.is is defined`, `2 in [x.in]`, `x.contains contains 7`; `{title: 1}` and comma-separated shape fields and `rule let` still fail to parse.
- Website build passes (/tmp/website-159-keyword-field-names-build.log). Website commit 53ef399.
- _Model: claude-sonnet-5-5_
