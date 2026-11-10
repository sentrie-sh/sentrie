## Binaek Sarkar (binaek89@gmail.com)

### 2026-11-10 18:10 - Pick up #157
- The pack's `[engine]` version constraint is never checked.
- Picked up in batch 3 of the v0.0.5 milestone (#115, #116, #157, #316), locally while SSH signing is down: this stub, its push and the draft PR are queued in the orchestration state file. Paired `sentrie-sh/website` branch of the same name, cut from `next`.
- Decisions recorded with the user on 2026-11-10, which the reviewer judges the change against:
  - Escape-hatch flag `--skip-engine-check` on `exec`, `serve` and `validate`.
  - A dev or unversioned build passes with a warning.
  - Prereleases use STRICT semver: `0.0.5-rc1` fails `>=0.0.5` (user overrode the recommendation).
  - A mismatch under `validate` is an error.
- Branch cut from `main`.
- Docs PR: sentrie-sh/website#90 (draft, base `next`); Sentrie PR #330.
- _Model: claude-opus-5-5_
