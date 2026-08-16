---
id: api.http
type: Class
language: Go
file_path: api/http.go
tags: http-server, routing, lifecycle, listeners, problem-details
---

# Node: api.HTTPAPI (Server Lifecycle and Routing)

## 1. Architectural Role & Intent
Owns the HTTP surface: route registration, multi-address listener setup, server start and stop, the health endpoint, and the shared RFC 9457 error responder. It supports binding several addresses at once — one `http.Server` per listener, all sharing a single mux — so one process can serve both a loopback and a network interface.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `api.http` | `CALLS` | [[api.net]] | `resolveBindings` turns the port and listen names into concrete addresses. |
| `api.http` | `CALLS` | [[api.middleware]] | Wraps the decision handler with request-ID assignment. |
| `api.http` | `CALLS` | [[api.handle_decision]] | The registered POST handler. |
| `api.http` | `DEPENDS_ON` | [[api.problem_details]] | `writeErrorResponse` builds a `ProblemDetails`. |
| `api.http` | `MUTATES` | [[infra.network_sockets]] | Binds and closes TCP listeners. |
| [[cmd]] | `CALLS` | `api.http` | `serve` drives `Setup`, `StartServer`, and `StopServer`. |

## 3. Interface Contracts & Public Surface

- **Signature:** `Setup(ctx context.Context, port int, listen []string) -> error`
  - **Behavior:** Registers `POST /decision/{target...}` (Go 1.22+ pattern routing with a wildcard tail) and `GET /health`, resolves bindings, and opens a listener per address with a 30-second read and write timeout. On any bind failure it closes every already-opened listener and clears the slice, so a partial bind never leaves sockets dangling.
  - **Side Effects:** Binds sockets; populates `api.listeners`.
  - **Exceptions:** `failed to listen on %s: %w`; binding-resolution errors.

- **Signature:** `StartServer(ctx context.Context, port int, listen []string)`
  - **Behavior:** Spawns a goroutine per listener calling `server.Serve`. Non-`ErrServerClosed` errors are sent to a buffered channel.
  - **Side Effects:** Serves traffic.
  - **Exceptions:** None returned.

- **Signature:** `StopServer(ctx context.Context) -> error`
  - **Behavior:** Closes each listener/server pair and nils the slice. `ListenerServerPair.Close` closes the listener then the server, returning early if the listener close fails.
  - **Side Effects:** Terminates connections.
  - **Exceptions:** Always returns nil — pair errors are discarded.

- **Signature:** `handleHealth(w, r)`
  - **Behavior:** Returns status and current time with HTTP 200, unconditionally.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `writeErrorResponse(w, r, statusCode int, title, detail string)`
  - **Behavior:** Emits `application/problem+json` with the request ID as `instance` and a timestamp extension.
  - **Side Effects:** Writes the response.
  - **Exceptions:** Encoding failures are logged at debug only.

## 4. Operational Context & Gotchas
- **Statefulness:** Holds the listener slice; not safe for concurrent `Setup`/`Stop`. One executor is shared by all handlers.
- **Performance/Scale Notes:** Timeouts are hardcoded at 30 seconds each way and are not configurable, so a policy that legitimately takes longer is cut off mid-evaluation with the response half-written. There is no `IdleTimeout` or `ReadHeaderTimeout`, leaving the server open to slow-header connection exhaustion.
- **Dependencies Risk:**
  - **`StartServer` cannot report a serve error.** Errors go into `errChan`, and the deferred `wg.Wait()` followed by `close(errChan)` runs immediately — the function returns without waiting, and nothing ever reads the channel. A listener that fails after startup dies silently. `serve` in [[cmd]] launches this in a goroutine and then blocks on `ctx.Done()`, so the process stays alive with no listeners.
  - **`StopServer` is not a graceful shutdown.** It calls `Close`, which terminates active connections immediately, rather than `http.Server.Shutdown`, which drains them. An in-flight decision is dropped at exit.
  - **`ListenerServerPair.Close` returns early on the first error**, so a failed listener close leaves the server unclosed.
  - **`BaseContext` returns the setup context**, so cancelling it cancels every in-flight request — correct for shutdown, but it means request contexts are tied to process lifetime rather than to the request.
  - **`Setup`'s `port` and `listen` are re-passed to `StartServer`**, which ignores them and uses the stored listeners. The parameters are vestigial and misleading.
  - **The `{target...}` wildcard swallows the entire path tail**, so path parsing and validation are entirely delegated to `ResolveSegments`.
