## Binaek Sarkar (binaek89@gmail.com)

### 14:46 - Why this change
- While picking up the twelve JS backlog issues (#102, #105, #111, #138, #139, #142, #153, #154, #156, #158, #168, #191), the agent cut branches and spawned implementers without first pushing a stub-journal commit and opening each draft PR.
- It read [Batch origin access](../.agentrules/origin-access.md) as "push once when everything is done" and told every subagent to make no origin touches. The issue workflow instead puts the journal stub, commit, push and draft PR (steps 4 to 7) before any implementation.
- The rules don't contradict each other, but nothing says batching must not delay the pickup push, and nothing covers picking up several issues at once or handing them to subagents. This PR closes those gaps.
- _Model: claude-opus-5-5_
