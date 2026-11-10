## Binaek Sarkar (binaek89@gmail.com)

### 2026-11-10 18:10 - Pick up #116
- IPv6 predefined listen addresses `local6` and `network6` produce `[[::1]]` (api/net.go:31,37), and an empty listen list panics (api/net.go:26).
- Picked up in batch 3 of the v0.0.5 milestone (#115, #116, #157, #316), locally while SSH signing is down: this stub, its push and the draft PR are queued in the orchestration state file. Paired `sentrie-sh/website` branch of the same name, cut from `next`.
- Decisions recorded with the user on 2026-11-10, which the reviewer judges the change against:
  - No design decision needed: use the issue's fix (bare `::1` and `::` so host:port joining brackets them once; an empty listen list is an error, not a panic).
- Branch cut from `main`.
- Docs PR: sentrie-sh/website#89 (draft, base `next`); Sentrie PR #329.
- _Model: claude-opus-5-5_
