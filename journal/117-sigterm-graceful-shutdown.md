## Binaek Sarkar (binaek89@gmail.com)

### 15:40 - Pick up #117
- `main` listens for `os.Interrupt` and the uncatchable `os.Kill`, not `SIGTERM`, and `StopServer` closes connections instead of draining them.
- Picked up with #236, #237, #134, #135, #136, #150, #151 and #117 in one batch; one pickup push and a draft PR per branch before implementation.
- Decisions agreed with the maintainer before pickup: #134 treats `null` as defined and `null`/`undefined` as empty; #135 normalises known booleans to `ValueBool`; #136 keeps the envelope and reads its `state`; #150 implements the shape check.
- _Model: claude-opus-5-5_

### 15:50 - SIGTERM and graceful shutdown
- `main`: `signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)`, `os.Kill` dropped. No build tags needed: `syscall.SIGTERM` compiles for windows, linux and darwin (`go build ./...` on all three).
- Found that `newServer`'s `BaseContext` returned the serve context, so the signal itself cancelled every in-flight request and `Shutdown` would have drained nothing. It now returns `context.WithoutCancel(ctx)`; closing a connection still cancels its request, which is how the `Close` fallback drops stragglers.
- `StopServer` shuts all pairs down concurrently with `http.Server.Shutdown`, bounded by the caller's ctx; a pair that does not drain in time is logged, force-closed and its error joined, so an unclean shutdown exits 1. `serve` derives the context from `context.WithoutCancel(ctx)` with `api.DefaultShutdownGracePeriod` (10s, under Kubernetes' 30s).
- No flag: cling has no duration input, so it would need a string flag with its own parser and validation like `--eval-timeout`. Left as a constant.
- Second signal: a goroutine in `main` calls `stop()` once the context is cancelled, so the next SIGINT/SIGTERM takes the default action and kills the process. Applies to every command.
- `StartServer` already ignores `http.ErrServerClosed`. Built binary: SIGTERM and SIGINT both exit 0.
- Added `api/shutdown_test.go` (in-flight request survives cancel plus StopServer; grace expiry returns a deadline error); the test agent adds the rest.
- _Model: claude-sonnet-5-5_

### 16:10 - Shutdown tests
- `api/shutdown_test.go`: several listeners drain concurrently (all stop accepting while their requests are still in flight, then every request completes and `StopServer` returns nil); with one idle and two stuck listeners the error wraps `context.DeadlineExceeded`, names exactly the stuck addresses and the handlers that ignore cancellation are dropped by the force-close; `Shutdown` releases a listener whose `Serve` never ran. The two existing tests now use `Require().Failf` so a regression fails fast instead of hanging on a second channel read.
- `cmd/serve_signal_test.go`: builds the binary into a temp dir and runs `serve` against a minimal pack on a free port. The in-flight request is a decision POST whose body is only half sent, so the handler is blocked in the body read. SIGTERM, wait until the port refuses connections, send the rest of the body: expects 200 with decisions and exit 0. Also: an idle serve exits 0 within the grace period; a second SIGTERM during a stuck drain kills the process by signal well inside the grace period (the signal is resent on a ticker until it dies, since the first and second signal can race the restore in `main`). Skipped on Windows. No fixed sleeps as synchronisation: polling with deadlines and channels.
- Bite checks (production file reverted, test run, restored): old `main.go` fails the in-flight and idle tests; `BaseContext` returning `ctx` fails the process in-flight test and `TestStopServerLetsAnInFlightRequestFinish`; the old close-only `StopServer` fails every `StopServer` test and the process test; a sequential drain loop fails the concurrency test; `Shutdown` without the listener close fails the unserved-listener test; `main.go` without the `stop()` goroutine fails the second-signal test.
- Ruled out: asserting the first SIGTERM does not kill the process in the second-signal test (the old behaviour dies by the same signal, so it cannot be told apart without a race); a test of SIGINT (same `NotifyContext` registration as SIGTERM).
- _Model: claude-sonnet-5-5_

### 16:12 - Review round 1: StopServer no longer writes the listener slice
- Review found a data race: `StopServer` set `api.listeners = nil` while `serve`'s `StartServer` goroutine reads the slice. It predated the branch, but graceful signals make it reachable when a pod is terminated during startup.
- Fix: `StopServer` only reads the slice. A second call stays harmless because `Shutdown` is idempotent and `ErrClosed` is filtered. Added `TestStopServerDoesNotRaceWithStartServer` (Setup, `go StartServer`, immediate StopServer twice); it reported DATA RACE under `-race` before the fix and passes after.
- Docs: `api.http.md` no longer says StopServer nils the slice; the StartServer bullet now says the deferred `wg.Wait()` blocks (it returns after every `Serve`) while the unread `errChan` still loses serve errors; the BaseContext bullet notes that with `--eval-timeout 0` and no ceiling only the grace-period `Close` ends a request.
- _Model: claude-sonnet-5-5_

### 16:24 - Make the process tests deterministic
- Review found a flake in `openPartialDecision`: `net.Dial` returns once the kernel finishes the handshake, before the child accepts, and nothing waited for the handler to start before SIGTERM. If `Shutdown` closed the listener first, the queued connection was reset, failing the in-flight test and turning the second-signal test into a clean exit 0.
- The request now carries `Expect: 100-continue` and no body. Go writes `100 Continue` when the handler first reads the body, so reading it proves the request is in flight; only then is SIGTERM sent and the body written.
- The binary is built once per test process (`sync.Once`, own temp dir removed and the Once reset in `TearDownSuite`, so `-count=N` rebuilds), not once per test. Resetting matters: the first version removed the dir and a second suite run in the same process found no binary.
- Checked: `go test -race ./cmd/... -count=5` passes, lint clean, old `main.go` still fails the in-flight test.
- _Model: claude-sonnet-5-5_
