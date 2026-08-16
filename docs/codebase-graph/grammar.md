---
id: grammar
type: System / Package
language: EBNF, PEG
file_path: grammar/
tags: language-specification, documentation, reference-artifact, syntax
---

# Node: Grammar (Reference Language Specification)

## 1. Architectural Role & Intent
`grammar` holds the normative written specification of Sentrie's surface syntax in two notations — `grammar.ebnf` and `grammar.peg` — covering program structure, declarations (`namespace`, `policy`, `fact`, `use`, `let`, `rule`, `shape`, `derive`, exports), the full expression precedence ladder, and type references with constraints. It exists as the human- and agent-readable contract that [[lexer]] and [[parser]] are expected to implement. It is **documentation, not build input**: no Go file, `Makefile` target, or CI job reads these files, and no parser is generated from them.

## 2. Graph Edges (Strict Relational Data)

| Source (Subject) | Relationship (Predicate) | Target (Object) | Context / Data Payload Flow |
| :--- | :--- | :--- | :--- |
| [[parser]] | `DEPENDS_ON` | [[grammar]] | Conformance-only dependency: the recursive-descent parser is a hand-written implementation of these production rules. No code or generated artifact links them. |
| [[lexer]] | `DEPENDS_ON` | [[grammar]] | The token vocabulary realises the terminals (`IDENT`, `STRING`, `INT`, `FLOAT`, keywords, operators) declared here. |
| [[ast]] | `DEPENDS_ON` | [[grammar]] | Each non-terminal has a corresponding AST node type; the grammar is the map from syntax to node kinds. |

## 3. Interface Contracts & Public Surface

This node exposes no callable surface. Its "interface" is the set of productions that define the language:

- **Signature:** `program ::= (comment)* namespaceDecl (toplevelDecl | comment)*`
  - **Behavior:** Every file must declare exactly one namespace before any top-level declaration; comments are first-class and may appear before it.
  - **Side Effects:** N/A.
  - **Exceptions:** N/A.

- **Signature:** `toplevelDecl ::= policyDecl | shapeDecl | exportShape | deriveDecl | exportDerive`
  - **Behavior:** Fixes the four namespace-level declaration forms. `derive` and `shape` exist at both namespace and policy scope; visibility is opt-in via `export`.
  - **Side Effects:** N/A.
  - **Exceptions:** N/A.

- **Signature:** `ruleDecl ::= 'rule' IDENT '=' ('default' expr)? ('when' expr)? (blockExpr | ruleImportClause)`
  - **Behavior:** The core policy unit: an optional default outcome, an optional applicability guard, and either an inline block body or a cross-namespace decision import.
  - **Side Effects:** N/A.
  - **Exceptions:** N/A.

- **Signature:** `expr ::= ternaryExpr` → `orExpr` → `xorExpr` → `andExpr` → `unaryExpr` → `cmpExpr` → `addExpr` → …
  - **Behavior:** The authoritative operator precedence chain, including the elvis form `a ? : b` and the pipeline operator `|>` with its `#` hole.
  - **Side Effects:** N/A.
  - **Exceptions:** N/A.

- **Signature:** `typeRef` with `constraint` suffixes
  - **Behavior:** Declares that type references carry named, argument-bearing constraints — the syntax counterpart to the checker tables in [[constraints]].
  - **Side Effects:** N/A.
  - **Exceptions:** N/A.

## 4. Operational Context & Gotchas
- **Statefulness:** Stateless static artifacts. Two text files, no runtime presence, no embedding.
- **Performance/Scale Notes:** Not applicable — nothing loads these at build or run time.
- **Dependencies Risk:** The single significant hazard is **specification drift**. Because the parser is hand-written and nothing validates it against these files, a grammar change is purely advisory and a parser change can silently diverge from the spec. Treat these files as a *starting hypothesis* when answering "what syntax is legal?" and confirm against [[parser]] before relying on them; where the two disagree, the parser is what actually runs. The `.peg` and `.ebnf` variants are maintained in parallel and can also drift from each other (e.g. `varDecl` carries an optional type annotation in the PEG version but not the EBNF version).
