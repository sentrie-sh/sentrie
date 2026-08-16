---
id: infra.os_environment
type: Infrastructure
language: N/A
file_path: (external)
tags: infrastructure, boundary, configuration, permissions, security
---

# Node: OS Environment (Process Environment Variables)

## 1. Architectural Role & Intent
The process environment is Sentrie's out-of-band configuration channel and, separately, a data source that policy-authored JavaScript can read. Those two uses have different threat models and are handled by different code, which is why this is a distinct node rather than an implementation detail of either consumer.

As configuration, [[main]] reads two keys defined in [[constants]] to bootstrap logging before anything else runs. As policy-visible data, [[runtime.js.stdlib]] projects a **filtered** view of the environment into every JavaScript VM as the `env` global — filtered per-variable against the pack's declared permissions, so a pack sees only what its manifest asked for.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| [[main]] | `MUTATES` | `infra.os_environment` | Sets `EnvLogLevel` to `DEBUG` when debug mode is detected. |
| [[runtime.js.stdlib]] | `READS_FROM` | `infra.os_environment` | `os.Environ()` supplies the candidate variables. |

## 3. Interface Contracts & Public Surface

- **Signature:** `SENTRIE_LOG_LEVEL`
  - **Behavior:** Selects `DEBUG`/`INFO`/`WARN`/`ERROR`. Anything unrecognised falls back to `INFO`.
  - **Side Effects:** Read once at process start by [[main]].
  - **Exceptions:** None — there is no error path, only the silent fallback.

- **Signature:** `SENTRIE_DEBUG`
  - **Behavior:** Presence-only flag; the value is never inspected. When present it forces `DEBUG` and adds `args`/`executable` attributes to every log record.
  - **Side Effects:** Causes [[main]] to **write** `SENTRIE_LOG_LEVEL` back into the process environment, so the variable observed by the rest of the process is not necessarily the one the operator set.
  - **Exceptions:** None.

- **Signature:** The `env` global inside a JavaScript VM
  - **Behavior:** An object containing only those variables the pack manifest permits, decided one key at a time by `PackFile.Permissions.CheckEnvAccess`. Installed by [[runtime.js.stdlib]] before any module body runs.
  - **Side Effects:** None on the host environment — the projection is a copy.
  - **Exceptions:** A denied variable is absent rather than throwing, so policy code cannot distinguish "not permitted" from "not set".

## 4. Operational Context & Gotchas
- **Statefulness:** Process-global and mutable. [[main]] writes to it during startup, which is the only write in the system.
- **Performance/Scale Notes:** `os.Environ()` is called per VM setup, not per read, so the cost is proportional to VM count rather than to policy access.
- **Dependencies Risk:**
  - **`SENTRIE_DEBUG` is presence-checked with `os.LookupEnv`**, so setting it to the empty string or to `false` still enables debug mode — the opposite of the operator's evident intent.
  - **Debug logging widens the blast radius of the `env` projection.** Debug mode adds process `args` to every log record, and the environment filter is the only thing standing between a pack and the host's secrets.
  - **The permission filter is the sole boundary.** There is no denylist for well-known secret-bearing variables, so a pack manifest that requests a broad set gets it.
