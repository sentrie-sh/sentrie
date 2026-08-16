---
id: api
type: System / Package
language: Go
file_path: api/
tags: http, server, decision-endpoint, entrypoint, security
---

# Node: api (HTTP Service Layer)

## 1. Architectural Role & Intent
The network-facing façade over the policy engine. It translates HTTP requests into executor calls and executor results into JSON, and it is the boundary at which Sentrie stops being a library and becomes a service. It holds no policy logic of its own - every decision comes from [[runtime.executor]] - which keeps the layer thin but also means every operational concern (authentication, limits, error mapping) has to be added here deliberately.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `api` | `LAYERED_ON` | [[runtime]] | Holds a `runtime.Executor` and calls `ExecPolicy` / `ExecRule`. |
| `api` | `LAYERED_ON` | [[index]] | `executor.Index().ResolveSegments` turns a URL path into namespace, policy, and rule. |
| `api` | `LAYERED_ON` | [[api.middleware]] | Request ID assignment. |
| `api` | `READS_FROM` | [[runtime.trace]] | `ExecutorOutput` carries the trace tree, which is serialised into responses. |
| [[cmd]] | `CALLS` | `api` | `serve` constructs the HTTP API and starts it. |

## 3. Interface Contracts & Public Surface

- **Signature:** `NewAPI(executor runtime.Executor) -> *API`
  - **Behavior:** A minimal wrapper holding an executor. Currently has **no methods** - a transport-agnostic entry point that nothing uses; the HTTP path goes through `HTTPAPI` instead.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `NewHTTPAPI(executor runtime.Executor) -> *HTTPAPI` and its lifecycle - see [[api.http]]
  - **Behavior:** Routing, listener management, health, and error responses.
  - **Side Effects:** Binds sockets.
  - **Exceptions:** Bind failures.

- **Signature:** `POST /decision/{target...}` - see [[api.handle_decision]]
  - **Behavior:** The only functional endpoint. Body is `{"facts": {...}}`; response is `{"decisions": [...], "error": "..."}`.
  - **Side Effects:** Full policy evaluation.
  - **Exceptions:** RFC 9457 Problem Details on the handled error paths.

- **Signature:** `GET /health`
  - **Behavior:** Returns `{"status":"healthy","time":"<RFC3339>"}` with HTTP 200. Unconditional - it does **not** check the executor, the index, or the VM pool, so it reports healthy for a process that cannot serve a single decision.
  - **Side Effects:** None.
  - **Exceptions:** None.

## 4. Operational Context & Gotchas
- **Statefulness:** `HTTPAPI` owns its listeners and a shared executor. The executor is concurrency-safe by design, so one instance serves all requests and all connections share the JavaScript VM pool and the memoization cache.
- **Performance/Scale Notes:** Read and write timeouts are fixed at 30 seconds with no configuration. There is **no concurrency limit**, so inbound request concurrency maps directly onto executor concurrency; the VM pool (max 10) becomes the implicit bottleneck under load, and requests queue on it invisibly.
- **Dependencies Risk:**
  - **The success path panics.** `handleDecision` calls `runErr.Error()` on a nil error, so a successful evaluation never produces a response. Filed as [#114](https://github.com/sentrie-sh/sentrie/issues/114) - see [[api.handle_decision]].
  - **No authentication, no authorization, no rate limiting, wildcard CORS, and no request body limit.** Filed as [#115](https://github.com/sentrie-sh/sentrie/issues/115). The default bind of `local` is the one thing keeping the out-of-the-box posture safe.
  - **`local6` and `network6` cannot bind** because the addresses are double-bracketed - see [[api.net]].
  - **Responses embed the full trace tree**, which contains evaluated fact values. Any caller that can reach the endpoint can read back the data the policy saw.
  - **`API` and `HTTPAPI` are parallel, unrelated types.** The former looks like the intended abstraction and is entirely unused; a reader looking for the service entry point should go to `HTTPAPI`.
  - **Health is a liveness probe only.** Using it as a readiness probe will route traffic to an instance whose pack failed to load - though in practice `serve` exits before starting the server if loading fails, so the window is narrow.
