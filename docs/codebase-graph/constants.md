---
id: constants
type: System / Package
language: Go
file_path: constants/
tags: configuration, environment, conventions, cross-cutting
---

# Node: Constants (Global Conventions & Environment Keys)

## 1. Architectural Role & Intent
`constants` is a zero-dependency leaf package holding the process-wide literals that would otherwise be duplicated as magic strings: the application name, the policy and pack file extensions that drive filesystem discovery, the environment-variable names that control logging, and reserved runtime context keys. It exists so that the filesystem loader, the CLI, the logger bootstrap, and the JS runtime all agree on the same identifiers without importing each other.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| [[main]] | `DEPENDS_ON` | [[constants]] | Reads `EnvDebug` and `EnvLogLevel` to configure the `slog` JSON handler at startup. |
| [[loader]] | `LAYERED_ON` | [[constants]] | Uses `PackFileExtension` and `PolicyFileExtension` to discover pack manifests and policy sources on disk. |
| [[runtime]] | `LAYERED_ON` | [[constants]] | Seeds the execution context with `ExecutionStartTimeUnixKey`. |
| [[runtime.js]] | `LAYERED_ON` | [[constants]] | JS `time` builtins read the injected execution start time to keep clock reads deterministic within one evaluation. |

## 3. Interface Contracts & Public Surface

- **Signature:** `APPNAME = "sentrie"`
  - **Behavior:** Canonical application identifier; also the base for the policy file extension.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `PackFileExtension = "pack.toml"`
  - **Behavior:** Filename suffix identifying a policy pack manifest. This is what [[loader]] scans for when resolving a pack root.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `PolicyFileExtension = "sentrie"`
  - **Behavior:** Extension for policy source files, derived from `APPNAME`. Drives recursive policy discovery within a pack.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `EnvLogLevel = "SENTRIE_LOG_LEVEL"`
  - **Behavior:** Environment key selecting `DEBUG`/`INFO`/`WARN`/`ERROR`; anything unrecognized falls back to `INFO`.
  - **Side Effects:** Read at process start by [[main]].
  - **Exceptions:** None.

- **Signature:** `EnvDebug = "SENTRIE_DEBUG"`
  - **Behavior:** Presence-only flag (the value is irrelevant). When set it forces log level to `DEBUG` and adds `args`/`executable` attributes to every log record.
  - **Side Effects:** Causes [[main]] to **write** `EnvLogLevel` back into the process environment.
  - **Exceptions:** None.

- **Signature:** `ExecutionStartTimeUnixKey = "__executionStartTimeUnix"`
  - **Behavior:** Reserved key under which the runtime injects the evaluation's start timestamp into the execution context, so that time-dependent policy logic observes a single consistent "now" for the whole evaluation rather than a drifting wall clock.
  - **Side Effects:** None at definition; the value is populated by [[runtime.exec_ctx]].
  - **Exceptions:** None.

## 4. Operational Context & Gotchas
- **Statefulness:** Entirely stateless compile-time constants.
- **Performance/Scale Notes:** None — all values are inlined at compile time with no runtime cost.
- **Dependencies Risk:** No failure domain. The notable hazards are conventional: (1) `ExecutionStartTimeUnixKey` uses a double-underscore prefix to signal reservation, but it lives in the same namespace as user-visible context data, so a policy defining that identifier would collide; (2) changing `PolicyFileExtension` or `PackFileExtension` silently breaks discovery of every existing on-disk pack, since [[loader]] does pure suffix matching with no fallback; (3) `EnvDebug` is checked with `os.LookupEnv`, so setting it to an empty string or `false` still enables debug mode.
