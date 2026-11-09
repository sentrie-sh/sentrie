---
id: api.http
type: Class
language: Go
file_path: api/http.go
tags: http-server, routing, lifecycle, listeners, problem-details
---

# Node: api.HTTPAPI (Server Lifecycle and Routing)

## 1. Architectural Role & Intent
Owns the HTTP surface: route registration, multi-address listener setup, server start and stop, the health endpoints, and the shared RFC 9457 error responder. It supports binding several addresses at once - one `http.Server` per listener, all sharing a single mux - so one process can serve both a loopback and a network interface.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `api.http` | `CALLS` | [[api.net]] | `resolveBindings` turns the port and listen names into concrete addresses. |
| `api.http` | `CALLS` | [[api.middleware]] | Wraps the decision handler with request-ID assignment. |
| `api.http` | `CALLS` | [[api.handle_decision]] | The registered POST handler. |
| `api.http` | `CALLS` | [[api.health]] | Registers the `/health` and `/health/deep` handlers. |
| `api.http` | `DEPENDS_ON` | [[api.problem_details]] | `writeErrorResponse` builds a `ProblemDetails`. |
| `api.http` | `MUTATES` | [[infra.network_sockets]] | Binds and closes TCP listeners. |
| [[cmd]] | `CALLS` | `api.http` | `serve` drives `Setup`, `StartServer`, and `StopServer`. |

## 3. Interface Contracts & Public Surface

- **Signature:** `NewHTTPAPI(executor runtime.Executor, opts ...HTTPAPIOption) -> *HTTPAPI` / `WithRequestTimeout(d time.Duration) -> HTTPAPIOption` / `WithMaxRequestTimeout(d time.Duration) -> HTTPAPIOption`
  - **Behavior:** `requestTimeout` is the default deadline, `runtime.DefaultEvalTimeout` (10s) unless set; `serve` sets it from `--eval-timeout` ([#142](https://github.com/sentrie-sh/sentrie/issues/142)). `maxRequestTimeout` is the ceiling a request may ask for with `Sentrie-Timeout`; `serve` sets it from `--eval-max-timeout` ([#243](https://github.com/sentrie-sh/sentrie/issues/243)). Unset (<= 0), the ceiling is the default. A default that is unbounded or above the ceiling is capped to it. [[api.handle_decision]] picks each request's deadline from these (`requestDeadline` in `api/request_timeout.go`) and hands it to the executor with `runtime.ContextWithEvalTimeout`; the API adds no deadline of its own.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `Setup(ctx context.Context, port int, listen []string) -> error`
  - **Behavior:** Registers `POST /decision/{target...}` (Go 1.22+ pattern routing with a wildcard tail) `GET /health` and `GET /health/deep` ([[api.health]]), resolves bindings, and opens a listener per address with a 30-second read timeout and a write timeout of `max(30s, ceiling + 5s)` (saturating), where the ceiling is `maxRequestTimeout` or else `requestTimeout`, so a request that asks for the longest allowed deadline can still be answered. With no ceiling and no default the write timeout stays at 30 seconds, which then caps a response. On any bind failure it closes every already-opened listener and clears the slice, so a partial bind never leaves sockets dangling.
  - **Side Effects:** Binds sockets; populates `api.listeners`.
  - **Exceptions:** `failed to listen on %s: %w`; binding-resolution errors.

- **Signature:** `StartServer(ctx context.Context, port int, listen []string)`
  - **Behavior:** Spawns a goroutine per listener calling `server.Serve`. Non-`ErrServerClosed` errors are sent to a buffered channel.
  - **Side Effects:** Serves traffic.
  - **Exceptions:** None returned.

- **Signature:** `StopServer(ctx context.Context) -> error`
  - **Behavior:** Gracefully shuts every pair down concurrently with `ListenerServerPair.Shutdown(ctx)` (`http.Server.Shutdown`, then a listener close in case `Serve` never ran) and waits for all of them. It only reads `api.listeners` (it does not clear it), because `StartServer` reads the slice from another goroutine; a second call is harmless since `Shutdown` is idempotent and an already closed listener is ignored. In-flight requests finish until `ctx` is done ([#117](https://github.com/sentrie-sh/sentrie/issues/117)). A pair whose shutdown fails (normally `ctx` ending with requests still running) is logged at WARN, then closed with `http.Server.Close` so the process can exit. `serve` passes a context derived with `context.WithoutCancel` and a timeout of `DefaultShutdownGracePeriod` (10s, a constant, not a flag).
  - **Side Effects:** Stops accepting connections; waits for in-flight requests; drops any still running at the deadline.
  - **Exceptions:** Joined errors, one per pair that did not drain in time (wrapping `context.DeadlineExceeded`), so an unclean shutdown makes `serve` exit 1. Nil on a clean drain.

- **Signature:** `writeErrorResponse(w, r, statusCode int, title, detail string)`
  - **Behavior:** Emits `application/problem+json` with the request ID as `instance` and a timestamp extension.
  - **Side Effects:** Writes the response.
  - **Exceptions:** Encoding failures are logged at debug only.

## 4. Operational Context & Gotchas
- **Statefulness:** Holds the listener slice; not safe for concurrent `Setup`/`Stop`. One executor is shared by all handlers.
- **Performance/Scale Notes:** The read timeout is fixed at 30 seconds. Evaluation is bounded by the request's deadline, and the write timeout is derived from the longest one allowed, so a slow policy fails with a 503 rather than a half-written response. Every listener's server comes from `newServer`, which also sets `ReadHeaderTimeout` (5s), `IdleTimeout` (60s) and `MaxHeaderBytes` (64 KiB) as named constants ([#247](https://github.com/sentrie-sh/sentrie/issues/247)), so a client sending headers slowly or idling on a keep-alive connection is disconnected well before the 30-second read timeout. None of them is a flag.
- **Dependencies Risk:**
  - **`StartServer` cannot report a serve error.** Errors go into `errChan`, which nothing ever reads, so a listener that fails after startup dies silently. `serve` in [[cmd]] launches this in a goroutine and then blocks on `ctx.Done()`, so the process stays alive with no listeners. The deferred `wg.Wait()` does block, so `StartServer` returns only after every `Serve` has returned, which in `serve` means after `StopServer`'s `Shutdown`.
  - **`ListenerServerPair.Close` returns early on the first error**, so a failed listener close leaves the server unclosed. It is now only the fallback after the grace period and the cleanup for a partial `Setup`.
  - **`BaseContext` returns `context.WithoutCancel(setupCtx)`**: cancelling the serve context (the shutdown signal) must not cancel in-flight requests, or the drain would be pointless. Request contexts are still cancelled when their connection closes, which is how the `Close` fallback drops them, and the evaluation deadline bounds how long a request can run (with `--eval-timeout 0` and no ceiling there is no such bound, and only the grace-period `Close` ends the request).
  - **`Setup`'s `port` and `listen` are re-passed to `StartServer`**, which ignores them and uses the stored listeners. The parameters are vestigial and misleading.
  - **The `{target...}` wildcard swallows the entire path tail**, so path parsing and validation are entirely delegated to `ResolveSegments`.
