## Binaek Sarkar (binaek89@gmail.com)

### 23:45 - Pick up #247
- Branch cut from `243-request-timeout-header`: #243 resizes the server's write timeout in `api/http.go`, which is where these limits go too.
- Plan: named constants for `ReadHeaderTimeout`, `IdleTimeout` and `MaxHeaderBytes` on the server, no new flags, a test that a slow header write gets its connection closed, and an update to the `api.http` graph node.
- Opened as a draft PR before implementation, per the issue workflow.
- _Model: claude-opus-5-5_

### 2026-10-23 00:05 - Implement the connection limits
- Moved server construction into `HTTPAPI.newServer`, so `Setup` and the tests build the same server.
- Named constants: `serverReadTimeout` (30s, unchanged), `serverReadHeaderTimeout` (5s), `serverIdleTimeout` (60s), `serverMaxHeaderBytes` (64 KiB, down from net/http's 1 MiB default). No flags, as the issue suggested; they can come later if an operator needs them.
- 64 KiB leaves ample room for real headers. The `Sentrie-Timeout` header itself is capped at 32 bytes by #243.
- The header deadline is an unexported `readHeaderTimeout` field, defaulting to the constant, so the slow-header test can shorten it to 100ms instead of waiting 5s.
- Tests:
  - `TestServerSetsConnectionLimits` checks every limit, and that the header deadline is tighter than the whole read.
  - `TestServerClosesASlowHeaderConnection` opens a raw TCP connection, sends a request line with no end of headers, and checks the server closes it at the header deadline rather than the client's own 5s read deadline firing. Without the fix the header deadline falls back to the 30s read timeout and the test fails.
  - Both pass under `-race -count=3`.
- Updated the `api.http` graph node, which recorded the gap.
- _Model: claude-opus-5-5_
