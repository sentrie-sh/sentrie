---
name: issue 77 lambda builtins
overview: Replace special collection keywords/forms with built-in higher-order functions and inline block-bodied lambdas. The decisive work is callable-aware builtin execution and boundary-safe boxed values; parser and keyword changes are necessary but secondary. Includes explicit arity/capture/boundary contracts, distinct key semantics, repo + website docs, and **≥90% test coverage on new/updated Sentrie Go code** (Coverage bar). Tracks [sentrie-sh/sentrie#77](https://github.com/sentrie-sh/sentrie/issues/77).
todos:
  - id: lexer-token-arrow
    content: Add => tokenization; remove collection words from keyword table early so downstream parse/call paths see them as identifiers.
    status: pending
  - id: callable-arity-contract
    content: "Define and document arity rules for higher-order builtins and callables: extra/missing args, filter/map/distinct 1 vs 2 param lambdas, reduce 2 vs 3 param lambdas; duplicate lambda params forbidden; empty param list policy; enforce at runtime minimum, validate at parse time where cheap."
    status: pending
  - id: parser-lambda-disambiguation
    content: Implement grouped-vs-lambda parsing, LambdaExpression AST, and parser validation (identifiers-only params, no duplicates, empty () policy per contract).
    status: pending
  - id: lambda-capture-semantics
    content: "Document and implement v1 lambda capture: closure holds parent execution context by reference (not a snapshot of resolved locals at creation time)."
    status: pending
  - id: runtime-callable-value
    content: Add boxed callable representation, runtime callable abstraction (arity, invoke, trace), and lambda eval returning callable per capture contract.
    status: pending
  - id: builtin-hof-refactor
    content: "Hinge: refactor eval_call/builtins so higher-order builtins consume boxed values and invoke callables natively; replace Builtin with func(context.Context, ...box.Value) (box.Value, error); boundary convert only at true external seams."
    status: pending
  - id: boundary-rejection
    content: Reject callable values at non-native boundaries (module/JS, ToBoundaryAny/FromBoundaryAny, any []any-only paths) with targeted errors—never silent stringify, nil, or drop.
    status: pending
  - id: collection-builtins
    content: Implement filter/map/reduce/any/all/first/distinct as higher-order built-ins per arity contract and distinct key rules.
    status: pending
  - id: remove-legacy-paths
    content: Delete special collection parser/AST/runtime forms and dispatch; do not maintain dual systems.
    status: pending
  - id: tests-and-repo-docs
    content: Add lexer/parser/runtime/integration tests (incl. aggressive distinct matrix, named lambda capture regression test per §9); meet ≥90% coverage on new/updated Sentrie Go code (Coverage bar); update in-repo grammar (grammar.peg/ebnf) and migration/changelog notes.
    status: pending
  - id: website-user-docs
    content: Update website reference, homepage/feature/quick examples, and any marketing snippets that show list operations; migration callouts.
    status: pending
isProject: false
---

# Issue #77: Replace collection keywords with built-in higher-order calls and inline lambdas

## Scope and goals
- Remove grammar-level special forms for `any`, `all`, `filter`, `first`, `map`, `reduce`, and `distinct`.
- Add inline block-bodied lambda expressions using `=>`.
- Make higher-order collection operations ordinary built-in calls that accept callable runtime values.
- Remove old parser/AST/runtime special-form paths and migrate coverage—**no dragging two systems**.
- **Documentation is in scope for this issue:** ship updated end-user reference and examples alongside the language change (see [issue #77](https://github.com/sentrie-sh/sentrie/issues/77) acceptance criteria and compatibility section).

## Coverage bar (Sentrie repo, Go)
- **Target:** **≥90%** statement (or line) coverage for **all new Go code** and **all Go code materially changed** by this work (`lexer`, `tokens`, `parser`, `ast`, `box`, `runtime`, and any new packages/files for lambdas/callables/builtins).
- **How to verify:** `go test` with coverage on affected packages (e.g. `-coverprofile` or the repo’s coverage script/CI); use **90% as a merge gate** for those packages.
- **Scope:** applies to **Go implementation** only (not docs/grammar prose). If CI already enforces a minimum, **meet the stricter of this bar and repo defaults**.

## Central hinge (what actually decides success)
- **Parser and keyword work are noisy but not the core risk.** The decisive shift is **moving from special-form evaluation to callable-aware builtin execution**: boxed callables, native invoke from builtins, and **no** smuggling AST through `[]any` boundaries.
- **Strong sequencing choices:** remove collection words from the keyword table **early**; treat lambda as a **boxed callable value**; **delete** legacy evaluator/parser paths once parity exists; keep **`distinct` key semantics** as a first-class design area, not an afterthought inside builtin work.

## Current architecture anchors
- Keyword/tokenization is centered in [tokens/token_kind.go](tokens/token_kind.go) and [lexer/lexer.go](lexer/lexer.go).
- Prefix parser registration for collection special forms is in [parser/lookups.go](parser/lookups.go), with dedicated parsers in [parser/quantifier.go](parser/quantifier.go), [parser/reduce.go](parser/reduce.go), and [parser/distinct.go](parser/distinct.go).
- Grouping logic lives in [parser/block.go](parser/block.go), and calls are parsed in [parser/call.go](parser/call.go).
- Runtime dispatch/evaluation of special forms is in [runtime/eval.go](runtime/eval.go), [runtime/eval_quantifier.go](runtime/eval_quantifier.go), [runtime/eval_reduce.go](runtime/eval_reduce.go), and [runtime/eval_distinct.go](runtime/eval_distinct.go).
- Builtin call resolution currently uses boundary conversion in [runtime/eval_call.go](runtime/eval_call.go) and [runtime/builtins.go](runtime/builtins.go).
- Value representation is in [box/value.go](box/value.go).
- Today **`Builtin`** in [runtime/builtins.go](runtime/builtins.go) is `func(ctx context.Context, args []any) (any, error)`—that shape must move to **`func(ctx context.Context, args ...box.Value) (box.Value, error)`** (variadic args, boxed return) so callables and lists/maps are not forced through `[]any` before builtin bodies run. This matches **`getTarget`** / **`wrappedTarget`** in [runtime/eval_call.go](runtime/eval_call.go), which already call targets as `target(ctx, args...)`.

## Planned code updates (Sentrie repo)

Paths are relative to the **sentrie** repository root. **Add** = new file; **Edit** = substantive change; **Delete** = remove after replacement is live.

### `tokens/`
- **Edit** [tokens/token_kind.go](tokens/token_kind.go): add token kind for `=>`; remove `any`, `all`, `filter`, `first`, `map`, `reduce`, `distinct` from the keyword map so they lex as identifiers.

### `lexer/`
- **Edit** [lexer/lexer.go](lexer/lexer.go): recognize `=>` as a single token (before `=` / `==` split); keep `(` / `)` behavior otherwise.

### `parser/`
- **Edit** [parser/lookups.go](parser/lookups.go): drop prefix registrations for collection keywords; wire lambda / grouped-or-lambda as needed.
- **Edit** [parser/block.go](parser/block.go): disambiguate grouped expression vs lambda parameter list + `=>`; enforce invalid `(a, b)` without `=>`.
- **Add** `parser/lambda.go` (or similar): parse `(id, …) => { … }` into `LambdaExpression`; validate identifier-only params, no duplicates, allow `()`.
- **Edit** [parser/call.go](parser/call.go) only if call / grouping interaction requires it (precedence, edge cases).
- **Delete** after migration: [parser/quantifier.go](parser/quantifier.go), [parser/reduce.go](parser/reduce.go), [parser/distinct.go](parser/distinct.go).

### `ast/`
- **Add** `ast/lambda.go` (or similar): `LambdaExpression` with `Params []string`, `Body *BlockExpression` (or agreed block type).
- **Edit** [ast/quantifier.go](ast/quantifier.go), [ast/reduce.go](ast/reduce.go): remove `AnyExpression`, `AllExpression`, `FilterExpression`, `FirstExpression`, `MapExpression`, `DistinctExpression`, `ReduceExpression` (or shrink files to zero and delete).
- **Edit** any **String / FQN / transform / codegen** helpers that switch on removed node kinds (e.g. [ast/node.go](ast/node.go) consumers, [ast/gen.go](ast/gen.go) if applicable).

### `box/`
- **Edit** [box/value.go](box/value.go): new `ValueKind` for callables; constructors/accessors; ensure `ToBoundaryAny` / `FromBoundaryAny` **reject or error** on callables per boundary policy.
- **Edit** other `box` helpers if equality/hash for `distinct` keys lives here or is shared with memoization.

### `runtime/` (core)
- **Edit** [runtime/eval.go](runtime/eval.go): remove `evalAny`, `evalFilter`, … dispatch branches for deleted AST types; add **`evalLambda`** (or equivalent) producing a boxed callable.
- **Delete** after parity: [runtime/eval_quantifier.go](runtime/eval_quantifier.go), [runtime/eval_reduce.go](runtime/eval_reduce.go), [runtime/eval_distinct.go](runtime/eval_distinct.go) (if present as separate file).
- **Add** `runtime/callable.go` (name TBD): interface or struct for arity + `Invoke(ctx, ec, args ...box.Value) (box.Value, error)` (or an internal slice API that forwards to the same semantics) + trace hooks—**stay consistent with variadic `Builtin`** at public invoke sites.
- **Add** `runtime/eval_lambda.go` (name TBD): build callable from `LambdaExpression` + parent execution context (capture-by-reference v1).

### `runtime/` (calls, builtins, modules)
- **Edit** [runtime/builtins.go](runtime/builtins.go):
  - Change **`type Builtin`** to **`func(ctx context.Context, args ...box.Value) (box.Value, error)`** (variadic—**decided**).
  - Rewrite **every** `Builtin*` implementation: use **`args` as a variadic slice** (`args[0]`, `len(args)`, etc.) and **`box.Value`** accessors (`ListValue`, `NumberValue`, etc.).
  - Register **`filter`**, **`map`**, **`reduce`**, **`any`**, **`all`**, **`first`**, **`distinct`** in `Builtins` map as higher-order implementations.
  - Remove or rewrite helpers like `isUndefinedAny` / `toIntAny` if they become **`box.Value`**-first.
- **Edit** [runtime/eval_call.go](runtime/eval_call.go):
  - **`evalCall`** already builds **`[]box.Value`** for arguments; keep that.
  - **`getTarget`**: for builtins, **stop** converting args with `ToBoundaryAny` before invoking; invoke with **`builtin(ctx, args...)`** (variadic spread from the evaluated arg slice).
  - For **module** `modulebinding.Call`, keep `[]any` only if the JS boundary stays that way; **reject** callable **`box.Value`** before conversion with a **targeted** error (no silent drop).
  - **`calculateHashKey`**: today hashes `box.ToBoundaryAny` per arg—**update** so callables do not produce misleading hashes (reject, or document-only path).
- **Edit** [runtime/modules.go](runtime/modules.go) (and related) if `Call` signature or argument marshaling must explicitly forbid callables.

### Tests and fixtures
- **Edit** [lexer/lexer_test.go](lexer/lexer_test.go) (and related) for `=>` and demoted keywords.
- **Edit** [parser/expression_test.go](parser/expression_test.go), [parser/precedence_test.go](parser/precedence_test.go) as needed for lambdas and calls.
- **Edit** [runtime/eval_call_test.go](runtime/eval_call_test.go), [runtime/eval_chain_test.go](runtime/eval_chain_test.go), [runtime/eval_distinct_test.go](runtime/eval_distinct_test.go), [runtime/eval_branches_test.go](runtime/eval_branches_test.go), [runtime/eval_chain_gap_wave2_test.go](runtime/eval_chain_gap_wave2_test.go): replace `[]any` builtin stubs with **`box.Value`** where applicable; add arity/boundary/`distinct` matrix coverage; add the **named lambda capture** test (late-bound lexical / parent-context-by-reference—see **§9**).
- **Edit** `lang_test/*.sentrie` that use old syntax (e.g. [lang_test/0016-comprehensions-aggregations.sentrie](lang_test/0016-comprehensions-aggregations.sentrie), [lang_test/0047-quantifier-grouped-expr.sentrie](lang_test/0047-quantifier-grouped-expr.sentrie)) to **builtin + lambda** forms.

### Grammar (in-repo spec)
- **Edit** [grammar/grammar.peg](grammar/grammar.peg), [grammar/grammar.ebnf](grammar/grammar.ebnf): lambda, grouped-or-lambda, remove special collection productions.

### Other call sites to grep while implementing
- Any **`Builtins[`** test override, **`ToBoundaryAny`/`FromBoundaryAny`** on eval paths, **`ast.*Expression`** type switches, and **trace** stringification of callee/args—ensure they tolerate or explicitly handle **`ValueCallable`**.

## Callable arity contract (design — write down before coding)
Define explicitly (and enforce—**do not improvise mid-implementation**):
- **Extra arguments** to a callable invocation: error.
- **Missing arguments**: error.
- **`filter` / `map` / `any` / `all` / `first` / selector `distinct`**: callable must be **arity 1 or 2** (item-only vs item+index); builtin picks iteration shape accordingly.
- **`reduce`**: callable must be **arity 2 or 3** (acc+item vs acc+item+idx).
- **`distinct(list)`**: no callable; **`distinct(list, callable(item|item, idx))`** (contract notation): same 1/2 arity rule as other item iterators.
- **Where enforced**: at minimum **runtime** on every invoke; add **parse-time** checks only where cheap and unambiguous (e.g. duplicate parameter names, malformed param list).
- **Lambda parameter list**: **no duplicate** parameter names; **parameters are bare identifiers only** (already in issue).

## Lambda parameter list (parser — decide and enforce)
- **Only** bare identifiers; reject anything else at parse time with a clear message.
- **No duplicate** names in the parameter list.
- **Empty parameter list `() => { ... }`**: **allow in the grammar** for a uniform “param list” model and forward compatibility; higher-order builtins still **reject** arity mismatch at runtime when the callable is **used**. Document that empty-arity lambdas are syntactically valid but not usable as collection callbacks until something needs arity 0.

## Lambda capture semantics (v1 — explicit)
- **v1 choice: capture the parent execution context by reference** (closure links to the creating scope’s execution context / environment), **not** a snapshot of all resolved lexical values at lambda creation time. Those models diverge for mutability and late-bound names; parent-context-by-reference is the simpler, explicit v1 contract.
- Document this in code comments and contributor-facing notes so implementers do not accidentally build the other model.
- **Regression lock:** add at least one **named** runtime or integration test (see **§9 Test matrix**) that proves **late-bound lexical resolution** follows the **live parent context**, so snapshot capture cannot slip in unnoticed.

## Boundary rejection (hard rule)
- **Callables must not cross** `[]any` / `ToBoundaryAny` / `FromBoundaryAny` / module–JS (or any non-native) paths.
- **Reject with a targeted runtime error** if a callable is forced through such a boundary. **Do not** silently stringify, coerce to nil, or drop— that produces cursed, undebuggable behavior.
- Implement this as part of the **builtin/call refactor** and audit all boundary sites once `ValueCallable` exists.

## Implementation plan

### 1) Lexer and token model for lambdas
- Add a token kind for `=>` in [tokens/token_kind.go](tokens/token_kind.go).
- Update `=` lexing in [lexer/lexer.go](lexer/lexer.go) to emit arrow token before plain assignment/equality paths.
- Remove `any/all/filter/first/map/reduce/distinct` from keyword lookup so they lex as identifiers (**early**, consistent with the issue).
- Preserve all existing behavior for grouped expressions and normal calls.

### 2) AST: introduce lambda node and retire special-form nodes
- Add `LambdaExpression` AST node (params + block body), likely in a dedicated file under `ast/`.
- Remove/deprecate `AnyExpression`, `AllExpression`, `FilterExpression`, `FirstExpression`, `MapExpression`, `ReduceExpression`, `DistinctExpression` once parser/runtime migration is complete.

### 3) Parser: grouped-or-lambda disambiguation and call-based collection syntax
- Extend `(` parsing in [parser/block.go](parser/block.go) to disambiguate:
  - grouped expression `(x + 1)`
  - lambda parameter list `(item)` / `(a, b)` followed by `=>`
- Per issue: **`(a, b)` without `=>` remains invalid** (not valid grouping).
- **Explicit parser rules:** parameters **identifiers only**; **duplicate names rejected**; **empty `()`** allowed per **Lambda parameter list** above—state these in §3 so implementers do not infer divergent behavior.
- Add lambda parser entrypoint (new parser file), producing `LambdaExpression` with block body; apply **parameter constraints** above.
- Remove special prefix parser registrations from [parser/lookups.go](parser/lookups.go) for old collection keywords.
- Delete old special parser files (`quantifier`, `reduce`, `distinct`) once call-based parsing fully replaces them.
- Ensure ordinary call parsing in [parser/call.go](parser/call.go) remains unchanged for non-lambda scenarios.

- **Breaking change, normal parse failures:** remove old collection grammar entirely; **do not** invest in curated parser errors for dead `filter … as …` / `reduce … from …` / `distinct … as …` syntax. **Migration story lives in docs/changelog only.**

### 4) Runtime callable model
- Add callable value kind/support in [box/value.go](box/value.go) for runtime-native callable values.
- Introduce callable abstraction in `runtime/` (**arity**, **invoke with boxed args**, **trace hooks**).
- Implement lambda evaluation (new runtime file) per **lambda capture semantics (v1)**; return boxed callable values.
- Update evaluator dispatch in [runtime/eval.go](runtime/eval.go) to evaluate lambda AST.

### 5) Call evaluation + builtin invocation refactor (hinge)
- Replace the **`Builtin`** type and all builtin implementations in [runtime/builtins.go](runtime/builtins.go): **`type Builtin func(context.Context, ...box.Value) (box.Value, error)`**, aligned with [box/value.go](box/value.go). **Do not** keep `[]any` as the primary builtin wire format; use `ToBoundaryAny` / `FromBoundaryAny` only at edges that truly cross the non-Sentrie boundary, not inside every builtin.
- Refactor [runtime/eval_call.go](runtime/eval_call.go) so **`getTarget`** invokes builtins as **`builtin(ctx, args...)`** and returns **`box.Value`** without per-arg `ToBoundaryAny` for the builtin path; higher-order builtins then receive **boxed values**, invoke **callables** through the native abstraction, and **boundary rules** stay explicit.
- **Operational boundary rule:** non-native runtime boundaries (including module/JS interop and any `[]any` conversion path) must **reject** callable boxed values with **targeted errors**; **module calls must not** silently accept, stringify, or nil-out callables.
- Keep module-call and non-higher-order builtin behavior stable except where **boundary rejection** must apply to callables.

### 6) Re-implement collection operations as built-ins
- Implement builtin forms per **callable arity contract** (**`callable(...)` is contract notation** for a boxed callable with that arity shape—not a keyword and **not** a function named `callable`; at the source level the argument is an inline lambda such as `(item) => { yield … }`):
  - `filter(list, callable(item|item, idx))`
  - `map(list, callable(item|item, idx))`
  - `any(list, callable(item|item, idx))`
  - `all(list, callable(item|item, idx))`
  - `first(list, callable(item|item, idx))`
  - `reduce(list, initial, callable(acc, item|acc, item, idx))`
  - `distinct(list)` and `distinct(list, callable(item|item, idx))`
- Validate argument **counts** and **collection types** with targeted runtime errors.
- Preserve deterministic output ordering and explicit key semantics for `distinct`.

### 7) Distinct key semantics hardening
- **v1 minimum:** support dedupe keys derived from **scalar** boxed values at least: **`string`**, **`number`**, **`boolean`**, **`trinary`**, and **`null` / `undefined`** if the value model treats them as distinct kinds. Anything else (e.g. list, map, document as keys) requires an **explicit** stable canonical strategy or a **targeted** “unsupported key kind” error—no ambiguous behavior.
- Document supported key kinds in runtime comments/errors; ensure **direct** `distinct(list)` and **selector** `distinct(list, callable(...))` use the **same** key semantics.
- Reject unsupported key kinds with explicit runtime errors instead of implicit/inconsistent behavior.

### 8) Remove obsolete evaluator paths
- Remove special-form evaluator files and dispatch entries after parity is reached.
- Remove any dead AST/parser paths and update references.

### 9) Test matrix
- **Coverage:** drive **≥90%** coverage on new/updated Go code (see **Coverage bar**); include error branches (arity, boundaries, `distinct` keys), not only happy paths.
- **Lexer:** arrow tokenization; collection words lex as identifiers.
- **Parser:** lambda single/multi-param; grouped-vs-lambda; invalid `(a, b)` without `=>`; **duplicate lambda params** and **non-identifier params** rejected; **`() =>`** accepted; call syntax for higher-order builtins.
- **Runtime / builtins:** callable boxing; **arity mismatch** (wrong callable arity for each of `filter`/`map`/`any`/`all`/`first`/`reduce`/selector-`distinct`); each higher-order builtin success + builtin argument-count failures; trace expectations for lambda invocation and per-item outcomes.
- **Lambda capture (contract test):** one explicit test, e.g. **`TestLambdaCapture_LateBoundLexicalFollowsParentContextByReference`** (name can vary—intent must not): a lexical binding is **assigned or updated after** the lambda value is created; the lambda body (invoked via a higher-order builtin) must observe the **current** value, proving **reference** semantics and ruling out **snapshot-at-creation** capture.
- **`distinct` (explicit, not hand-wavy):**
  - **Direct `distinct(list)`:** scalar elements dedupe correctly; **unsupported element kinds** fail with a **clear** error if v1 rejects non-scalars.
  - **Selector form:** **scalar** selector results; **duplicate selector keys** keep **first** occurrence order; **unsupported selector key kinds** fail clearly.
  - Align tests with the same key/equality rules for direct vs selector paths.
- **Integration:** parse + eval end-to-end snippets using new syntax only.

### 10) In-repository language documentation
- Update grammar specs in the Sentrie repo: [grammar/grammar.peg](grammar/grammar.peg) and [grammar/grammar.ebnf](grammar/grammar.ebnf) (lambda, grouped-or-lambda, removal of special collection productions).
- Add or extend in-repo migration notes if the project keeps them (e.g. changelog or language notes), describing the breaking removal of `filter list as …` / `reduce … from … as …` / `distinct … as …` forms.

### 11) Website / user-facing documentation (separate repo: `website`)
Today the public docs still describe the **old** keyword/special-form collection syntax. Update them so they match the new model from [issue #77](https://github.com/sentrie-sh/sentrie/issues/77).

Paths below are **relative to the website repository root**.

- **[src/content/docs/reference/collection-operations.md](src/content/docs/reference/collection-operations.md)** — primary page: builtin call + lambda examples; `=>` rules; `distinct(list)` vs selector form and v1 key/error story.
- **[src/content/docs/reference/let.md](src/content/docs/reference/let.md)** — rewrite `reduce` examples to builtin + lambda form.
- **[src/content/docs/reference/index.md](src/content/docs/reference/index.md)** — refresh keyword/builtin listings; link to lambda and collection builtins.
- **Homepage, landing page, feature pages, quickstart, and any marketing or hero snippets** that mention list/collection operations — grep beyond `src/content/docs` if the site keeps examples in `src/app`, marketing MDX, or shared components; **no stale special-form syntax** in user-visible copy.
- **Search and fix** — grep for old patterns: `filter … as`, `map … as`, `reduce … from`, `distinct … as`, quantifier-style docs; update cross-links.
- **Explicit breaking-change callout** — “Migrating from …” with before/after for at least `filter`, `reduce`, and `distinct`.
- **Trace / explainability** — if documented, align with higher-order builtin + per-invocation traces.

Ship website doc updates in the same release window as the language change so readers are not pointed at removed syntax.

## Sequencing / rollout checkpoints
- Milestone A: `=>` + **keyword demotion** + parser lambda/grouping + **arity/param validation** documented and tested; **≥90%** coverage on touched lexer/parser/ast (or incremental path to C).
- Milestone B: **callable value + capture semantics + builtin/boundary refactor** + higher-order builtins + **boundary rejection** wired; **≥90%** on changed runtime/box/call code.
- Milestone C: legacy paths **removed**; Sentrie grammar/migration notes updated; **website reference + marketing/quick examples** updated; full test matrix green; **≥90%** on all new/updated Go code for this change set.

## Risks to watch
- Parser ambiguity around `(` and call precedence; enforce explicit tests for edge cases.
- Accidentally implementing **snapshot** capture instead of **parent-context-by-reference** for v1.
- Callable leakage or silent coercion at boundaries—treat as **hard failure** only.
- `distinct` key semantics for non-scalars; prefer explicit rejection in v1 if canonical hashing is not already stable.

## Validation checklist
- **Coverage:** new/updated Sentrie Go code meets **≥90%** test coverage (per **Coverage bar**); CI or `go test -cover` confirms.
- `=>` tokenizes correctly.
- Collection names are no longer reserved keywords.
- Lambdas parse and evaluate as callable values per **arity contract** and **capture semantics**; the **named late-bound lexical / parent-context-by-reference** test (§9) passes and guards against snapshot capture regressions.
- Higher-order builtins work via ordinary call syntax; **legacy AST/runtime special forms are gone** (no dual system).
- **`Builtin`** is **`func(context.Context, ...box.Value) (box.Value, error)`**; builtin bodies and **`getTarget`** use variadic spread—no `[]any`/`any` as the primary contract; boundary conversion only at true external seams.
- **Callables rejected** at unsupported boundaries with **targeted** errors.
- Existing non-collection call semantics remain stable where callables are not involved.
- Lexer/parser/runtime/integration tests cover success and failure paths, including the **`distinct` matrix** above.
- **Docs:** grammar.peg/ebnf + website (**reference and** homepage/landing/feature/quick examples) describe the new syntax only; breaking-change migration is documented (changelog/migration notes)—**no** dedicated parser paths for obsolete collection syntax.
