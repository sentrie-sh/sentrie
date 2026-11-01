# 166-1-exact-integers

## Binaek Sarkar (binaek89@gmail.com)

### 2026-10-31 03:07 - Pick up issue #166
- Picked up as part of a wave-1 batch of eleven issues (180, 163, 161, 166, 40, 242, 162, 165, 160, 169, and a first slice of 164), each on its own branch and worktree off `main` at 3301d9a.
- Problem: All numbers are float64, so integers above 2^53 lose precision in literals and JSON facts.
- Decision (default proposed to the user on 2026-10-31, not overridden): keep integers that fit in int64 exact (JSON facts via `UseNumber`, integer literals the same); `int op int` stays exact and overflow is an error; mixed int/float promotes to float; exact integers serialise without exponent notation.
- Work runs through the implement, test and CI, and review loop with separate subagents; the parent session owns origin.

### 2026-10-31 14:48 - Map the number API and choose the representation
- Mapped every consumer of `NumberValue`, `MustNumbers` and `ValueNumber` outside tests: `constraints/number.go` (30 sites), `runtime/eval_infix.go`, `eval_unary.go`, `eval_cast.go`, `eval_access.go`, `typeref_number.go`, `builtins/collection.go` (`distinct` keys), `builtins/scalar.go` (`toIntV`, `error` formatting), `index/builtin_kind.go`, plus the fact decoders in `cmd/exec.go` and `api/handle_decision.go` and the JS argument path in `runtime/modules.go`. Estimated well under 800 hand-written lines, so no stack split.
- Found a second precision loss the issue did not name: `ast.IntegerLiteral.Value` was a `float64`, so literals were rounded in the AST before reaching `box`. Changed it to `int64`.
- Representation: kept the single kind `ValueNumber` and marked an exact integer with a zero-size `exactIntTag` in the `ref` slot (int64 bits in `u64`). Every existing `case ValueNumber` keeps working and nothing allocates. A separate kind would have meant touching every switch on the kind.
- `NumberValue()` stays as a lossy float64 view; exact paths use the new `IntValue`, `IntegralValue`, `CompareNumbers` and `Add`/`Sub`/`Mul`/`Div`/`Mod`/`Negate` in `box/number.go`. Removed `MustNumbers`, which was the lossy entry point for arithmetic.
- _Model: claude-opus-5-5, role: implementer subagent_

### 2026-10-31 14:49 - Implement exact integers end to end
- Facts: `UseNumber` on both CLI decoders and the API decoder; `FromAny` boxes a `json.Number` through `ParseNumber` (int64 if it parses, else float). A JSON number beyond float64 (`1e400`) now boxes as an infinity where the plain decoder used to reject it; noted for review.
- Literals: new `parseIntegerValueLiteral` is the `tokens.Int` prefix and constraint-literal parser; an integer beyond int64 becomes a `FloatLiteral` (the issue's fallback rule). The memoization TTL keeps the strict `parseIntegerLiteral`, so its out-of-range test still errors.
- Arithmetic: int op int is exact, overflow wraps `ErrIntegerOverflow` (as decided). Decision beyond the recorded one: `/` of two integers returns an integer only when exact and a float otherwise, so `7 / 2` stays `3.5`; `%` on two integers is Go's truncated remainder, matching `math.Mod`'s sign. `MinInt64 / -1` and `-MinInt64` are overflow errors.
- Comparison and equality: `CompareNumbers` orders int64 against float64 without rounding the integer; NaN is unordered (all relational operators false, as before). `EqualValues` uses it, so `5 == 5.0` and `2^53+1 != 2^53`.
- Output: `Any()` returns `int64` for an exact integer, so JSON and `String()` print digits with no exponent. `Any()` returning `int64` changed many test expectations from `N.0` to `int64(N)`; these are representation-only.
- Constraints: bounds, `eq`/`neq`, `in`/`not_in` and `range` compare through `CompareNumbers`/`EqualValues` (NaN behaviour preserved); `even`, `odd` and `multiple_of` use integer arithmetic for exact integers.
- Builtins: `distinct` keys an integral number by its exact integer so ints and equal floats collide and big ints stay distinct; `error(...)` wraps an int64 argument so both `%d` and `%.2f` work (`%.2f` worked before when every number was a float64).
- JS boundary decision: goja rounds an int64 beyond 2^53 to a double on `ToValue` (checked with a sandbox run, then reset). Chose to make an unsafe integer (|n| > 2^53-1) anywhere in an argument tree an error, mirroring #154's "error rather than a lossy value" on the return side. Alternative was documented lossy rounding; it would keep pass-through of whole fact objects to modules working but lets two IDs compare equal in JS. Flagged as an open question.
- Rewrote legacy license headers on touched files (`parser/literal.go`, `runtime/eval_infix.go`, `runtime/eval_unary.go`, `ast/literal_integer.go`, and touched tests).
- Updated graph nodes: box, box.value, parser.primary, runtime.eval_infix, runtime.eval_unary, runtime.eval_cast, runtime.typeref_number, runtime.modules, constraints, api.handle_decision, cmd.
- Tests: `box/number_test.go`, `runtime/exact_integers_test.go` (end to end and JS boundary), `constraints/number_exact_test.go`, `builtins/exact_integers_test.go`, `api/exact_integers_test.go`, parser large-literal test. `go test ./...` and `make docs-graph` pass; CLI smoke test echoes `9007199254740993`. `make lint` panics inside golangci-lint (binary built with go1.26, module needs go1.27), an environment issue; `go vet ./...` passes.
- _Model: claude-opus-5-5, role: implementer subagent_

### 2026-11-01 20:58 - Settle the JS boundary question
- The maintainer chose the error: an integer beyond ±(2^53-1) anywhere in an argument tree passed to a TypeScript module raises an Error rather than rounding. This closes the open question above; documented lossy rounding is rejected because two distinct IDs could compare equal inside JS.
- _Model: claude-opus-5-5, role: parent orchestrator_

### 2026-11-01 02:32 - Test and CI pass for exact integers
- Added `runtime/exact_integers_edge_test.go` (relational operators and int/float equality beyond 2^53, `in`, `contains`, `distinct`, list and dict equality, overflow at both ends of int64 for `+ - *`, JSON rendering of `2.0`, `1.5` and `-0.0`, shape `@min`/`@max` with large bounds, memoized calls keyed apart for IDs one apart, a huge memo TTL literal, and the JS boundary: negative, list, nested dict, arithmetic result and fact pass-through errors for |n| > 2^53-1, safe values and floats cross), `cmd/exec_exact_integers_test.go` (`--facts`, `--fact-file`, flag-over-file merge, table output) and fixture `lang_test/0056-exact-integers.sentrie`.
- Bite proof: cmd tests fail with cmd/exec.go from main; JS edge tests fail with the safe-range guard in runtime/modules.go disabled; the edge and constraint tests and the fixture fail against a main checkout.
- Gate: make test, race on non-runtime packages, gofmt, docs-graph and the em dash check pass. Lint fails with 2 errcheck findings in `builtins/scalar.go` (`fmt.Fprintf` results unchecked at lines 253 and 255), non-test code, left for the implementer.
- Found: JSON `1e400` now decodes to +Inf where main rejected it with a 400/error; the CLI `--output json` and the HTTP API then return an empty body (HTTP 200, Content-Length 0) because +Inf cannot be marshalled and the encode error is dropped. Reported as a bug, not covered by a queued test.
- Minor: the literal `-9223372036854775808` is a float (negation of a float literal), so it echoes as `-9223372036854776000`; `-9223372036854775807 - 1` is the exact way to write MinInt64.
- _Model: claude-sonnet-5-5, role: test and CI subagent_

### 2026-11-01 02:36 - Fix the test and CI findings
- Lint: the two `fmt.Fprintf` calls in `formatInt.Format` now discard their results explicitly (`_, _ =`).
- Regression fixed: with `UseNumber` the decoder stopped rejecting a JSON number float64 cannot hold, so `1e400` became +Inf and produced an empty CLI and API body. `box.ParseNumber` now returns an error for a range overflow instead of an infinity, and the new `box.CheckJSONNumbers` walks decoded facts; both CLI decoders return `facts: number 1e400 is out of range` (or `fact file: ...`), and the API returns 400 `Invalid JSON`. `FromAny` cannot return an error, so the check lives at the decoders; a `json.Number` that bypasses them still boxes as the float strconv returns.
- Encoding errors surfaced: `formatOutputJSON` returns the encoder error (an infinity from arithmetic, for example `1e308 * 10`), so `exec --output json` fails instead of printing nothing and exiting 0. The API encodes into a buffer before writing the status, so the same case is a 500 problem response rather than a 200 with an empty body.
- MinInt64: `parseUnaryExpression` folds `-` applied directly to the integer token `9223372036854775808` into an `IntegerLiteral` holding `math.MinInt64`. Nothing else is folded, so constraint arguments and other negations keep their unary shape.
- JS boundary wording aligned with the maintainer's decision: an integer beyond Number.MAX_SAFE_INTEGER (2^53-1) in magnitude, 2^53 included, is an error. Added assertions that ±2^53 error.
- Tests: `cmd/exec_number_range_test.go` (both fact paths and the JSON encoding error), API tests for 400 on out-of-range facts and 500 on an unencodable decision, `TestCheckJSONNumbers`, and a parser test for the MinInt64 fold. `go test ./...`, `make docs-graph` and `make lint` with the go1.25 golangci-lint pass (0 issues).
- Queue: no new lines; the new paths are folded into the existing lines for the same logical change (`parser/unary.go` and its graph node into the literal commit, `cmd/exec_number_range_test.go` into the UseNumber commit). Edits to files already queued travel with their lines.
- _Model: claude-opus-5-5, role: implementer subagent_

### 2026-11-01 07:33 - Address review round 1
- W1: `multiple_of` took the exact branch only when both sides were exact ints, so `@multiple_of(2.0)` accepted 9007199254740993. Both sides now read `IntegralValue()`, so an integral float divisor or value takes the exact branch. `even` and `odd` switched to `IntegralValue()` too, for consistency only: float arithmetic on an integral float was already exact. No other number constraint needs integer arithmetic (the rest compare through `CompareNumbers`).
- W2: fixed the `ParseNumber` entry in the box node; it returns `Undefined()` with the error, never an infinity.
- S1: `calculateHashKey` now passes each boundary value through `canonicalHashNumbers`, which rewrites integral float64s inside int64 range as int64, through lists and maps. `-0` keeps its own key because the existing `TestCalculateHashKeyNumericEdges` requires it, and a module can tell `-0` from `0` (`1 / -0` is `-Infinity`). Containers are copied because a host document's `[]any` or map passes through `TryToBoundaryAny` unchanged.
- S2: added the test agent's files to the PR testing notes.
- S3: `now()` and `now_ms()` return `box.Int`. Nothing downstream needs a float: the only consumer was the golden test, now `int64`. `now_ms()` is about 1.7e12, well inside the JS safe range.
- Bite proof: with each fix reverted, `multiple_of integral float divisor`, `TestNow_ExactIntegers` and `TestMemoKeyTreatsEqualNumbersAlike` fail. The new `even`/`odd` integral-float cases pass either way, as expected.
- Gate: `go test ./...`, `make docs-graph`, and `make lint` with the go1.25 golangci-lint all pass (0 issues). Graph nodes runtime.eval_call and builtins updated, plus `builtins/ADDING_BUILTINS.md`.
- Queue: `runtime/eval_call.go` and its graph node are on the box arithmetic line; `builtins/ADDING_BUILTINS.md` and the builtins node are on the builtins line, renamed "Keep exact integers exact in distinct, error, flatten and now". `builtins/table_test.go` stays on the box line, so at that intermediate commit the `now` golden case fails until the builtins line lands.
- _Model: claude-opus-5-5, role: implementer subagent_
