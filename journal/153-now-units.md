# 153-now-units

## Binaek Sarkar (binaek89@gmail.com)

### 2026-10-21 14:51 - Start the now() units change
- Cut from `139-2-jwt-asymmetric`, the top of the #139 stack: this branch reuses `executionStartUnix` from `139-1-jwt-claims` so `@sentrie/time.now()` and the JWT `exp`/`nbf` checks read the execution start one way.
- Decision (from the task brief): unify on seconds. Builtin `now()` returns Unix seconds, a new `now_ms()` returns milliseconds, both pinned to the execution start; `@sentrie/time.now()` throws instead of falling back to the wall clock.
- _Model: claude-opus-5-5_
