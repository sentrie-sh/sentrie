---
id: infra.filesystem
type: Infrastructure
language: N/A
file_path: (external)
tags: infrastructure, boundary, io, untrusted-input, security
---

# Node: Filesystem (Host Storage Boundary)

## 1. Architectural Role & Intent
The host filesystem is the only source of policy input that Sentrie reads without a network hop, and it is where every artefact the engine executes originates: pack manifests, `.sentrie` policy sources, and JavaScript/TypeScript modules. It is modelled as an explicit node because reads here cross a trust boundary - the bytes are attacker-controlled in any deployment where pack authorship is not the same as operator identity - and because the containment rules that are supposed to confine those reads to a pack root are enforced inconsistently across the two readers.

Two subsystems touch it, with different guarantees. [[loader]] walks a pack directory tree during startup and produces the parse inputs. [[runtime.js.registry]] reads module files lazily, during evaluation, in response to `require` specifiers that originate in policy source; `@local/` and relative specifiers are confined to the pack root, including after symlink resolution.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| [[loader]] | `READS_FROM` | `infra.filesystem` | `os.Stat`, `os.ReadFile`, and `fs.WalkDir` over the pack directory tree. |
| [[runtime.js.registry]] | `READS_FROM` | `infra.filesystem` | `os.ReadFile` for on-disk modules; `os.Stat` for extension probing. |

## 3. Interface Contracts & Public Surface

- **Signature:** Pack discovery (via [[loader]])
  - **Behavior:** Locates a `pack.toml` manifest, then recursively walks the containing directory collecting files whose extension matches `PolicyFileExtension`. Both suffixes come from [[constants]].
  - **Side Effects:** Read-only. Sentrie never writes to the filesystem outside of `sentrie init`.
  - **Exceptions:** Missing manifest, unreadable file, and malformed TOML all surface as load errors before any evaluation begins.

- **Signature:** Module resolution (via [[runtime.js.registry]])
  - **Behavior:** Resolves a `require` specifier to a path, probes candidate extensions with `os.Stat`, and reads the winning file for compilation.
  - **Side Effects:** Read-only, but the bytes read are subsequently **executed** as JavaScript.
  - **Exceptions:** An unresolvable specifier fails the evaluation that triggered it, not the load.

## 4. Operational Context & Gotchas
- **Statefulness:** The filesystem is mutable underneath a running process. Pack contents are read once at load, but module reads happen lazily during evaluation, so a file edited mid-process can be picked up by a later evaluation while earlier ones used the previous bytes.
- **Performance/Scale Notes:** `fs.WalkDir` cost is proportional to the pack tree, paid once at startup. Module reads are cached by [[runtime.js.registry]] after first compilation.
- **Dependencies Risk:**
  - **There is no read-size limit anywhere.** Neither reader bounds the bytes it will pull in, so a large file in a pack tree is fully materialised in memory.
