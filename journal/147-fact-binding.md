## Binaek Sarkar (binaek89@gmail.com)

### 2026-11-10 22:54 - Pick up #147
- Facts are matched by alias instead of name, an omitted optional fact is unreferenceable, a wrongly typed fact returns 500 instead of 400, a wrongly typed optional default passes validate, and the CLI merges facts shallowly (cmd/exec.go:143).
- Picked up in batch 2 of the v0.0.5 milestone (#148, #145, #144, #155, #147), locally while the user is AFK: SSH signing is down, so this stub and its push are queued in the orchestration state file. Paired `sentrie-sh/website` branch of the same name, cut from `next`.
- Decisions recorded with the user on 2026-11-10, which the reviewer judges the change against:
  - For `fact user: User as u`, callers send `user`: the declared name is the external key (consistent with `attach name as expr` and `import x as y`); `u` is the local binding. Clean break with a release note, no alias fallback.
  - An omitted optional fact with no default binds to undefined (if it still reproduces).
  - A wrongly typed optional default fails at validate and at index build.
  - A wrongly typed fact is a 400.
  - CLI `--fact-file` plus `--facts`: deep merge, `--facts` wins, lists are replaced.
- PROVISIONAL (user AFK), needs confirmation: branch stacked on `143-per-rule-errors`, because #143 (a0b04cb) made fact errors of a top-level policy request a request-level error (interim 500) that this fix turns into a 400. PR base `143-per-rule-errors` until #143 merges.
- Docs PR: sentrie-sh/website#84 (draft, base `next`); Sentrie PR #324.
- _Model: claude-opus-5-5_

### 2026-11-10 23:11 - Implement #147
- Reproduced on the base (5818ddd) with a built binary: (1) `fact user: document as u` rejects `{"user": ...}` with `required fact not found: u`; (3) `fact n?: number as n default "x"` passes `sentrie validate` and fails every request with `value x is not a number`; (4) a wrongly typed fact is a 500 on `serve`; (5) `--fact-file` plus `--facts` drops nested keys. Claim (2) did NOT reproduce: an omitted optional fact with no default already reads as undefined (`runtime/eval_ident.go`, a guard added before this branch), so no change there.
- The fact syntax note from #143's test agent: defaults use `default <expr>`, not `= <expr>` (`parser/fact.go`, grammar `'fact' @ident '?'? ':' <type> ('as' @ident)? ('default' <expr>)?`), so `fact n: number = 1` was a plain parse error and does not block the default decision.
- (1) Facts are supplied by declared name and bound by alias. `Policy.Facts` stays alias-keyed (the local identifier space that `eval_ident`, derive purity and the kind checks read); `ExecRule` looks each fact up in the injected map by `Name` and injects it under the alias, and the errors name the declared name. New `Policy.FactByName` (a scan) serves the caller-facing lookups: import `with` at index (`index/import_withs.go`) and runtime (`runtime/imports.go`). `FactNamesByAlias` is replaced by `FactAliasesByName`, and `xerr.ImportUnknownFactMessage` now lists accepted fact names and says when a `with` name is a fact's alias. `AddFact` rejects two facts with the same declared name (they would claim one key). No alias fallback.
- (3) `AddFact` runs `checkFactDefault`: a `null` default for a type that does not accept null, or a literal of the wrong kind (reusing `literalYieldMismatch` from `index/contribute.go`, so a block that yields a literal counts too), is an index error. It runs at index build, so `validate`, `exec`, `bench` and `serve` all reject it before any request. A non-literal default (`["x"][0]`) is still only caught at runtime, as `default value of fact 'n': value x is not a number`, a request-level 500 (the policy's fault).
- (4) `factError` now carries the fact's declared name and whether the value was its default, and reads `fact 'n': value x is not a number`. `asRequestError` wraps a supplied (non-default) fact's type error of the top-level call as `xerr.ErrInvalidInvocation`, so `serve` answers 400 (`fact 'n': value x is not a number: invalid invocation`) for both a named rule and a whole policy; it stays request-level. Checked on `serve`.
- PROVISIONAL (user AFK), needs confirmation: an imported policy's fact failing its type (`import ... with n as "x"`) stays the importing rule's per-rule 500, not a 400, since the binding is the importing policy's, not the caller's. A wrongly typed default evaluated at runtime also stays a 500.
- (5) `cmd.mergeFacts` deep-merges objects, `--facts` wins, lists and other values are replaced. `exec` and `bench` both use it (bench had the same shallow `maps.Copy`).
- Updated tests the decision breaks: `TestImportWithFactsMatchByAlias` became `TestImportWithFactsMatchByName` (the name is accepted, the alias rejected); `TestImportWithByDeclaredNameOfAliasedFactHintsAlias` became `TestImportWithByAliasOfAliasedFactHintsName` with the new message; the "accepts fact aliases" message strings in `index/import_withs_test.go`, `runtime/imports_test.go`, `cmd/validate_test.go` and `xerr/import_with_test.go` now read "accepts facts".
- Codebase graph: index.policy, index.import_withs, runtime.imports, runtime.executor, runtime.eval_ident, runtime (stale fact alias/name nil-deref note), parser.fact, xerr, cmd, api.handle_decision. `make docs-graph` and `make test` pass (`/tmp/sentrie-147-fact-binding-impl-test.log`).
- Signing down: nothing committed; the queued commits are in the report.
- _Model: claude-opus-5-5_

### 2026-11-10 23:27 - Test and CI for #147
- Added tests for every recorded decision. `runtime/fact_binding_test.go`: a fact supplied by declared name binds the alias; the alias alone does not bind (required fact missing, 400; optional fact keeps its default); an omitted optional fact with no default reads as undefined (pinned); a wrongly typed supplied fact is an invalid invocation (400) for a policy and a rule, naming the declared name; a non-literal wrong default stays a request-level 500 (PROVISIONAL), and supplying the fact avoids it; an import `with` binds by declared name at runtime and rejects the alias; an imported policy's wrongly typed fact stays the importing rule's per-rule 500 (PROVISIONAL). `index/fact_binding_test.go`: duplicate declared fact names are an index error, literal-kind and `default null` mismatches fail at index build while matching, nullable, `any` and non-literal defaults pass, plus `FactByName` and `FactAliasesByName`. `api/fact_binding_test.go`: serve answers 200 by name, 400 for alias-only, 400 problem document (no `decisions`, not 207) for a wrongly typed fact on a policy and a rule, 500 for a non-literal wrong default, 207 with a per-rule 500 for the imported wrongly typed fact. `cmd/exec_fact_binding_test.go`: exec by name and not by alias, a wrongly typed fact is an error, `--fact-file` plus `--facts` deep merge (nested keys kept, `--facts` wins, lists replaced, scalars override) for exec and bench, `mergeFacts` replaces mismatched kinds and never mutates its inputs, `validate` rejects wrong literal and null defaults.
- Bite proofs against `143-per-rule-errors` (copy/restore, then mutations where a plain swap does not compile): `runtime/executor.go` and `runtime/rule_error.go` at base fail the runtime, api and cmd tests; the import-with files at base fail the runtime import tests; mutations of the invalid-invocation wrap, the default-is-500 rule, the duplicate-name check, the default check, deep merge, non-mutating merge and bench merge each fail their test. No implementation bug found.
- Gate: `make test`, `make lint` (0 issues), `make docs-graph`, race on api, cmd, index, xerr and runtime, gofmt, em dash check all pass.
- Queued commit: the four test files in one commit before the journal commit. Nothing committed (signing down).
- _Model: claude-sonnet-5-5_

### 2026-11-11 03:58 - Docs for #147
- Website branch 147-fact-binding (PR sentrie-sh/website#84): facts.md (breaking-change caution, declared name vs alias, defaults type-checked, supplying facts), serving-policies.md, exec.md (deep merge), errors.md, running-as-service.md, plus alias statements in exporting-and-importing-rules, policies, policy-composition, running-your-policy, quick-start.
- Verified on the branch binary: declared name binds, alias gives 400 required fact not found, wrong type 400, validate rejects wrong/null default and duplicate names, merge, omitted optional fact undefined. Build passes.
- Post-#147 wording (wrongly typed fact 400) replaces website#79's "500 until #147" lines; merge #79 first and resolve conflicts to the new wording.
- Follow-up: import examples in exporting-and-importing-rules.md and policy-composition.md key `with` by alias and a body inconsistent with the alias; need rewriting for declared names.
- Signing refused, nothing committed in either repo (website changes staged).
- _Model: claude-sonnet-5-5_

### 2026-11-11 10:22 - Validate test fix for #155
- Dropped the positional rule argument from the `runValidateCLI` calls in the validate tests, because #155 makes `sentrie validate` reject a rule target. The tests pass alone and merged with #155.
- _Model: claude-sonnet-5-5, subagents: none, tokens: ~30k_

### 2026-11-11 10:51 - Imported wrongly typed fact is a 400 (user decision 11)
- USER DECISION 11 (2026-11-11, "Make it a 400") supersedes the PROVISIONAL choice of the 2026-11-10 23:11 entry: an imported policy's wrongly typed `with` fact now fails the whole request with a 400, the same as a caller-supplied fact error (consistent with #143's user decision 2, one 400 for fact errors).
- Reuses #143's request-error path: `execRuleFrom` with a caller now runs `asImportedFactError` (`runtime/rule_error.go`), which marks a non-default `factError` of the imported context, or a null supplied for a fact that does not accept null, as a `requestError` wrapping `xerr.ErrInvalidInvocation`. `asRequestError` passes an existing `requestError` through, so the top-level call returns it alone (nil outputs) and `ErrorStatus` gives 400. Nested imports carry it up the same way.
- The message names the imported policy, since the caller did not supply the fact: `imported policy com/ex/target: fact 'n': value x is not a number: invalid invocation`.
- Checked: whole-policy and named-rule requests on serve (400 problem document, no `decisions`), `ExecPolicy`/`ExecRule`, and `sentrie exec` for a rule and a policy (an error, no "rules failed", nothing on stdout, as for a caller's wrongly typed fact). The importing policy's other rule requested alone is still 200.
- Unchanged and still per-rule: a missing required `with` fact (no index check; `ErrRequiredFact` is an invalid invocation, so it is a per-rule 400 inside a 207), an imported default failing its type, and any evaluation error of the imported rule.
- Tests updated (they pinned the per-rule 500): `TestImportedWronglyTypedFactIsPerRule500` became `TestImportedWronglyTypedFactIsARequestError` (policy and rule), `api` `TestDecisionImportedWronglyTypedFactIsPerRule500` became `TestDecisionImportedWronglyTypedFactIs400` (both paths), and `TestExecPolicyFactTypeErrorIsARequestError` expects the importer's request error. Added `TestExecImportedWronglyTypedFactIsAnError` (cmd); it fails with the runtime change reverted.
- Graph: runtime.executor, runtime.imports, api.handle_decision. PR description updated. `make test`, `make lint` (0 issues), `make docs-graph` pass (`/tmp/sentrie-147-fact-binding-r2-test.log`).
- _Model: claude-opus-5-5_

### 2026-11-11 10:55 - Handoff (plan usage limit)
- Done: user decision 11 implemented, tests and graph updated, committed signed (82938ca, c12710e); `make test`, `make lint`, `make docs-graph` passed before the stop.
- Next: the test agent adds regression tests; then docs (website lines on imported `with` fact type errors to 400), re-review, combined check re-run.
- Open findings: a missing required `with` fact for an imported policy is not caught at index build (runtime per-rule 400 inside a 207); `api/fact_binding_test.go` comment on `TestDecisionWronglyTypedNonLiteralDefaultIs500` still says PROVISIONAL although user decision 12 confirmed it.
- _Model: claude-opus-5-5_
