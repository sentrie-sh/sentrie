# 184-replace-x-exp-with-std

## Binaek Sarkar (binaek89@gmail.com)

### 2026-10-13 16:20 - Replace golang.org/x/exp with the standard library
- Follow-up from #181's list, filed as #184 because it changes product code (`pr-conventions`, `current-libraries`: migrations get their own issue).
- Branched off `main` (`11070ea`). It touches nothing that #181 adds.
- The work was done in a worktree named `exp-to-std`, then renamed to `184-replace-x-exp-with-std` once the issue existed, before any push.
- `tokens/token.go`:
  - The pinned `x/exp/maps.Values` returns `[]V`, but the standard `maps.Values` returns `iter.Seq[V]`. The membership check therefore uses `slices.Collect(maps.Values(keywords))`.
  - Order is irrelevant to `slices.Contains`, so behaviour is unchanged.
  - The pre-existing allocation on every `String()` call is untouched; out of scope.
- `api/net.go`: import swap from `golang.org/x/exp/slices` to `slices`, with the same `Contains` signature.
- `go mod tidy` removed `golang.org/x/exp` from `go.mod` and `go.sum`. Direct requirements went from 14 to 13.
- Rewrote both touched files' legacy Apache headers to the two-line SPDX form for 2026.
- Graph: `tokens` and `api.net` now import `std.maps` and `std.slices` instead of `ext.golang.x-exp`. `ext.dependencies` drops the two `x-exp` rows and the "removable" note, and changes "14 direct" to "13 direct" in both places.
- Verified, with `GOWORK=off`: `go build ./...`, `go vet` on `tokens` and `api`, `go test ./...` (19 packages ok) and `make docs-graph` all pass.
