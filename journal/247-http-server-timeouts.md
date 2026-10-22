## Binaek Sarkar (binaek89@gmail.com)

### 23:45 - Pick up #247
- Branch cut from `243-request-timeout-header`: #243 resizes the server's write timeout in `api/http.go`, which is where these limits go too.
- Plan: named constants for `ReadHeaderTimeout`, `IdleTimeout` and `MaxHeaderBytes` on the server, no new flags, a test that a slow header write gets its connection closed, and an update to the `api.http` graph node.
- Opened as a draft PR before implementation, per the issue workflow.
- _Model: claude-opus-5-5_
