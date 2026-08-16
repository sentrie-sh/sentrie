---
id: parser.precedence
type: Module / File
language: Go
file_path: parser/precedence.go
tags: operator-precedence, pratt-parser, language-semantics, binding-power
---

# Node: parser.precedence (Operator Binding Powers)

## 1. Architectural Role & Intent
`parser/precedence.go` declares the `Precedence` ladder and the `precedences` table that maps each infix-capable token kind to its binding power. It is the authoritative answer to "how does `a or b and c |> f(#)` group?" — the Pratt loop in [[parser.expression]] consults nothing else when deciding whether to keep extending an expression. Keeping it in one small file makes operator precedence auditable without reading any production.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| `parser.precedence` | `DEPENDS_ON` | [[tokens]] | The table is keyed by `tokens.Kind`. |
| [[parser.expression]] | `READS_FROM` | [[parser.precedence]] | The Pratt loop condition is `precedences[current.Kind] > precedence`. |
| [[parser.lookups]] | `DEPENDS_ON` | [[parser.precedence]] | The `infixParser` signature carries a `Precedence` argument. |
| [[grammar]] | `DEPENDS_ON` | [[parser.precedence]] | The EBNF/PEG precedence chain is the documentation counterpart of this table; this table is what actually runs. |

## 3. Interface Contracts & Public Surface

- **Signature:** `Precedence uint8` with constants, lowest to highest — `LOWEST`, `PIPELINE` (`|>`), `TERNARY` (`? :`), `OR`, `XOR`, `AND`, `EQUALITY` (`==`, `!=`, `is`), `COMPARISON` (`<`, `>`, `<=`, `>=`, `matches`, `contains`, `in`), `SUM` (`+`, `-`), `PRODUCT` (`*`, `/`, `%`), `UNARY` (`!x`, `-x`, `+x`, `not`, `as`), `CALL` (`(`), `INDEX` (`[`, `.`), `PRIMARY`
  - **Behavior:** An `iota` ladder; higher values bind tighter. `LOWEST` is the entry precedence used when parsing a fresh expression, and `PRIMARY` is the ceiling for primary expressions.
  - **Side Effects:** None.
  - **Exceptions:** None.

- **Signature:** `precedences map[tokens.Kind]Precedence` (package-level var)
  - **Behavior:** Binding power per operator token. Any token kind **absent** from this map yields the zero value `LOWEST`, which terminates the Pratt loop — this is the mechanism by which statement keywords, `)`, `}`, `,`, `;`, and EOF naturally end an expression without an explicit terminator list.
  - **Side Effects:** None.
  - **Exceptions:** None.

## 4. Operational Context & Gotchas
- **Statefulness:** Package-level immutable-by-convention table; read concurrently by any number of parsers.
- **Performance/Scale Notes:** One map probe per token in the Pratt loop. Trivial.
- **Dependencies Risk:** No external failure domain, but several semantics worth knowing before changing anything here:
  - **`|>` binds loosest of all real operators** (just above `LOWEST`), so a pipeline stage swallows the entire expression to its right up to the next `|>` — `x |> f(#) or y` parses the `or` *inside* the stage, not around the pipeline.
  - **Boolean precedence is `or` < `xor` < `and`**, matching the [[trinary]] Kleene algebra's conventional grouping.
  - **`is` sits at `EQUALITY`, not `COMPARISON`**, so `a is defined == b` groups differently than the neighbouring keyword operators.
  - **`as` (cast) is registered at `UNARY`**, higher than every arithmetic and comparison operator, so `a + b as string` casts only `b`. This is easy to misread as a low-precedence trailing cast.
  - **`.` and `[` share `INDEX`, above `CALL`**, which makes `f(x).y` and chained access parse as expected but means an index/field access binds tighter than a call on the same chain head.
  - **The map is the only source of truth.** Registering an infix handler in [[parser.lookups]] without adding a precedence entry silently makes the operator unreachable: its binding power defaults to `LOWEST`, so the loop never enters the handler.
