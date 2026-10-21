---
id: capabilities
type: System / Package
language: Go
file_path: capabilities/
tags: sandboxing, configuration, security, operator-grant, allowlist
---

# Node: capabilities (Operator Capabilities Grant)

## 1. Architectural Role & Intent
The operator's grant: a deployment-level JSON file, passed with `--capabilities`, that limits which builtins and which `@sentrie/*` modules a pack may use, and whether it may load its own TypeScript or JavaScript files (#168). It is the counterpart of [[pack]]'s `[permissions]` table, which is the pack's request. The package is a leaf: it parses and answers membership questions. The checks live in [[index]] (load time, with spans) and in [[runtime.eval_call]] and [[runtime.js.registry]] (runtime). The file format is documented in `capabilities/FORMAT.md`.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `capabilities` | `LAYERED_ON` | [[constants]] | `APPNAME` forms the required `@sentrie/` prefix of module entries. |
| `capabilities` | `IMPORTS` | `std.encoding/json` | Strict decoding with `DisallowUnknownFields` and a trailing-data check. |
| `capabilities` | `READS_FROM` | [[infra.filesystem]] | `Load` reads the file named by `--capabilities`. |
| [[cmd]] | `CALLS` | `capabilities` | `exec`, `serve` and `validate` load the file, check the index, and hand the grant to the executor. |
| [[index]] | `READS_FROM` | `capabilities` | `CheckCapabilities` asks `AllowsBuiltin`, `AllowsModule` and `AllowsLocalModules` for every builtin call and `use`. |
| [[runtime.executor]] | `READS_FROM` | `capabilities` | `WithCapabilities` stores the grant for builtin dispatch and installs it on the JS registry. |
| [[runtime.js.registry]] | `READS_FROM` | `capabilities` | `PrepareUse` and `LoadRequire` refuse ungranted modules, including `require()` inside TypeScript. |

## 3. Interface Contracts & Public Surface

- **Signature:** `Capabilities` - `{ Version int, Builtins *[]string, Modules *[]string, LocalModules *bool }`
  - **Behavior:** JSON keys `version`, `builtins`, `modules`, `local_modules`. A nil field (absent or `null`) leaves that dimension unrestricted; an empty list grants nothing. Module entries are full specifiers (`@sentrie/hash`).
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `Parse(data []byte) -> (*Capabilities, error)` / `Load(path string) -> (*Capabilities, error)` / `CurrentVersion = 1`
  - **Behavior:** `Parse` rejects unknown fields, trailing data, any version other than `CurrentVersion`, empty builtin names, and module entries without the `@sentrie/` prefix. It deliberately does not check names against this build's builtins and modules, so a file written for a newer build loads and an unknown name grants nothing. `Load("")` returns `(nil, nil)`, meaning no file was given.
  - **Side Effects:** `Load` reads the file.
  - **Exceptions:** Decode, version and entry errors prefixed `capabilities:`; `Load` prefixes the path.

- **Signature:** `(*Capabilities).AllowsBuiltin(name) -> bool` / `AllowsModule(key) -> bool` / `AllowsLocalModules() -> bool`
  - **Behavior:** Membership tests. Every method is nil-safe and a nil receiver grants everything.
  - **Side Effects:** None.
  - **Exceptions:** None.

## 4. Operational Context & Gotchas
- **Statefulness:** A parsed grant is read-only after `Load` and is shared by the index check, the executor and the JS registry.
- **Nil means allow, unlike pack permissions.** [[pack]]'s `Permissions` must be read as "deny when nil", because a missing `[permissions]` table must not open the sandbox. A missing capabilities file must keep today's behaviour, so here nil grants everything. Do not copy one convention into the other.
- **The load-time check is only half.** `index.CheckCapabilities` cannot see `require()` calls inside TypeScript, so the registry gate is what holds TypeScript to the grant. Dropping either half opens a gap.
- **Membership is a linear scan.** That is fine for hand-written lists, and it runs once per builtin dispatch.
- **Pack permissions are not yet intersected with the grant.** #157 tracks enforcing `[permissions]`. A future change could refuse a pack whose declared `env`, `fs_read` or `net` requests exceed the grant.
