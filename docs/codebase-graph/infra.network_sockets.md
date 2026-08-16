---
id: infra.network_sockets
type: Infrastructure
language: N/A
file_path: (external)
tags: infrastructure, boundary, networking, serving, security
---

# Node: Network Sockets (TCP Listeners)

## 1. Architectural Role & Intent
The TCP listeners that make Sentrie a service rather than a command. This is the system's only inbound network boundary: every request that reaches [[api.handle_decision]] arrives through a socket opened here, and the facts in those request bodies are the least trusted input the engine handles.

Modelled as its own node because listener lifecycle is where several independent defects converge - address resolution produces unbindable strings for IPv6, bind failures after startup are unreportable, and shutdown severs connections rather than draining them. None of those are visible from any single consuming node.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| [[api.http]] | `MUTATES` | `infra.network_sockets` | Binds and closes TCP listeners. |

## 3. Interface Contracts & Public Surface

- **Signature:** Listener set
  - **Behavior:** One listener per resolved binding. Bindings come from [[api.net]], which maps `--http-listen` values and `--http-port` onto `host:port` strings.
  - **Side Effects:** Occupies host ports for the process lifetime.
  - **Exceptions:** A bind failure at startup is returned; a failure afterwards is not - see below.

- **Signature:** Shutdown
  - **Behavior:** `StopServer` calls `Close` on the server.
  - **Side Effects:** Terminates active connections immediately.
  - **Exceptions:** None surfaced.

## 4. Operational Context & Gotchas
- **Statefulness:** Long-lived OS resources held for the process lifetime. Nothing reopens a listener that dies.
- **Performance/Scale Notes:** No connection limit, no read timeout, and no request body size limit are configured, so concurrency is bounded only by the host.
- **Dependencies Risk:**
  - **IPv6 predefined names cannot bind at all.** `local6` and `network6` resolve to already-bracketed hosts that `net.JoinHostPort` brackets again, producing `[[::1]]:7529` and `[[::]]:7529`, which `net.Listen` rejects. Filed as [#116](https://github.com/sentrie-sh/sentrie/issues/116).
  - **A listener that fails after startup dies silently.** `StartServer` writes serve errors to a channel nothing reads, and [[cmd]]'s `serve` then blocks on `ctx.Done()` - leaving a live process with no listeners and no error. See [[api.http]].
  - **Shutdown is abrupt, not graceful.** `Close` rather than `http.Server.Shutdown` means an in-flight decision is dropped at exit. Compounded by [#117](https://github.com/sentrie-sh/sentrie/issues/117): `SIGTERM` is not registered at all, so orchestrated shutdown never reaches this path.
  - **The boundary is unauthenticated.** No authentication, wildcard CORS, and no body limit - filed as [#115](https://github.com/sentrie-sh/sentrie/issues/115). Anything that can reach the socket can request a decision.
