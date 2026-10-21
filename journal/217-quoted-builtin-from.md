## Binaek Sarkar (binaek89@gmail.com)

### 22:04 - Pick up #217
- Branch cut from `216-use-span-end`. The new error sits in `parseUseStatement` and points at the quoted string, whose span #216 just made trustworthy, and the fixtures it fixes are the ones #214 touched.
- _Model: claude-opus-5-5_

### 22:07 - Reject quoted @ paths at parse time
- Decision (given by the parent, option 1 of the issue): a quoted `from` whose value starts with `@` is an error. Parse time is the earliest stage and has the best span (the string token itself, end rebuilt with #216's `tokenEnd`), so the check sits in `parseUseStatement` before the string is consumed. Index never sees the statement.
- Added `Parser.errorAtf(rnge, ...)`; `errorf` now delegates to it with the current range. Needed because the lexer end of a string followed by a newline is the next line's column 0, which `errorf` would print.
- Message: `quoted module path "@sentrie/js" is read as a file next to the policy; drop the quotes to import the module: from @sentrie/js`.
- Scope of the check is a leading `@` only. `"./@util.ts"`, `"lib/@scope.ts"` and `" @sentrie/js"` still parse as file paths; a file whose name starts with `@` is reachable as `"./@name"`. Tests cover both sides.
- Fixtures: `rg -n 'from\s+"@'` over the tree found 0043, 0044, 0045 (two lines, one of them `@sentrie/json`) and 0046. No Go test, example pack or doc used the quoted form (`example_pack/user.ts` has a TS `import ... from "@local/something"`, which is TypeScript, not a `use`). Unquoted all of them and added SPDX headers to the three that had none. Committed before the rejection so every commit passes `lang_test`.
- Grammar files (`grammar.ebnf`, `grammar.peg`) gained a comment noting the rule; the productions are unchanged.
- **Breaking:** a pack that spells a module as `from "@sentrie/..."` (or any quoted `@...`) now fails to parse. Before, it parsed, indexed and validated, then failed at first evaluation looking for a file named `@sentrie/...` next to the policy, so no working pack can depend on the old behaviour.
- Not done here: the issue's "have `sentrie validate` check that every local `use` path exists" belongs to #155, which the issue cites for it.
- _Model: claude-opus-5-5_
