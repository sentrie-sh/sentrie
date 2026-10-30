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
