---
id: api.middleware
type: System / Package
language: Go
file_path: api/middleware/
tags: http-middleware, request-tracing, correlation-id
---

# Node: api/middleware (Request Middleware)

## 1. Architectural Role & Intent
The HTTP middleware chain, currently consisting of request-ID assignment and an empty pass-through placeholder. The request ID is propagated through the request context and surfaces as the `instance` field of every Problem Details error response, giving operators a correlation handle between a client-visible error and server logs.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `api.middleware` | `DEPENDS_ON` | `ext.google_uuid` | Generates the request identifier. |
| [[api.http]] | `CALLS` | `api.middleware` | Wraps the decision handler with `RequestIDMiddleware`. |
| [[api.problem_details]] | `READS_FROM` | `api.middleware` | The request ID becomes the `instance` field. |

## 3. Interface Contracts & Public Surface

- **Signature:** `RequestIDMiddleware(next http.Handler) -> http.Handler`
  - **Behavior:** Ensures a request ID exists in the request context, then delegates. Idempotent — an ID already present is preserved, so nesting is safe.
  - **Side Effects:** Replaces the request with one carrying an augmented context.
  - **Exceptions:** None.

- **Signature:** `GetRequestIDFromRequest(req *http.Request) -> string`
  - **Behavior:** Reads the ID from the context with an **unchecked type assertion**.
  - **Side Effects:** None.
  - **Exceptions:** **Panics** when no ID is present — see below.

- **Signature:** `HasRequestIDInRequest(req *http.Request) -> bool`
  - **Behavior:** Reports whether the context carries an ID. The intended guard for the getter.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `Middleware(next http.Handler) -> http.Handler`
  - **Behavior:** Calls `next` and nothing else — an unused placeholder.
  - **Side Effects:** None.
  - **Exceptions:** None.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless. The ID lives for one request's lifetime.
- **Performance/Scale Notes:** One UUID generation and one context allocation per request. Negligible.
- **Dependencies Risk:**
  - **`GetRequestIDFromRequest` panics on a request that never passed through the middleware.** `req.Context().Value(key).(string)` is an unchecked assertion on a nil interface. It is called from `writeErrorResponse`, which is only reachable from the decision handler — and that handler *is* wrapped — so it is safe today. But the coupling is invisible: registering any new route that calls `writeErrorResponse` without also wrapping it turns every error response on that route into a panic. Using the comma-ok form and falling back to an empty string would remove the trap entirely, and `HasRequestIDInRequest` already exists to express the check.
  - **The request ID is generated, never read from the client.** There is no `X-Request-ID` header ingestion, so an ID assigned by an upstream proxy or gateway is discarded and correlation breaks at the Sentrie boundary.
  - **The ID is not returned in a response header**, only in the `instance` field of error responses — so a **successful** request gives the caller no correlation handle at all.
  - **The ID is not attached to the logger.** `slog` calls in the handlers use the context but nothing installs the ID as a log attribute, so correlating a request to its log lines requires the error response to have been produced.
  - **`Middleware` is dead code** whose presence suggests the chain was expected to grow — which it should, given authentication and rate limiting are both absent.
