---
id: ext.dependencies
type: System / Package
language: Go
file_path: go.mod
tags: third-party, dependencies, supply-chain, manifest
---

# Node: ext.dependencies (Third-Party Dependency Manifest)

## 1. Architectural Role & Intent
The complete external surface of the Sentrie module, declared in `go.mod` against **Go 1.25.0**. The dependency set is deliberately small for a language runtime — 14 direct requirements — and every one of them sits at a specific architectural seam. This node exists so an agent can reason about which third-party failure affects which subsystem without re-reading the manifest.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| [[runtime.js]] | `DEPENDS_ON` | `ext.goja` | `github.com/dop251/goja` — the JavaScript interpreter. |
| [[runtime.js.tscompile]] | `DEPENDS_ON` | `ext.esbuild` | `github.com/evanw/esbuild` — TypeScript to CommonJS transpilation. |
| [[runtime.executor]] | `DEPENDS_ON` | `ext.perch` | `github.com/binaek/perch` — TTL cache backing module bindings and call memoization. |
| [[runtime.modules]] | `DEPENDS_ON` | `ext.puddle` | `github.com/jackc/puddle/v2` — the JavaScript VM pool. |
| [[cmd]] | `DEPENDS_ON` | `ext.cling` | `github.com/binaek/cling` — CLI command, flag, and argument framework. |
| [[loader]] | `DEPENDS_ON` | `ext.gojsonschema` | `github.com/xeipuuv/gojsonschema` — pack manifest schema validation. |
| [[loader]] | `DEPENDS_ON` | `ext.toml` | `github.com/pelletier/go-toml/v2` — pack manifest parsing and encoding. |
| [[pack]] | `DEPENDS_ON` | `ext.semver` | `github.com/Masterminds/semver/v3` — pack version and engine constraint checking. |
| [[runtime.eval_call]] | `DEPENDS_ON` | `ext.hashstructure` | `github.com/mitchellh/hashstructure/v2` — memoization key derivation. |
| [[api.middleware]] | `DEPENDS_ON` | `ext.google_uuid` | `github.com/google/uuid` — request and instance identifiers. |
| [[api.net]] | `DEPENDS_ON` | `ext.gocoll` | `github.com/binaek/gocoll` — generic collection helpers. |
| [[box]] | `DEPENDS_ON` | `ext.structs` | `github.com/fatih/structs` — struct-to-map reflection at the value boundary. |

## 3. Interface Contracts & Public Surface

- **Signature:** Direct requirements — the load-bearing set
  - **`goja`** (pinned to a commit, not a tag) — a pure-Go ECMAScript 5.1 interpreter with some later features. **No JIT**, so JavaScript is markedly slower than Go. Supplies the `Interrupt` mechanism that makes policy JavaScript cancellable.
  - **`esbuild` v0.25.11** — used only through `api.Transform`, never the bundler. Targets ES2019.
  - **`perch` v0.0.3 / `puddle` v2.2.2** — caching and pooling. `perch` is a v0.0.x dependency owned by the same author as this project.
  - **`cling` v0.3.8** — likewise a pre-1.0 dependency by the same author; the entire CLI surface rests on it.
  - **`gojsonschema` v1.2.0** — JSON Schema validation of the pack manifest.
  - **`go-toml/v2` v2.2.4** — manifest read and write.
  - **`semver/v3` v3.4.0** — SemVer parsing and constraint matching for engine compatibility.
  - **`testify` v1.11.1** — test-only, but a direct requirement.
  - **`golang.org/x/exp`** — used for `slices` helpers in [[api.net]].

- **Signature:** Notable indirect requirements
  - **`dlclark/regexp2`** — goja's regex engine, which is .NET-flavoured rather than RE2; relevant because policy `matches` and the `regexp` constraint use Go's `regexp` while JavaScript regexes go through this. The two dialects differ.
  - **`go-sourcemap/sourcemap`** — pulled in by goja; note Sentrie parses source maps but never uses them (see [[runtime.js.tscompile]]).
  - **`olekukonko/tablewriter`** — the `exec --output table` renderer.

## 4. Operational Context & Gotchas
- **Statefulness:** N/A — this node describes the manifest.
- **Performance/Scale Notes:** `goja` is the dominant performance consideration in the whole system: interpreted, non-JIT, single-threaded per VM. Any policy hot path expressed in JavaScript rather than in a native builtin pays for it. `esbuild` is fast but runs on first load of each module.
- **Dependencies Risk:**
  - **`goja` is pinned to a bare commit, not a released version.** Reproducibility depends entirely on the module proxy and `go.sum`. Upgrading requires re-verifying the whole JavaScript surface, since goja's version numbering gives no compatibility signal.
  - **Three direct dependencies are pre-1.0 and authored by this project's own maintainer** — `perch` (v0.0.3), `cling` (v0.3.8), and `gocoll` (v0.2.0). They sit on critical paths: caching, the entire CLI, and address resolution. A breaking change is a coordinated release rather than an external event, which is an advantage, but the version numbers offer no stability guarantee to downstream consumers.
  - **Two regex engines are in play.** Go's RE2-based `regexp` serves the policy language's `matches` operator and the string constraints; `regexp2` serves JavaScript. RE2 has no backtracking (and so no catastrophic backtracking), while `regexp2` does — meaning a regex in module JavaScript can hang in ways the same pattern in a policy cannot. The goja `Interrupt` watchdog is the only mitigation.
  - **`fatih/structs` v1.1.0 is effectively unmaintained**, with no release in years. It sits at the value boundary in [[box]].
  - **`xeipuuv/gojsonschema` is likewise dormant** and implements an older JSON Schema draft; manifest validation is bounded by whatever draft it supports.
  - **The dependency count is a security asset.** A policy engine benefits from a small supply chain, and 14 direct dependencies with no web framework, no ORM, and no logging library beyond the standard `slog` is a deliberate and defensible position worth preserving.
