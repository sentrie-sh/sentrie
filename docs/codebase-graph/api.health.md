---
id: api.health
type: Function / Endpoint
language: Go
file_path: api/health.go
tags: health, liveness, readiness, deepcheck, probe
---

# Node: api.health (GET /health and GET /health/deep)

## 1. Architectural Role & Intent
The two probe endpoints ([#40](https://github.com/sentrie-sh/sentrie/issues/40)). `/health` is a liveness probe that answers for any process able to serve a request. `/health/deep` is a readiness probe: it inspects the executor, its index, the JavaScript registry and the pack directory, and reports a status per component. It reports only what the process can observe; the issue's executor cache reservations, JS instance pools, pack location and "operational" probes are left out because nothing in the runtime exposes them, and a made-up value would defeat a health check.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `api.health` | `CALLS` | [[runtime.executor]] | `Index()` and `BuiltinModules()` are read to build the deep report. |
| `api.health` | `CALLS` | [[index.status]] | `Index.Status()` supplies validated/committed state, errors and statistics without triggering validation. |
| `api.health` | `READS_FROM` | [[infra.filesystem]] | `os.Stat` on the pack directory, to report whether it is still accessible. |
| [[api.http]] | `CALLS` | `api.health` | Registers `GET /health` and `GET /health/deep`. |

## 3. Interface Contracts & Public Surface

- **Signature:** `GET /health`
  - **Behavior:** HTTP 200 with `{"check_id": <random nonce>, "status": "healthy", "time": <RFC3339 UTC>}`. `check_id` is a fresh `crypto/rand.Text()` value per response. Unconditional, so it works with no executor at all. `status` and `time` are unchanged from before #40, so existing callers keep working.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `GET /health/deep`
  - **Behavior:** `{"status", "time", "components", "errors"}`. Components: `index` (`validated`, `committed`, `validation_error` and `commit_error` (fixed strings or null), `statistics` of namespaces, policies, programs, rules, shapes), `executor` (`initialized`), `js_runtime` (`registry_initialized`, `builtin_modules` with the sorted registered module keys), `pack` (`loaded`, `name`, `version`, `accessible`, `error`). `status` is `healthy` or `unhealthy`; there is no `degraded`. Any unhealthy component makes the whole response `unhealthy` with HTTP 503 (200 otherwise), and adds `{"component", "message"}` to `errors`, which is always an array, never null (for the index, the message is one of the fixed strings above, not the underlying error). `validated` is false when validation failed, as in the issue's unhealthy example. A commit failure inside `Validate` is also recorded as the validation error, and is reported once, as `commit_error`.
  - **Side Effects:** None. It never validates or commits the index.
  - **Exceptions:** None.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless; every call reads live state.
- **Dependencies Risk:**
  - **The deep response is unauthenticated** like every other route ([#115](https://github.com/sentrie-sh/sentrie/issues/115)). It exposes module names, definition counts, the pack name and version. It never includes the pack's filesystem path: index failures are reported as the fixed strings `index validation failed` and `index commit failed` (the index's own error text embeds source spans with the absolute pack path, so the full error goes to the API logger at Debug), and the pack-directory error says only that the directory is not accessible. Error detail stays withheld until [#115](https://github.com/sentrie-sh/sentrie/issues/115) adds authentication.
  - **A closed executor still reports healthy.** `Executor` exposes no closed flag, so shutdown after `Close` is not visible here.
  - **`/health` is liveness only.** Use `/health/deep` as the readiness probe.
  - Both responses carry `Cache-Control: no-store`.
