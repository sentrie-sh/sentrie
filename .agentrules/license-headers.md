# License Headers and Copyright Year

Make sure every source file has the correct license header. Maintain headers on the files you touch, and only those: add a missing header, and bring an outdated one up to date.

## Source files

In this repository, source files are those with these extensions: `.go`, `.ts`, `.js`, `.sentrie`, `.sh`, `.ps1`, `.peg`, `.ebnf`. Follow repository conventions if more extensions are treated as source. Don't apply this rule to non-source files (for example `.md`, `.txt`, `.json`, `.yaml`, `.yml`, lockfiles, generated assets) unless asked.

## The header

New files, and any header you rewrite, use this two-line SPDX shape in the file's comment syntax:

```go
// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0
```

- **Year:** the current calendar year for the session. Use today's date from the workspace context when it is given. Otherwise use a reasonable "now", never a fixed past year.
- **Holder and email:** from `git config user.name` and `git config user.email`. If `user.email` is unset, omit the bracketed email. If git identity is unavailable, or the repo standard differs (for example a corporate copyright line), copy the holder style from sibling source files.
- **License:** `SPDX-License-Identifier` is always `Apache-2.0`.

## Touched files

- **Outdated year:** a touched source file whose header has an older year gets the current year.
- **Legacy header:** a touched source file with the older header (`SPDX-License-Identifier` plus a `Copyright ...` line and the Apache boilerplate) is rewritten to the two-line SPDX form with the current year.
- **Untouched files:** never change a file only to refresh its header, and never run a repository-wide header refresh unless explicitly asked.

**Why:** Accurate headers keep the licence clear for downstream users. Updating only touched files keeps diffs focused on the change being reviewed.

**How to apply:** Before committing, check the header of every source file in `git diff --name-only`.

Related: [Commits](commits.md)
