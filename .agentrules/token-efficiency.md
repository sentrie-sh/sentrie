# Token efficiency

Tool output stays in context and is re-billed on every later turn. Keep it small.

## Read the graph before the source

`docs/codebase-graph/` describes this architecture as one cross-linked Markdown node
per package or significant file, written to be read instead of grep-scanning sources.
Exhaust it first.

- **Orienting:** `sed -n '/^### Reading Order/,/^## Cross-Cutting/p' docs/codebase-graph/overview.md`
  gives the reading order and node catalog without the schema preamble.
- **One package:** node ids are dotted and map to the path, so `runtime/js` is
  `docs/codebase-graph/runtime.js.md`. `ls docs/codebase-graph/` is the index.
- **What a package depends on:** its section 2 edge table alone, not the whole node:
  `sed -n '/^## 2\./,/^## 3\./p' docs/codebase-graph/runtime.md`.
- **What breaks if I change X:** inbound edges, which no `go doc` gives you.
  `rg -l '\[\[box.value\]\]' docs/codebase-graph/`, then read those nodes' section 2.
- **Before changing behaviour:** section 4 (Operational Context & Gotchas) of that node.
- `make docs-graph` checks the schema, not the prose. Confirm against source any
  specific claim you are about to depend on, and update the node in the same change
  ([Codebase knowledge graph](codebase-knowledge-graph.md)).

## Output discipline

- Never `cat` a file to find something. Locate first with `rg -n 'pattern' --type go`,
  then read only the hit with `sed -n '120,180p' path/to/file.go`.
- Prefix every `go` command with `GOWORK=off` ([Build and test](build-and-test.md)).
- For a Go package's API, `GOWORK=off go doc ./box` or `GOWORK=off go doc ./index Index`
  beats opening sources.
- Test scope ladder, narrowest first, stopping as soon as the change is proven:
  1. `GOWORK=off go test ./parser/ -run 'TestParserTestSuite/TestParseNotExpressionForms' -count=1 -failfast`
  2. `GOWORK=off go test ./parser/... ./index/... -count=1 -failfast`
  3. `GOWORK=off go test ./... -count=1` only before pushing. `./runtime/...` is the slowest
     package tree, so leave it out of the inner loop unless the change is in it.
  Never pass `-v`: without it the run prints failures only.
- One language fixture: `GOWORK=off go test ./lang_test/ -run 'TestFixturesParseIndexAndValidate/0009' -count=1`.
- `git diff --stat` first. Full diffs only for named paths: `git diff -- runtime/js/`.
- Anything that can exceed ~100 lines (`go test ./...`, `golangci-lint run`, `make gen`)
  goes to a file outside the repo, read selectively on failure:
  `GOWORK=off go test ./... > /tmp/sentrie-test.log 2>&1 || rg -n '^(FAIL|---|panic)' /tmp/sentrie-test.log`.
- Never paste a log body back. State the failing command, the cause, and the log path.

## Session discipline

- One task per session. Write the handoff into `journal/<branchname>.md` per
  [Journal convention](journal-convention.md) and start the next task fresh from it.
- After two failed attempts at the same fix, stop and report the diagnosis. Every
  retry carries all prior failure output forward.

## Cache discipline

- Do not edit `AGENTS.md` or any `.agentrules/*.md` mid-task unless that is the task.
  Batch rule changes into their own session and PR.
- Any listing an agent reads must be deterministically ordered: pipe `find`/`ls`
  through `sort`, or use `git ls-files` (already sorted). Generated output (`make gen`
  into `ast/typeref_constraint_args_gen.go`, the catalog in `docs/codebase-graph/overview.md`)
  must stay stably ordered so regeneration and review do not churn.

## Delegation

- Delegate bulk reading and repo-wide sweeps to a subagent per
  [Parallelize with subagents](parallelize-with-subagents.md). Tell it to work the
  graph first and open sources only for what the graph cannot answer.
- Require `file:line` plus a one-line finding each, capped at about 20 lines. Raw
  file contents must never reach the main context.

Related: [Build and test](build-and-test.md), [Codebase knowledge graph](codebase-knowledge-graph.md)

Relevant keywords: token efficiency, context window, context budget, tool output, cost,
codebase graph, docs/codebase-graph, inbound edges, narrow tests, quiet flags, log
redirection, prompt cache, handoff, retry limit, delegation, subagent summary.
