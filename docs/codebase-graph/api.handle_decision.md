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
| `api.handle_decision` | `CALLS` | [[runtime.executor]] | `ExecPolicy` when no rule is named, `ExecRule` otherwise. |
| `api.handle_decision` | `CALLS` | [[index.segments]] | `ResolveSegments` parses the path tail into namespace, policy, and rule. |
| `api.handle_decision` | `READS_FROM` | [[runtime.trace]] | `ExecutorOutput` is serialised whole, including the trace tree. |
| `api.handle_decision` | `CALLS` | [[api.http]] | `writeErrorResponse` for every handled failure. |
| [[api.http]] | `CALLS` | `api.handle_decision` | Registered as the POST handler. |

## 3. Interface Contracts & Public Surface

- **Signature:** `POST /decision/{namespace}/{policy}[/{rule}]`
  - **Request:** `DecisionRequest` - `{"facts": {"name": <any>, ...}}`. Query parameters are collected into a `runConfig` map.
  - **Response:** `DecisionResponse` - `{"decisions": [ExecutorOutput...]}` on success; RFC 9457 Problem Details on evaluation failure.
  - **Behavior:** Sets CORS headers, validates the path, short-circuits `OPTIONS`, rejects non-POST, resolves segments, parses query parameters, decodes the body, evaluates, and encodes.
  - **Side Effects:** Full policy evaluation - JavaScript execution, memoization cache writes, VM pool acquisition.
  - **Exceptions:** Problem Details for missing path (400), unresolvable path (404), wrong method (405), unparseable JSON (400), and evaluation failures (400 for invalid invocation, 500 otherwise).

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless per request; all state lives in the shared executor.
- **Performance/Scale Notes:** Unbounded work per request. The body is decoded without a size limit, facts flow into evaluation where collection builtins multiply over them, and the trace tree allocates a node per evaluation step. Request cost is a function of both payload size and policy complexity, neither of which is capped.
- **Dependencies Risk:**
  - **`runConfig` is built from query parameters and never used.** Dead code that reads as though per-request configuration is supported.
  - **`strings.TrimPrefix(path, "/decision/")` is a no-op.** `r.PathValue("target")` already excludes the route prefix.
  - **CORS is wildcard and unconditional**, and the `OPTIONS` short-circuit is checked *after* the path validation, so a preflight for an empty path gets a 400 instead of the expected 200.
  - **The response embeds the trace tree, which contains evaluated fact values**, so the endpoint returns the caller's own data plus whatever intermediate values the policy computed. With no authentication in front of it, that is a disclosure surface.
  - **There is no authentication.** Anyone who can reach the port can evaluate any policy with arbitrary facts. Filed as [#115](https://github.com/sentrie-sh/sentrie/issues/115).
