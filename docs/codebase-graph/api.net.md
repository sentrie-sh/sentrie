---
id: api.net
type: Module / File
language: Go
file_path: api/net.go
tags: networking, address-resolution, configuration, defect
---

# Node: api.resolveBindings (Listen Address Resolution)

## 1. Architectural Role & Intent
Translates the operator-facing `--http-listen` values into concrete `host:port` strings. It supports six convenience names covering the loopback/all-interfaces and IPv4/IPv6 matrix, plus arbitrary explicit hosts, so common deployments do not require remembering interface literals.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `api.net` | `IMPORTS` | `ext.binaek.gocoll` | `collection.Map` over the explicit-address list. |
| `api.net` | `IMPORTS` | `ext.golang.x-exp` | `slices` membership check against the predefined listen names. |
| [[api.http]] | `CALLS` | `api.net` | `Setup` resolves bindings before opening listeners. |
| [[cmd]] | `DEPENDS_ON` | `api.net` | The `--http-listen` flag's accepted values are defined by this function. |

## 3. Interface Contracts & Public Surface

- **Signature:** `resolveBindings(port int, listen []string) -> ([]string, error)`
  - **Behavior:** If any element is one of the six predefined names, the list must contain **exactly one** element. A predefined name maps to a fixed host; anything else is treated as an explicit host and joined with the port.
  - **Side Effects:** None - pure.
  - **Exceptions:** `when using predefined listen addresses, there must be exactly one address`.

- **Signature:** The predefined names
  - **Behavior:**
    - `local` → `localhost` - resolver-dependent, may yield either family.
    - `local4` → `127.0.0.1`
    - `local6` → `::1` - **currently produces a malformed address**.
    - `network` → empty host, meaning all interfaces on both families.
    - `network4` → `0.0.0.0`
    - `network6` → `::` - **currently produces a malformed address**.
  - **Side Effects:** None.
  - **Exceptions:** None.

## 4. Operational Context & Gotchas
- **Statefulness:** Pure function.
- **Performance/Scale Notes:** Negligible; called once at startup.
- **Dependencies Risk:**
  - **`local6` and `network6` cannot bind.** The hosts are passed to `net.JoinHostPort` **already bracketed** (`"[::1]"`, `"[::]"`), and `JoinHostPort` brackets any host containing a colon - producing `[[::1]]:7529` and `[[::]]:7529`. `net.Listen` rejects both, so IPv6-only serving is impossible via the predefined names. The IPv4 and unbracketed cases are correct. Filed as [#116](https://github.com/sentrie-sh/sentrie/issues/116).
  - **`listen[0]` is read without a length check**, so an empty slice panics. The CLI default of `["local"]` prevents this in practice, but the function is package-level and offers no guard.
  - **The predefined/explicit distinction is decided by `listen[0]` alone.** The validation loop confirms a predefined name implies a single-element list, so the two agree - but the switch would silently mishandle a list whose predefined name is not first if that validation ever changed.
  - **Explicit IPv6 addresses must be supplied unbracketed** for `JoinHostPort` to bracket them correctly. Nothing documents this, and supplying the bracketed form produces the same double-bracket failure.
  - **`network` uses an empty host**, which binds all interfaces on both families where the platform supports dual-stack sockets - a meaningful behavioural difference from `network4`/`network6` that the flag description does not convey.
