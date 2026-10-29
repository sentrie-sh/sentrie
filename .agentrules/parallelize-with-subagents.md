# Parallelize with Subagents

Prefer spawning subagents to run independent parts of a task concurrently instead of doing everything serially. Choose each subagent's model prudently, and brief it tightly.

**Why:** Parallel subagents finish multi-part work faster, and matching the model to the subtask keeps cost down. But under-powering a task that needs a strong model wastes the whole run when the output is wrong, so accuracy comes first.

**How to apply:**

- When a task splits into parts that do not depend on each other, hand them to separate subagents that run at the same time rather than working through them one by one.
- Pick the cheapest model that will still be accurate for that specific subtask: a lighter model for mechanical search, collation, or boilerplate; a stronger model for reasoning-heavy, ambiguous, or correctness-critical work.
- Economy is the goal, but never at the cost of accuracy. When unsure whether a lighter model is good enough, use the more capable one.

## Every subagent is its own bill

A subagent does not share the parent's context. It starts cold, makes its own requests, re-reads whatever it needs, and is billed for all of it. Fanning out to five agents pays for five cold starts. So:

- Spawn only when parallelism, or keeping bulk reading out of the main context, pays for that cold start. A question two `rg` calls can answer does not need a subagent.
- For a follow-up on work a subagent already did, continue that agent (it keeps its context) instead of spawning a fresh one that re-derives it.

## Set the model on every spawn

Pass `model` explicitly on every Agent tool call, including the built-in types (`Explore`, `Plan`, `general-purpose`). Left unset, a subagent takes its definition's model, and failing that the parent's, which is usually the most expensive one in play.

### Cost of each tier

| `model` | Model | Input / output per 1M tokens | Relative cost | Context |
|---|---|---|---|---|
| `haiku` | Claude Haiku 5.5 | $0.10 / $0.50 (prompt up to 100K tokens); $0.50 / $2.50 (prompt over 100K) | 1x; 5x over 100K | 1M |
| `sonnet` | Claude Sonnet 5.5 | $2 / $10 | 20x | 1M |
| `opus` | Claude Opus 5.5 | $4 / $20 | 40x | 1M |
| `fable` | Claude Fable 5.1 | $10 / $50 | 100x | 1M |

Haiku 5.5 is priced by prompt length. A request whose prompt is over 100K tokens pays the higher price, counting every input token, cache reads included. A subagent's prompt grows with every turn, so a long-running Haiku job crosses the line partway through. Even above the line it costs a quarter of `sonnet`.

Prices as of 2026-10-13, from Anthropic's published models overview and pricing pages. When the lineup changes, update this table in its own `[AgentRule]` PR; the ratios are what drive the choice.

### Which model for which job

| Job in this repo | `model` |
|---|---|
| Find the callers, implementers or importers of a named symbol | `haiku` |
| Sweep `docs/codebase-graph/` for inbound edges to a node | `haiku` |
| List files, `lang_test` fixtures, journal entries or graph nodes matching a pattern | `haiku` |
| Run a narrowed `go test` or a named `make` target and report pass or fail with the log path | `haiku` |
| Collate `gh pr list`, `gh issue view` or label state into a short list | `haiku` |
| Check a diff for one mechanical rule: em dashes, AI signatures, missing or outdated license headers, a `PR_DESCRIPTION_*.md` staged for commit | `haiku` |
| Summarize one package from its graph node plus sources | `sonnet` |
| Implement a change specified down to the files and functions | `sonnet` |
| Write testify suite cases for a stated contract (the `test-writer` brief in `.cursor/agents/test-writer.md`) | `sonnet` |
| Add `lang_test` fixtures for a stated language change | `sonnet` |
| Update a `docs/codebase-graph/` node to match a finished change | `sonnet` |
| Draft a journal entry or PR description from a diff | `sonnet` |
| Review a diff for one named class of problem that needs reading, not matching | `sonnet` |
| Race and guard audits in `runtime.executor`, `runtime.exec_ctx` or the call memo ([Trace data flow for race/guard audits](trace-data-flow-for-races.md)) | `opus` |
| Changes where foreign code or untrusted input crosses in: `runtime.js`, `runtime.modules`, `api.handle_decision`, `@local` module path confinement | `opus` |
| Language semantics: grammar, parser precedence, trinary logic, typeref and constraint checking | `opus` |
| Derive purity and `evalDerive` resolution, and the paired static and runtime checks (`index.builtin_check` with `runtime.eval_call`, `index.derive_purity` with the runtime purity gates) ([`evalDerive` resolution](evalderive-resolution.md)) | `opus` |
| Design where the requirement is ambiguous or the approach is not yet chosen | `opus` |
| Whole-change review before a PR leaves draft (the `code-reviewer` brief in `.cursor/agents/code-reviewer.md`) | `opus` |
| Anything, only when the user asks for it by name | `fable` |

The briefs in `.cursor/agents/` (`code-reviewer`, `test-writer`, `config-writer`) are reusable prompt bases. Spawning one still means passing `model` from this table.

### Rules for jobs not in the table

- **Haiku only when the parent can check the answer cheaply.** A list of `file:line` hits the parent can spot-check is Haiku work. An answer the parent would have to take on trust is not.
- **Keep a Haiku job under 100K tokens.** Its context is 1M, so size no longer rules Haiku out, but its price steps up 5x over 100K. Scope the prompt and the reading so the job finishes under that line, for example one package or one pattern per spawn. A job that must read more stays on `haiku` only while the answer is still one the parent can check; otherwise it goes to `sonnet`.
- **Code that ships is at least `sonnet`.** Haiku may locate and report, never author a change that merges.
- **The `opus` areas above stay `opus`** however small the edit looks: a one-line change to a precedence table or a purity gate is still a correctness-critical change.
- **A reviewer is at least the tier of the author** of what it reviews.
- **Never `fable` by default.** It costs 2.5x `opus`. Use it only when the user asks for it.
- **Between two tiers, take the higher one.** Accuracy first, as above.

### When a cheap subagent gets it wrong

- Rerun the same prompt one tier up. Do not retry on the same tier, and do not escalate more than once: after that, do the work in the parent, or stop and report per the two-attempt limit in [Token efficiency](token-efficiency.md).
- First check whether the prompt was the problem. A vague prompt fails on every tier, and fixing it is cheaper than escalating.

### Record what ran

Name each subagent's model in the journal entry's model bullet, for example `_Model: claude-opus-5-5, subagents: 3x haiku, 1x sonnet, tokens: ~40k_` ([Journal convention](journal-convention.md)). That keeps the choices reviewable and shows when a tier is being over- or under-used.

## Tighten the prompt

A loose prompt makes the subagent spend tokens rediscovering what the parent already knows, then return more than the parent needs.

- **Self-contained:** state the goal, what is already known and what is ruled out, so the agent does not re-derive it. Give exact paths and graph node ids (`docs/codebase-graph/runtime.executor.md`), not "the executor code".
- **Scoped:** name the packages or files in bounds, what not to touch, and when to stop. Each subagent works in a worktree under `worktrees/` ([Always work in a worktree](work-in-worktree.md)) and runs `go` with `GOWORK=off` ([Build and test](build-and-test.md)).
- **Origin stays with the parent:** open every branch's draft PR before spawning its implementer ([Issue workflow](issue-workflow.md), "Several issues at once"). Tell each subagent not to push or open PRs, and to commit a stub journal entry for any new branch it cuts and report that branch, so the parent can push it in the next batch.
- **Point, don't paste:** pass paths and line ranges rather than pasting file contents or logs into the prompt.
- **Graph first:** tell it to read `docs/codebase-graph/` before sources, per [Token efficiency](token-efficiency.md).
- **Output contract:** say exactly what to return and how long it may be, for example `file:line` plus a one-line finding each, capped at about 20 lines. No raw file contents, no restating the task.

Loose: "Look into how derive works and tell me what you find."

Tight: "List every call site of `box.Value` constructors outside `box/`. Read the inbound edges to `[[box.value]]` in `docs/codebase-graph/` first, then confirm with `rg`. Return `file:line` and the enclosing function, one per line, nothing else." That is a Haiku task.

Related: [Always work in a worktree](work-in-worktree.md), [Token efficiency](token-efficiency.md), [Trace data flow for race/guard audits](trace-data-flow-for-races.md), [`evalDerive` resolution](evalderive-resolution.md), [Journal convention](journal-convention.md)
