# Implement, Test and Review Loop

Every issue is carried by three separate agents: an implementer, a test and CI agent, and a reviewer. Their findings loop back to the implementer until the reviewer approves and the local CI gate passes. When several issues are picked up together, the finished branches are also checked against each other before anything merges.

**Why:** An author is a poor judge of their own change. A separate test agent proves the tests bite, and a separate reviewer catches missed call sites and fail-open paths the author already reasoned past. Branches that each pass alone can still break each other, and that is cheapest to find before the first one merges.

## Before implementation

- Pick up every issue first, with its draft PR opened before any code is written ([Issue workflow](issue-workflow.md), "Several issues at once").
- When an issue leaves a design choice open, ask the user before spawning its implementer. Offer the options with a recommendation, and record the answer in the pickup journal entry. The reviewer judges the change against that recorded decision.
- Write the shared briefs once, in the session's scratch directory, and point every agent at them instead of pasting them into each prompt ([Parallelize with subagents](parallelize-with-subagents.md), "Tighten the prompt"):
  - a common brief: worktree path, `GOWORK=off`, no origin access, headers, em dashes, signatures, commit style, journal and PR description rules, test output discipline;
  - a test and CI brief;
  - a review brief.

## The three roles

Each role is its own agent, with `model` chosen from [Parallelize with subagents](parallelize-with-subagents.md). The reviewer is never the author, and is at least the author's tier.

- **Implementer:** makes the change, updates the graph nodes it makes false ([Codebase knowledge graph](codebase-knowledge-graph.md)), and keeps the journal and the PR description current.
- **Test and CI agent:**
  - Writes the regression tests the issue and the recorded decision call for, under `.cursor/agents/test-writer.md`.
  - Never changes non-test code. A test that exposes a bug is left out of the commit, and the bug is reported with the input, the expected result and the actual one.
  - Proves each test bites by running it against `main`'s copy of the production file (`git checkout main -- <file>`, run, `git checkout HEAD -- <file>`). Reasoning that it would fail is not enough.
  - Runs the local CI gate below and reports pass or fail per check, with log paths.
- **Reviewer:**
  - Read-only, under `.cursor/agents/code-reviewer.md`. It may run tests or throwaway programs to confirm a suspected bug before reporting it.
  - Returns `VERDICT: APPROVE` or `VERDICT: CHANGES` on its first line, then Critical, Warnings and Suggestions, each with `file:line` and a concrete fix.
  - Approves only when no Critical or Warning finding is left.

## The loop

1. Implementer, then the test and CI agent, then the reviewer.
2. Send every Critical and Warning finding, and every bug the test agent reported, back to the same implementer. Continue that agent so it keeps its context; do not spawn a fresh one.
3. Rerun the test and CI agent on the fix, then send the fix back to the same reviewer.
4. Repeat until the reviewer approves and the gate passes.
5. A finding still open after two rounds on the same fix goes to the user with the diagnosis, per the two-attempt limit in [Token efficiency](token-efficiency.md).

Suggestions may be taken or declined. Record a declined suggestion in the journal with the reason.

## Local CI gate

Run each check with its output redirected to a log, then grep the log ([Token efficiency](token-efficiency.md)):

- `make test`, then `rg -n '^(FAIL|---|panic)'` on the log;
- `make lint` ([Build and test](build-and-test.md) covers a local golangci-lint that panics on a newer standard library);
- `gofmt -l` over the changed Go files, which must print nothing;
- `make docs-graph`;
- `go test -race` on the non-runtime packages the branch touches, as CI does;
- no em dash in the changed files ([No em dashes](no-em-dash.md)), and no AI signature in `git log main..HEAD --format=%B` ([No agentic signatures](no-agentic-signatures.md)).

## Check the branches together

When two or more branches in the batch touch the same packages, merge them all in a throwaway clone before any of them leaves draft:

- Clone into the session's scratch directory, never into `worktrees/`, and never merge into the real branches.
- Merge every branch, run the local CI gate on the result, and exercise the combined `sentrie` binary on the behaviour the issues describe.
- Commits in that clone are scaffolding that is deleted and never pushed, so they may skip signing there. Nowhere else.
- Fix only what breaks in combination, such as a test expecting a position or a value another branch changed, or a graph node another branch made stale. Do not fold those fixes into the branches: each PR describes only its own diff ([PR description file](pr-description-file.md)).
- Save the conflict resolutions and fixes as a recipe and patches for whichever branch of each pair merges second. Note the merge order in the Review notes of each affected PR.

## When commits or pushes must wait

- If signing fails, retry once. If it fails again, never bypass it with `--no-gpg-sign` or `-c commit.gpgsign=false`. Tell the user.
- If signing is down, or the user says to hold commits or pushes, keep working and queue the work instead. Write each commit to a script exactly: the paths it stages and its subject.
- When the user clears it:
  - check that every worktree's staged and unstaged state still matches the script;
  - run the script;
  - confirm every branch is clean, builds, and has a signature on every new commit.
- Then push every branch in one `git push` ([Batch origin access](origin-access.md)), and refresh each PR's title and body from its PR description file with `gh pr edit`.
- PRs stay drafts. Marking one ready starts CI, and that is the user's call.

Related: [Issue workflow](issue-workflow.md), [Parallelize with subagents](parallelize-with-subagents.md), [Batch origin access](origin-access.md), [Token efficiency](token-efficiency.md), [Build and test](build-and-test.md), [Commits](commits.md), [PR description file](pr-description-file.md), [Codebase knowledge graph](codebase-knowledge-graph.md)
