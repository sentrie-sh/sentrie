---
name: Issue 62 flag rename plan
overview: "Implement issue #62 by renaming `serve` CLI flags from `--port`/`--listen` to `--http-port`/`--http-listen`, with explicit breaking-change behavior, docs/test updates, and required PR descriptions."
todos:
  - id: rename-serve-cli-flags
    content: Rename serve flag/input names and hydration tags in sentrie cmd/serve.go to http-port/http-listen.
    status: completed
  - id: sync-docs-sentrie
    content: Update any sentrie docs references (README and related text) to new flag names where applicable.
    status: completed
  - id: sync-docs-website
    content: Update all website CLI reference/getting-started/serving docs to use --http-port and --http-listen consistently.
    status: completed
  - id: validate-tests-help-search
    content: Run explicit help/legacy-flag behavior checks, targeted tests, CI/script scans, and final searches for stale old-flag references.
    status: completed
  - id: update-pr-descriptions
    content: Update required PR description markdown files in sentrie and website to accurately describe branch-only changes.
    status: completed
isProject: false
---

# Implement Issue #62 Flag Rename

## Goal
Deliver the breaking CLI rename requested in issue [#62](https://github.com/sentrie-sh/sentrie/issues/62):
- `--port` -> `--http-port`
- `--listen` -> `--http-listen`

while keeping server behavior unchanged and ensuring docs/tests stay aligned.

## Scope and Repositories

### Repo: `sentrie`
- Core CLI flag definition + hydration:
  - [`cmd/serve.go`](cmd/serve.go)
- Command registration context (verification only):
  - [`cmd/cmd.go`](cmd/cmd.go)
- Runtime HTTP/network internals (no-change expected; validate only if tests/compile indicate impact):
  - [`api/http.go`](api/http.go)
  - [`api/net.go`](api/net.go)
- Existing behavior tests (secondary smoke coverage if wiring remains unchanged):
  - [`api/helpers_test.go`](api/helpers_test.go)
  - [`api/listener_test.go`](api/listener_test.go)
- Primary CLI validation/test targets:
  - [`cmd/`](cmd/)
- User-facing repo docs to update if they reference old flags:
  - [`README.md`](README.md)
- Branch PR description required by repo rules:
  - [`PR_DESCRIPTION_62-breaking-change-rename-port-and-listen-to-http-port-and-http-listen.md`](PR_DESCRIPTION_62-breaking-change-rename-port-and-listen-to-http-port-and-http-listen.md)

### Repo: `website`
- Main CLI docs hub that currently duplicates serve options/examples:
  - [`src/content/docs/cli-reference/index.md`](src/content/docs/cli-reference/index.md)
- Dedicated serve CLI reference page:
  - [`src/content/docs/cli-reference/serve.md`](src/content/docs/cli-reference/serve.md)
- Additional docs pages containing old `--port`/`--listen` wording/examples:
  - [`src/content/docs/getting-started.md`](src/content/docs/getting-started.md)
  - [`src/content/docs/running-sentrie/serving-policies.md`](src/content/docs/running-sentrie/serving-policies.md)
- Website PR description required by repo rules:
  - [`PR_DESCRIPTION.md`](PR_DESCRIPTION.md)

## Implementation Plan

1. Update serve command flag names in `sentrie`.
- In [`cmd/serve.go`](cmd/serve.go), rename `cling` input names from `"port"`/`"listen"` to `"http-port"`/`"http-listen"`.
- Update hydration tags in `serveCmdArgs` to `cling-name:"http-port"` and `cling-name:"http-listen"`.
- Refresh flag descriptions to explicitly say HTTP (for example: "HTTP port to listen on", "HTTP address(es) to listen on").
- Keep internal runtime argument fields (`Port`, `Listen`) and API wiring intact unless compile checks force renames.

2. Lock breaking-change behavior for legacy flags.
- Intentionally treat `--port` and `--listen` as removed flags (no compatibility alias/shim unless requirements change).
- Verify and document expected failure behavior when legacy flags are used (unknown flag path and CLI error UX).

3. Confirm no hidden config/env aliasing needs migration.
- Run explicit searches in `sentrie` for env/config bindings and naming patterns (for example: `cling-env`, `SENTRIE_`, `port`, `listen`) to ensure no legacy mappings remain.
- If any bindings are discovered, rename to `http_`-prefixed equivalents and document as part of the breaking change.

4. Update all `website` docs references from old flags to new flags.
- Replace option tables and section headings (`--port`, `--listen`) with `--http-port`, `--http-listen`.
- Update every serve example command in listed docs files to new flags.
- Ensure explanatory text and troubleshooting notes use the new names consistently.
- Keep semantics/default values unchanged unless docs currently contradict implementation.

5. Align `sentrie` repo docs.
- Review [`README.md`](README.md) for serve option examples; update any explicit old flag mentions.

6. Validate behavior, tests, and automation touchpoints.
- Add or update CLI-focused tests in `sentrie/cmd` to assert `serve --help` contains `--http-port` and `--http-listen` and does not list the legacy flags.
- Add a negative test/assertion path for legacy flag rejection (`--port`, `--listen`) with expected error behavior.
- Run targeted tests in `sentrie` (prioritizing `cmd`; run API tests as smoke coverage if needed) to confirm no regressions.
- Search CI/workflows/scripts in both repos for stale `--port`/`--listen` invocations and update if found.
- Perform final repo-wide searches to ensure no unintended lingering old flag references in user-facing docs/help.

7. Delivery hygiene: update required PR description files.
- In `sentrie`, update [`PR_DESCRIPTION_62-breaking-change-rename-port-and-listen-to-http-port-and-http-listen.md`](PR_DESCRIPTION_62-breaking-change-rename-port-and-listen-to-http-port-and-http-listen.md) so it reflects only branch-introduced changes and includes review/testing/dependency notes per repo conventions.
- In `website`, update [`PR_DESCRIPTION.md`](PR_DESCRIPTION.md) to match doc changes and required title format.

## Verification Checklist
- `sentrie serve --help` shows `--http-port` and `--http-listen`.
- `sentrie serve --port ...` and `sentrie serve --listen ...` are rejected with expected unknown-flag behavior.
- Old flags no longer appear in `sentrie` source/help/docs.
- `website` docs contain no stale `--port`/`--listen` examples for serve.
- CI/workflow/script references in both repos are checked for stale old flags and updated if present.
- Relevant tests pass and no new lints introduced in touched files.
- Both required PR description files are updated and consistent with actual diffs.

## Risks and Mitigations
- Risk: missed doc occurrences due duplicated CLI examples.
  - Mitigation: run exhaustive text search in both repos before finalizing.
- Risk: accidental semantic drift while renaming (e.g., changing runtime bind logic).
  - Mitigation: keep API/net internals unchanged; limit changes to flag naming and docs.
- Risk: branch PR description drifting from actual diff.
  - Mitigation: regenerate contents from `base..HEAD` diff before final handoff.