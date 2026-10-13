# Journal Convention

Maintain a running journal at `journal/<branchname>.md` for the branch being worked on. It is committed with the branch.

## Why journals exist

Journals are institutional knowledge, not a passive log. The entries capture how we like to approach and solve problems in this repo: the conventions, methodologies, decisions, tradeoffs, and fixes that accumulate over time. Read them as a source of team knowledge to learn from.

They also record past context that may since have gone stale, or that turned out not to be a good fit. Adapt your approach to the problem in front of you rather than following past entries blindly, and where a journal entry conflicts with a current rule in `.agentrules/`, the current rule wins.

## Format

- **Append-only and strictly chronological.** Add new entries at the bottom. Never insert into the middle, and never rewrite or reorder past entries. (The one exception is merge conflict resolution, below.)
- **Entry heading:** a timestamped `###` heading, for example `### 15:13 - <short title>`, followed by bullets. Prefix the time with the date when the branch's work spans more than one day.
- **Author grouping:** group entries under a `##` header identifying whoever made the changes, taken from the local git identity (`git config user.name` and `git config user.email`), formatted as `## <git user name> (<git email>)`. Place it above that author's timestamped entries.
- **Do not repeat the author header** for successive entries by the same author.

```
## Jane Doe (jane.doe@example.com)

### 15:13 - <short title>
- ...
- ...
- _Model: claude-opus-5-5, tokens: ~18k_

### 15:41 - <short title>
- ...
- ...
- _Model: claude-opus-5-5, tokens: ~7k_
```

The trailing model and token bullet is optional: include it when the numbers are available, and leave it off otherwise. See [What to record](#what-to-record).

## What to record

- Log every nitty-gritty step: what was done, decisions made and why, problems faced, and how they were solved. This is a detailed running journal, not a summary written after the fact.
- Write entries as you go, incrementally. Never save it all up for the end.
- **Capture manual edits too.** Work done by hand, outside the current session, still belongs in the journal. Infer what changed from the diff, the commit history, and the surrounding context, and record it.
- **Ask when a manual change cannot be inferred.** Do not guess at intent from ambiguous evidence. Accuracy matters more than the inconvenience of asking.
- **Record model and token usage when available.** For entries covering work done by an AI agent, close the entry with a final italic bullet naming the model and the approximate tokens consumed for that stretch of work. Omit the bullet entirely for human-authored entries, or when the running tool does not expose the numbers. Never estimate a token count you do not actually have: an absent figure is fine, an invented one corrupts the record.

## Superseding earlier work

When a change in the current branch corrects, reverses, or supersedes something done in an older branch or PR, call it out explicitly in the entry. State what was wrong and how this change fixes it, and reference the older branch or PR.

The point is to avoid two conflicting sources of truth. A future reader must be able to tell which decision is no longer authoritative, so they don't follow the outdated approach. This cross-referencing is how the accumulated journal knowledge stays internally consistent as decisions evolve.

## Multi-author branches

When several people work in the same branch, resolve journal merge and rebase conflicts so that entries stay chronological across all authors. Interleave and reorder strictly by timestamp.

Chronological order is the invariant; append-only is simply how it is maintained during normal work. Reordering during a merge serves that invariant rather than breaking it.

Because ordering is by timestamp rather than by author, an author header repeats whenever authorship alternates. Consecutive entries by one author share a single header; when another author's entry falls between two of theirs, the header appears again after it:

```
## Jane Doe (jane.doe@example.com)

### 15:13 - <short title>
- ...

## John Roe (john.roe@example.com)

### 15:20 - <short title>
- ...

## Jane Doe (jane.doe@example.com)

### 15:41 - <short title>
- ...
```

Related: [Issue workflow](issue-workflow.md), [PR description file](pr-description-file.md)
