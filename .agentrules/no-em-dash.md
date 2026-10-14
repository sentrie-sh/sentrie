# No Em Dashes

Never use an em dash (Unicode U+2014) in any text. Use a hyphen (-) instead, or restructure the sentence.

**Why:** Keeps written output in one plain house style.

**How to apply:** Applies everywhere text is produced: chat responses, code comments, commit messages, PR titles and descriptions, journal entries, documentation, and rule files. Before finalizing any written output, check for and remove em dashes, for example with `rg -n '\x{2014}' <paths>`.

Related: [No agentic signatures](no-agentic-signatures.md)
