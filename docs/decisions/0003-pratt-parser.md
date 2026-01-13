# ADR-0003: Pratt Parser for Language Parsing

**Status:** Accepted  
**Date:** 2025-09-14  
**Deciders:** [@binaek](https://github.com/binaek)  
**Tags:** language-design, parser, implementation

## Context

Sentrie needs a parser to convert source code into an Abstract Syntax Tree (AST). The parser must:

- Handle operator precedence correctly
- Be extensible for new operators and syntax
- Provide clear error messages
- Be maintainable and readable
- Support the language's expression syntax

Several parsing approaches exist: recursive descent, generated parsers (YACC/Bison), parser combinators, and Pratt parsers.

## Decision

Sentrie uses a **Pratt parser** (also known as a **precedence climbing parser**) implemented manually in Go. The parser uses prefix and infix handler maps to register parsing functions for different token types.

## Rationale

1. **Operator precedence handling**: Pratt parsers naturally handle operator precedence through precedence levels, making it easy to add new operators with correct precedence.

2. **Extensibility**: Adding new operators or syntax constructs is straightforward - just register new prefix or infix handlers. No grammar file regeneration needed.

3. **Error handling**: Manual implementation allows for precise, context-aware error messages at parse time.

4. **No external dependencies**: A manual Pratt parser doesn't require parser generator tools or external dependencies, keeping the build process simple.

5. **Readability**: The parser code is straightforward to read and understand, making maintenance easier.

6. **Performance**: Pratt parsers are efficient, typically requiring a single pass through the tokens.

7. **Flexibility**: Easy to handle context-sensitive parsing (e.g., different parsing rules inside policies vs. at top level).

## Consequences

### Positive

- Easy to add new operators with correct precedence
- Clear, maintainable parser code
- Good error messages with precise locations
- No build-time code generation step
- Single-pass parsing is efficient
- Context-sensitive parsing is straightforward

### Negative

- Manual implementation requires more code than generated parsers
- Precedence levels must be carefully managed
- More code to maintain than using a parser generator
- Requires understanding of Pratt parsing algorithm
- Limited support for backtracking

### Neutral

- Parser is specific to Sentrie's syntax
- Changes to syntax require code changes (not just grammar file)

## Alternatives Considered

### Alternative A: Recursive Descent Parser

**Description:** Manually implement a recursive descent parser with functions for each grammar production.

**Pros:**

- Simple and intuitive
- Easy to understand
- Good error handling control
- No external dependencies

**Cons:**

- Operator precedence requires careful function call ordering
- Can become verbose for complex expressions
- Harder to add new operators without refactoring
- Left-recursion requires transformation

**Why not chosen:** Pratt parser provides better handling of operator precedence and is more extensible for adding new operators.

### Alternative B: Parser Generator (YACC/Bison, ANTLR, etc.)

**Description:** Use a parser generator tool with a grammar file.

**Pros:**

- Declarative grammar specification
- Automatic parser generation
- Well-established tools
- Can handle complex grammars

**Cons:**

- Requires build-time code generation
- Less control over error messages
- External tool dependency
- Generated code can be harder to debug
- Grammar files can become complex
- Less flexible for context-sensitive parsing

**Why not chosen:** The manual Pratt parser provides better control over error messages and doesn't require build-time code generation. The language syntax is manageable without a generator.

### Alternative C: Parser Combinators

**Description:** Use a parser combinator library (like in Haskell, Rust, etc.).

**Pros:**

- Very expressive
- Composable parsing logic
- Good for complex grammars

**Cons:**

- Requires a parser combinator library (external dependency)
- Less common in Go ecosystem
- Can have performance overhead
- Steeper learning curve
- Error messages can be less precise

**Why not chosen:** Go doesn't have a mature parser combinator ecosystem, and the manual Pratt parser provides sufficient expressiveness with better control.

### Alternative D: PEG Parser (Pigeon, etc.)

**Description:** Use a PEG (Parsing Expression Grammar) parser generator.

**Pros:**

- Handles ambiguity automatically
- Good error recovery
- Declarative grammar

**Cons:**

- Requires code generation
- Less control over parsing behavior
- Can have performance issues with backtracking
- Error messages may be less precise

**Why not chosen:** The manual Pratt parser provides better control and doesn't require code generation. PEG's backtracking can be inefficient for large files.

## Implementation Notes

- The parser is implemented in `parser/parser.go`
- Prefix and infix handlers are registered in `registerParseFns()`
- Precedence levels are defined in `parser/precedence.go`
- The parser maintains current and next tokens for lookahead
- Error messages include token positions for debugging
- Context-sensitive parsing is handled through separate statement handler maps (e.g., `policyStatementHandlers`)

## References

- Implementation: `parser/parser.go`, `parser/precedence.go`
- Related: ADR-0001 (The Sentrie Language)
- [Pratt Parsers](https://en.wikipedia.org/wiki/Operator-precedence_parser#Pratt_parsing)
- [Vaughn Pratt's original paper](https://dl.acm.org/doi/10.1145/512927.512931)
