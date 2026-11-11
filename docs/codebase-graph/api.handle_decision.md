---
id: api.handle_decision
type: Function / Endpoint
language: Go
file_path: api/handle_decision.go
tags: endpoint, decision, json, defect, cors, security
---

# Node: api.handleDecision (POST /decision)

## 1. Architectural Role & Intent
The only functional endpoint in the service and the whole reason the API exists: it accepts a policy or rule path plus a fact payload, invokes the executor, and returns the resulting decisions. Requesting a policy without a rule evaluates every exported rule; naming a rule evaluates just that one.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `api.handle_decision` | `CALLS` | [[runtime.executor]] | `ExecPolicy` when no rule is named (a `RuleErrors` error means 207 with the outputs), `ExecRule` otherwise, under `ContextWithEvalTimeout` carrying the request's deadline; `ErrorStatus` for problem statuses. |
| `api.handle_decision` | `CALLS` | [[index.segments]] | `ResolveSegments` parses the path tail into namespace, policy, and rule. |
| `api.handle_decision` | `READS_FROM` | [[runtime.trace]] | `ExecutorOutput` is serialised whole, including the trace tree. |
| `api.handle_decision` | `CALLS` | [[api.http]] | `writeErrorResponse` for every handled failure. |
| [[api.http]] | `CALLS` | `api.handle_decision` | Registered as the POST handler. |

## 3. Interface Contracts & Public Surface

- **Signature:** `POST /decision/{namespace}/{policy}[/{rule}]`
  - **Request:** `DecisionRequest` - `{"facts": {"name": <any>, ...}}` keyed by each fact's declared name, not its alias ([#147](https://github.com/sentrie-sh/sentrie/issues/147)), decoded with `json.Decoder.UseNumber` so an integral number that fits int64 stays exact ([#166](https://github.com/sentrie-sh/sentrie/issues/166)); a number float64 cannot hold (`1e400`) is rejected with 400 `Invalid JSON` via `box.CheckJSONNumbers`. The response is encoded into a buffer before the status is written, so a value JSON cannot represent (a non-finite result) is a 500 problem response rather than a 200 with an empty body. Query parameters are collected into a `runConfig` map. Optional header `Sentrie-Timeout`: a Go duration (`250ms`, `2s`, `1m30s`) choosing this request's evaluation deadline ([#243](https://github.com/sentrie-sh/sentrie/issues/243)).
  - **Response:** `DecisionResponse` - `{"decisions": [ExecutorOutput...]}`, 200 when every rule succeeded. For a policy request, a rule that failed does not fail the request ([#143](https://github.com/sentrie-sh/sentrie/issues/143)): the response is **207 Multi-Status** (even when every rule failed) with every rule's output, and a failed one carries `"error": {"title": "Evaluation Failed", "status": <400|503|500>, "detail": "..."}` next to an unknown decision (`runtime.RuleError`, [[runtime]]). A named-rule request that fails, a policy that does not resolve, and a fact error answer with RFC 9457 Problem Details as before. `Sentrie-Timeout-Applied` carries the deadline used (`time.Duration.String()`, the default or the requested value) whenever one applied, including on a 503.
  - **Behavior:** Evaluates under `r.Context()` (cancelled when the client goes away). The deadline is chosen per request by `requestDeadline`: the default (`--eval-timeout`, capped to `--eval-max-timeout`), or the `Sentrie-Timeout` value if it is within the ceiling. It is handed to the executor with `runtime.ContextWithEvalTimeout`, replacing the executor's own, so there is one deadline and a timeout error names it. Sets CORS headers (allowing `Sentrie-Timeout` and exposing `Sentrie-Timeout-Applied`), validates the path, short-circuits `OPTIONS`, rejects non-POST, checks `Sentrie-Timeout`, resolves segments, parses query parameters, decodes the body, evaluates, and encodes.
  - **Side Effects:** Full policy evaluation - JavaScript execution, memoization cache writes, fresh JS VMs per module call.
  - **Exceptions:** Problem Details for an invalid `Sentrie-Timeout` (400: unparseable, zero or negative, longer than 32 bytes, sent more than once, or above the ceiling, never clamped and never a fallback to the default), missing path (400), unresolvable path (404), wrong method (405), unparseable JSON (400), and a named rule's evaluation failure (status from `runtime.ErrorStatus`: 400 for invalid invocation, 503 when the deadline passed, 500 otherwise). A fact error (missing required fact, a supplied fact failing its type, or a fact an imported policy is given with `with` failing its type) is a request error, a 400 problem document for a policy request too ([#147](https://github.com/sentrie-sh/sentrie/issues/147)); a default failing its type is a 500; a deadline or other evaluation failure is per rule, inside the 207 body, with the same statuses.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless per request; all state lives in the shared executor.
- **Performance/Scale Notes:** Time per request is bounded by its deadline, and JS call depth by `js.MaxCallStackSize`; memory is not (see [[runtime.executor]]). The body is decoded without a size limit, facts flow into evaluation where collection builtins multiply over them, and the trace tree allocates a node per evaluation step. Request cost is a function of both payload size and policy complexity, neither of which is capped.
- **Dependencies Risk:**
  - **`runConfig` is built from query parameters and never used.** Dead code that reads as though per-request configuration is supported.
  - **`strings.TrimPrefix(path, "/decision/")` is a no-op.** `r.PathValue("target")` already excludes the route prefix.
  - **CORS is wildcard and unconditional**, and the `OPTIONS` short-circuit is checked *after* the path validation, so a preflight for an empty path gets a 400 instead of the expected 200.
  - **The response embeds the trace tree, which contains evaluated fact values**, so the endpoint returns the caller's own data plus whatever intermediate values the policy computed. With no authentication in front of it, that is a disclosure surface.
  - **There is no authentication.** Anyone who can reach the port can evaluate any policy with arbitrary facts, and with `Sentrie-Timeout` can ask for evaluations up to the ceiling. Filed as [#115](https://github.com/sentrie-sh/sentrie/issues/115).
  - **`Sentrie-Timeout` is untrusted input.** It is length-checked before `time.ParseDuration` (which rejects overflow), only chooses the deadline, is quoted in a 400 only once known to be short, and is never logged.
