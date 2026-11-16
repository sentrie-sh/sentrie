# 333-example-pack-engine

## Binaek Sarkar (binaek89@gmail.com)

### 2026-11-17 00:05 - Pick up #333
- Found during the v0.0.5 pre-release sanity pass on `main` at `47a77ac`. `make test`, `make lint` and the race runs were all green, including `./runtime/...` with `-race`, which CI skips.
- A binary built with `-X main.version=0.0.5` refuses `example_pack`: `pack requires sentrie >=0.1.0 <2.0.0, running 0.0.5`. The check came in with #330 (Closes #157). Dev builds only warn, so CI never sees this.
- `sentrie init` writes no `[engine]` block, so new packs are unaffected.
- Scope: the example pack's manifest only. The website `next` docs show the same `0.1.0` constraint in `packfile.md`, `cli-reference/index.md` and `cli-reference/serve.md`. That fix belongs in `sentrie-sh/website`.
- Test fixtures `loader/pack_test.go` and `cmd/serve_signal_test.go` also use `>=0.1.0 <2.0.0`. They are left alone: the tests never compare against a release-stamped version.
- Branch and worktree `worktrees/333-example-pack-engine` off local `main`; `gh api` confirmed local `main` matches origin.
- _Model: claude-opus-5-5_

### 2026-11-17 00:33 - Unblock signing, then fix the constraint
- The first stub commit hung in `ssh-keygen -Y sign`, both in the session and in the user's own terminal. The Secretive agent itself answered `ssh-add -l` and `ssh-add -L` fine.
  - Cause: `user.signingkey` pointed into `~/Library/Containers/com.maxgoedjen.Secretive.SecretAgent/Data/PublicKeys/`. Any read of that folder blocked, even `ls`, so `ssh-keygen` never reached the agent. Likely macOS app-data protection on another app's container. Restarting SecretAgent did not help.
  - Fix, made by the user in their global git config: export the signing key's public half from the agent (matched by fingerprint against `~/.gitallowedsigners`) to `~/.ssh/git-signing.pub`, and point `user.signingkey` there. The private key stays in Secretive.
  - Signing never bypassed. The stub commit `3321231` is signed with `SHA256:8Tj4JVXo...`.
- Pushed the branch and opened draft PR #334, linked to #333 under Development.
- Changed `example_pack/sentrie.pack.toml` `[engine] sentrie` from `>=0.1.0 <2.0.0` to `>=0.0.5 <1.0.0`.
  - Chose a range over a bare `>=0.0.5` so the example still shows a two-sided constraint, like the docs do.
  - Upper bound `<1.0.0` because the pack is maintained in-repo and can be bumped with the engine; it signals that 1.0 may break the language.
- Verified with binaries stamped by `-X main.version`: 0.0.4 is refused with the expected error; 0.0.5 and 0.0.6 validate with the same four warnings as before (two unenforced permissions, two nondeterminism notices).
- `go test -count=1` on every package that loads `example_pack` (`cmd`, `format`, `api`, `index`, `runtime`, `lexer`, `parser`) passes.
- _Model: claude-opus-5-5_
