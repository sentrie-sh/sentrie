---
id: pack
type: System / Package
language: Go
file_path: pack/
tags: manifest, packaging, permissions, semver, configuration
---

# Node: Pack (Policy Pack Manifest Model)

## 1. Architectural Role & Intent
`pack` is the pure data model for a Sentrie policy pack: the `sentrie.pack.toml` manifest (`PackFile`) and the in-memory pairing of that manifest with its parsed programs (`Pack`). It defines the pack's identity and version, the engine version constraint it requires, and — critically — the **permission grants** (filesystem read, network, environment variables) that bound what JavaScript modules invoked from a policy are allowed to touch. It contains no I/O: discovery, decoding, and schema validation all live in [[loader]].

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `pack` | `LAYERED_ON` | [[ast]] | `Pack.Programs` is `[]*ast.Program` — the parsed policy sources belonging to this manifest. |
| `pack` | `IMPORTS` | `ext.masterminds.semver` | `PackInformation.Version` is a `*semver.Version`; `Engine.Sentrie` is a `*semver.Constraints`. |
| `pack` | `IMPORTS` | `std.encoding/json`, `std.slices` | Custom JSON codec for the semver constraint; membership check for env permissions. |
| [[loader]] | `MUTATES` | [[pack]] | `LoadPack` decodes TOML into a `PackFile` and sets `Location` to the manifest's directory. |
| [[loader]] | `READS_FROM` | [[pack]] | `LoadPrograms` reads `PackFile.Location` to decide where to walk for `.sentrie` sources. |
| [[index.package]] | `LAYERED_ON` | [[pack]] | Indexing is performed against a pack's manifest plus programs. |
| [[runtime.js]] | `READS_FROM` | [[pack]] | The JS standard library consults `Permissions` before granting env/file/network access to sandboxed modules. |
| [[cmd]] | `LAYERED_ON` | [[pack]] | `sentrie init` writes a new manifest via `NewPackFile`; `exec`/`validate`/`serve` load and carry one. |

## 3. Interface Contracts & Public Surface

- **Signature:** `Pack` struct — `{ Pack *PackFile, Programs []*ast.Program }`
  - **Behavior:** The fully-loaded unit of policy: manifest plus every parsed program under it. This is the value handed to indexing and execution.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `PackFile` struct — `{ SchemaVersion *SentrieSchema, Pack *PackInformation, Permissions *Permissions, Engine *Engine, Metadata map[string]any, Location string }`
  - **Behavior:** Direct TOML/JSON mapping of `sentrie.pack.toml`. `Location` is tagged `toml:"-"`/`json:"-"` — it is **not** a manifest field but a runtime-populated absolute directory path, injected by [[loader]] and used as the root for source discovery and relative `use` resolution.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `NewPackFile(name: string) -> *PackFile`
  - **Behavior:** Scaffolds a manifest at schema version `1` and pack version `0.0.1`. Used by `sentrie init`.
  - **Side Effects:** None.
  - **Exceptions:** **Panics** if the hardcoded default version fails to parse (`semver.MustParse`) — unreachable in practice.

- **Signature:** `SentrieSchema` — `{ Version uint64 }` / `PackInformation` — `{ Name, Version, Description, License, Repository, Authors }`
  - **Behavior:** Manifest identity. `Name` must be a dotted identifier (validated in [[loader]], not here); `Version` is a parsed semver.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `Engine` — `{ Sentrie *semver.Constraints }` with `MarshalJSON()` / `UnmarshalJSON(data)`
  - **Behavior:** Declares which Sentrie engine versions the pack supports. The custom codec exists solely because `semver.Constraints` implements neither JSON interface: it serializes to its string form and re-parses on the way back. This round-trip is what makes the manifest validatable against the embedded JSON Schema in [[loader]].
  - **Side Effects:** None.
  - **Exceptions:** `UnmarshalJSON` returns the `encoding/json` error, or a semver parse error for a malformed constraint string.

- **Signature:** `Permissions` — `{ FSRead []string, Net []string, Env []string }`
  - **Behavior:** Allowlists that bound sandboxed execution. Declarative only — this package never enforces them.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `(*Permissions).CheckEnvAccess(name: string) -> bool`
  - **Behavior:** Exact-match membership test against the `Env` allowlist. The one enforcement helper in the package, consulted by [[runtime.js]] before exposing an environment variable.
  - **Side Effects:** None.
  - **Exceptions:** **Panics on a nil receiver** — it dereferences `p.Env`. Because `Permissions` is `omitempty` and legitimately absent from a manifest, callers must nil-guard before calling.

## 4. Operational Context & Gotchas
- **Statefulness:** Plain data structs; a `PackFile` is loaded once and treated as read-only configuration for the lifetime of a command or server request.
- **Performance/Scale Notes:** Negligible. Decoded once per invocation. `CheckEnvAccess` is a linear scan, which is fine for hand-written allowlists but is called per env lookup from JS — keep allowlists short.
- **Dependencies Risk:**
  - **Optional pointers everywhere.** `Permissions`, `Engine`, `Pack`, and `SchemaVersion` are all pointers marked `omitempty`; a manifest without a `[permissions]` table yields a nil pointer, and calling `CheckEnvAccess` on it panics. Permission checks must be written as "deny when nil", and treating nil as "allow all" would be a sandbox escape.
  - **Enforcement lives elsewhere.** This package only *declares* permissions. Auditing what a pack can actually reach means reading [[runtime.js]], not this node.
  - **`Location` is trusted.** It is populated from disk by [[loader]] and used as a filesystem root; anything that constructs a `PackFile` by hand must set it deliberately, since an empty `Location` makes program discovery walk the process working directory.
  - **Engine constraint is advisory here.** `Engine.Sentrie` is parsed and stored but not checked against the running binary in this package.
