## Binaek Sarkar (binaek89@gmail.com)

### 2026-11-10 22:54 - Pick up #145
- Shapes accept fields they don't declare, `dict[T]` values are unchecked, a shape can declare the same field twice, and duplicate dict literal keys overwrite each other silently.
- Picked up in batch 2 of the v0.0.5 milestone (#148, #145, #144, #155, #147), locally while the user is AFK: SSH signing is down, so this stub and its push are queued in the orchestration state file. Paired `sentrie-sh/website` branch of the same name, cut from `next`.
- Decisions recorded with the user on 2026-11-10, which the reviewer judges the change against:
  - Shapes are closed by default, nested shapes too. The opt-out is a `...` marker as the last entry inside the shape body (user picked this over `@open`).
  - `dict[T]` values are checked against T, including constraints, and the error names the key (check whether #299, already merged, covers this).
  - A duplicate shape field is an error that shows both spans.
  - A duplicate dict literal key keeps last-one-wins, and the docs say so (user overrode the recommendation).
- PROVISIONAL (user AFK), needs confirmation: branch stacked on `159-keyword-field-names`, which changes shape field parsing (parser/shape.go). PR base `159-keyword-field-names` until #159 merges.
- Docs PR: recorded once opened.
- _Model: claude-opus-5-5_
