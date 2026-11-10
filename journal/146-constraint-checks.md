## Binaek Sarkar (binaek89@gmail.com)

### 2026-11-10 15:42 - Pick up #146
- `@min(-5)` fails to parse, `@in` cannot be written, `@url` rejects ports and recompiles its regex per call, `@length` counts bytes (constraints/string.go:29), and the parse error is garbled (`%!s(int=3)`).
- Picked up in batch 1 of the v0.0.5 milestone (#307/#308, #149, #143, #146, #141, #159): one pickup push and a draft PR per branch, plus a paired `sentrie-sh/website` branch of the same name cut from `next`, before implementation.
- Decisions recorded with the user on 2026-11-10, which the reviewer judges the change against:
  - `@in` and `@not_in` take variadic arguments only (`@in(1, 2)`), not a list (the user chose this over accepting both).
  - On a string, `@in("a", "b")` means the value equals one of the listed strings.
  - Any keyword is accepted as a constraint name after `@`; the checker table decides whether it exists.
  - `@url` parses with net/url and needs a scheme and a host; ports, IPs and localhost are fine; with no arguments only http and https pass, and variadic scheme arguments widen it (`@url("ftp", "s3")`).
  - `@length` counts Unicode code points.
  - Constraint arguments accept unary minus on numeric literals only.
- Docs PR: recorded once opened.
- _Model: claude-opus-5-5, subagents: 1x sonnet (design questions)_
