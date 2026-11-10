## Binaek Sarkar (binaek89@gmail.com)

### 2026-11-10 18:10 - Pick up #115
- The decision endpoint reads request bodies without a size limit. Scope is the body size limit only; authentication, CORS, TLS, logging and traces moved to #315 (out of scope; the issue was rescoped to the limit), so the PR says `Closes #115`.
- Picked up in batch 3 of the v0.0.5 milestone (#115, #116, #157, #316), locally while SSH signing is down: this stub, its push and the draft PR are queued in the orchestration state file. Paired `sentrie-sh/website` branch of the same name, cut from `next`.
- Decisions recorded with the user on 2026-11-10, which the reviewer judges the change against:
  - Wrap the body in `http.MaxBytesReader`. Default limit 1 MiB, overridable on `serve` with `--http-max-body-bytes` (plain integer bytes); 0 means unlimited (user overrode the recommendation).
  - A body over the limit returns 413 with a problem document.
- Branch cut from `main`.
- Docs PR: recorded once opened.
- _Model: claude-opus-5-5_
