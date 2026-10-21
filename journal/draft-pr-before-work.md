## Binaek Sarkar (binaek89@gmail.com)

### 14:46 - Why this change
- While picking up the twelve JS backlog issues (#102, #105, #111, #138, #139, #142, #153, #154, #156, #158, #168, #191), the agent cut branches and spawned implementers without first pushing a stub-journal commit and opening each draft PR.
- It read [Batch origin access](../.agentrules/origin-access.md) as "push once when everything is done" and told every subagent to make no origin touches. The issue workflow instead puts the journal stub, commit, push and draft PR (steps 4 to 7) before any implementation.
- The rules don't contradict each other, but nothing says batching must not delay the pickup push, and nothing covers picking up several issues at once or handing them to subagents. This PR closes those gaps.
- _Model: claude-opus-5-5_

### 14:49 - Close the gaps
- `origin-access.md`: the phrase "one push at the end" was the direct source of the misreading. It now says "one push per batch of branches that are due", and a new bullet states that batching never moves a push a workflow fixes in time, naming "push once at the end" as the misreading. The "do the local work first" bullet now spells out that, for a pickup, local work means branch, journal, PR description and commit, not the implementation.
- `issue-workflow.md`: states that all seven steps precede implementation and that later implementation pushes never replace step 7. New "Several issues at once" section: steps 3 to 6 per issue, one shared push and one draft PR per branch, then implement. When work fans out to subagents, the parent owns origin and subagents report each new branch after committing its stub journal.
- `parallelize-with-subagents.md`: one "Origin stays with the parent" bullet under prompt tightening, pointing back to the workflow section rather than restating it.
- `pr-conventions.md`: records the `addCloseIssueReferences` input shape (`issueId` plus a list of `pullRequestIds`). The first attempt in this session guessed the reverse shape and failed; it is a one-line addition in the same family of "rule text that leads agents astray".
- `AGENTS.md`: index descriptions and keywords updated so a search for "several issues", "fan-out" or "push at the end" lands on the right rules.
- Kept the change to clarifications. No new rule file: the convention already existed, it was just easy to misread.
- _Model: claude-opus-5-5_
