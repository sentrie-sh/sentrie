# 153-now-units

## Binaek Sarkar (binaek89@gmail.com)

### 2026-10-21 14:51 - Start the now() units change
- Cut from `139-2-jwt-asymmetric`, the top of the #139 stack: this branch reuses `executionStartUnix` from `139-1-jwt-claims` so `@sentrie/time.now()` and the JWT `exp`/`nbf` checks read the execution start one way.
- Decision (from the task brief): unify on seconds. Builtin `now()` returns Unix seconds, a new `now_ms()` returns milliseconds, both pinned to the execution start; `@sentrie/time.now()` throws instead of falling back to the wall clock.
- _Model: claude-opus-5-5_

### 2026-10-21 14:53 - Seconds everywhere, now_ms(), and no fallback
- Builtin `now()` (`builtins/scalar.go`) now returns `ExecutionStart().Unix()`; the new `now_ms()` returns `UnixMilli()`. Both are `DeriveSafe`, both read the same pinned `createdAt`. Why seconds: JWT `exp`/`nbf` and every `@sentrie/time` function are already seconds, so only the builtin changes unit.
- `@sentrie/time.now()` reads through `executionStartUnix` (added on `139-1-jwt-claims`) and throws `execution start time is not set on this VM` when the global is missing. The edit to `builtin_time.go` is confined to the `now` body, so #154's change to how the global is read lands in `builtin_clock.go` alone.
- Tests: `builtins/table_test.go` pins both units against a fixed instant with a sub-second part; `JSTestSuite.TestTimeNowIsPinned` covers the pinned value and the missing-global throw; `RuntimeTestSuite.TestNowUnitsAgree` checks from policy code that `now() == time.now()` and that `now_ms()` is the same instant.
- Swept fixtures and examples: no `.sentrie` fixture calls `now()`; `example_pack/minimal.sentrie` imports `@sentrie/time.now` but never calls it, so nothing there changes.
- Not changed, reported instead: `@sentrie/js` exports `now = Date.now`, a third `now` that is wall-clock milliseconds and not pinned.
- The user docs for builtins live in the separate `website` repo (`src/content/docs/reference/functions.md`) and need the same unit change.
- _Model: claude-opus-5-5_
