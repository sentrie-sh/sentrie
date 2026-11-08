## Binaek Sarkar (binaek89@gmail.com)

### 18:23 - Pin website PRs to the release branch
- Asked to make sure every `sentrie-sh/website` PR targets `v0.0.5`. All 22 open feature PRs already did; the only PR on `main` is website #36, the `v0.0.5` into `main` release PR, which must stay there. Nothing was retargeted.
- No rule recorded the convention, and `gh pr create` defaults to `main`, so a future website PR could silently land on the wrong base. Added `.agentrules/website-pr-base.md` and indexed it under "Merging" in `AGENTS.md`.
- The rule names `v0.0.5` explicitly rather than "the current release branch" so agents have a concrete `--base` value, and says to update the name in an `[AgentRule]` PR when the next release branch is cut.
- Placed in this repo rather than the website's `AGENTS.md`, which defers to this one for guidance.
- _Model: claude-opus-5-5_

### 21:25 - Target a version-free next branch instead of v0.0.5
- Supersedes the 18:23 entry's choice to name `v0.0.5` in the rule: a version-named base forces an `[AgentRule]` PR every release. The rule now targets `next`, the next-publish branch, and names no version.
- On the website repo, renamed `v0.0.5` to `next` through the branch rename API (chosen over creating `next` alongside). GitHub retargeted all 21 open docs PRs (#21, #49 to #71 except #67) to `next` within seconds.
- The rename closed release PR website #36: GitHub retargets PRs whose base is renamed, but a PR whose head is renamed is treated as having its head deleted. Its body was empty and it had no labels or reviewers. The replacement `next` into `main` PR could not be opened from this session (permission denied) and was left to the user.
- Found the website repo is squash-only with `delete_branch_on_merge` on, so merging the release PR would delete `next` and push the open docs PRs onto `main`. The rule now says to protect `next` or restore it and retarget before merging the release PR.
- _Model: claude-opus-5-5_

### 21:28 - Replace the release PR
- Opened website #73 (`next` into `main`, title "V0.0.5") to replace #36 on the user's go-ahead.
- _Model: claude-opus-5-5_
