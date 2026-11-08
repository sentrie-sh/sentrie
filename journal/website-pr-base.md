## Binaek Sarkar (binaek89@gmail.com)

### 18:23 - Pin website PRs to the release branch
- Asked to make sure every `sentrie-sh/website` PR targets `v0.0.5`. All 22 open feature PRs already did; the only PR on `main` is website #36, the `v0.0.5` into `main` release PR, which must stay there. Nothing was retargeted.
- No rule recorded the convention, and `gh pr create` defaults to `main`, so a future website PR could silently land on the wrong base. Added `.agentrules/website-pr-base.md` and indexed it under "Merging" in `AGENTS.md`.
- The rule names `v0.0.5` explicitly rather than "the current release branch" so agents have a concrete `--base` value, and says to update the name in an `[AgentRule]` PR when the next release branch is cut.
- Placed in this repo rather than the website's `AGENTS.md`, which defers to this one for guidance.
- _Model: claude-opus-5-5_
