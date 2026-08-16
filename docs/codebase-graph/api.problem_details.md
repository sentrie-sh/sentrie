---
id: api.problem_details
type: Class
language: Go
file_path: api/problem_details.go
tags: error-format, rfc9457, serialization, http
---

# Node: api.ProblemDetails (RFC 9457 Error Format)

## 1. Architectural Role & Intent
The standard error representation for every handled HTTP failure, implementing RFC 9457 Problem Details. Using a specified format rather than an ad-hoc error shape means clients can parse failures generically, and the extension map allows Sentrie-specific context without breaking that contract.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `api.problem_details` | `DEPENDS_ON` | `ext.encoding_json` | Implements a custom `MarshalJSON` to flatten extensions. |
| [[api.http]] | `CALLS` | `api.problem_details` | `writeErrorResponse` constructs and encodes one per failure. |
| [[api.middleware]] | `READS_FROM` | `api.problem_details` | The request ID populates `instance`. |

## 3. Interface Contracts & Public Surface

- **Signature:** `ProblemDetails` — `{Type, Title, Status, Detail, Instance string/int, Ext map[string]any}`
  - **Behavior:** The RFC's five standard members plus an extension map. `Ext` is tagged `json:"-"` because it is flattened manually rather than nested.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `NewProblemDetails(type_, title, detail, instance string, status int, ext map[string]any) -> *ProblemDetails`
  - **Behavior:** Plain constructor. **Unused** — `writeErrorResponse` builds the struct literally.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `(*ProblemDetails) MarshalJSON() -> ([]byte, error)`
  - **Behavior:** Builds a map, adds each standard field **only when non-zero**, then copies every extension key in at the top level, matching the RFC's flat extension model.
  - **Side Effects:** None.
  - **Exceptions:** Propagated from `json.Marshal`.

## 4. Operational Context & Gotchas
- **Statefulness:** Value type; no shared state.
- **Performance/Scale Notes:** One map allocation per error response. Irrelevant at any realistic error rate.
- **Dependencies Risk:**
  - **Extensions can silently overwrite standard fields.** The extension loop runs **after** the standard fields and does not check for collisions, so an `Ext` entry keyed `"title"` or `"status"` replaces the real one. Today the only extension is `timestamp`, set by `writeErrorResponse`, so this is latent — but the map is public and any new extension key is an unguarded overwrite.
  - **`Title` is documented as required by the RFC but omitted when empty**, so a `ProblemDetails` constructed without a title serialises to a body that is not RFC-conformant.
  - **`Status` of zero is omitted**, meaning a caller who forgets to set it produces a body with no status member while the HTTP status is whatever `writeErrorResponse` was given — the two can disagree with no signal.
  - **The `type` URI points at `https://sentrie.sh/problems/{status}`**, which mirrors the HTTP status rather than identifying the specific problem. The RFC intends `type` to distinguish problem *kinds*, so two unrelated 400s are indistinguishable by type.
  - **`NewProblemDetails` is dead code**, and the divergence between it and the literal construction in `writeErrorResponse` means a future change to one will not affect the other.
  - **Problem Details are only used for handler-level failures.** Evaluation errors come back as HTTP 200 with an `error` string in the normal response body, so clients need two error-handling paths — see [[api.handle_decision]].
