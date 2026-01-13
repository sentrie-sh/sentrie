# Sentrie Decision Ledger

This directory contains Architecture Decision Records (ADRs) documenting significant architectural, design, and strategic decisions made in the Sentrie project.

## What is an ADR?

An Architecture Decision Record is a document that captures an important architectural decision made along with its context and consequences. ADRs help:

- **Preserve knowledge**: Understand why decisions were made, not just what was decided
- **Onboard new contributors**: Quickly understand the rationale behind design choices
- **Avoid repeating discussions**: Reference past decisions when similar questions arise
- **Track evolution**: See how decisions have changed over time

## ADR Index

### Language Design

- [ADR-0001: The Sentrie Language](0001-sentrie-language.md) - Overview of the Sentrie domain-specific language
- [ADR-0002: Non-Turing Complete Language Design](0002-non-turing-complete-design.md) - Intentional limitation for safety and predictability
- [ADR-0003: Pratt Parser for Language Parsing](0003-pratt-parser.md) - Choosing Pratt parser over alternatives
- [ADR-0004: Namespace-Based Organization](0004-namespace-based-organization.md) - Organizing policies with namespaces
- [ADR-0005: Shape-Based Type System](0005-shape-based-type-system.md) - Type definitions using shapes
- [ADR-0007: Fact-Based Input Model](0007-fact-based-input-model.md) - How external data is provided to policies
- [ADR-0008: Rule Evaluation Model](0008-rule-evaluation-model.md) - How rules are evaluated and decisions exported
- [ADR-0009: Trinary Logic System (True/False/Unknown)](0009-trinary-logic-system.md) - Using three-valued logic instead of boolean
- [ADR-0012: Default Values in Rules](0012-default-values-in-rules.md) - Handling missing or undefined rule results

### Runtime & Execution

- [ADR-0010: TypeScript Module Support via goja](0010-typescript-modules-via-goja.md) - JavaScript runtime for extensions

### Architecture & Distribution

- [ADR-0011: Pack System for Policy Organization](0011-pack-system.md) - Packaging policies and related files

## Status Legend

- **Proposed**: Decision is under consideration
- **Accepted**: Decision has been made and implemented
- **Deprecated**: Decision has been superseded or is no longer relevant
- **Superseded**: Decision has been replaced by another ADR

## Contributing

When making a significant architectural decision:

1. Create a new ADR using the [template](template.md)
2. Number it sequentially (next available number)
3. Use descriptive, kebab-case filenames
4. Update this README with a link to the new ADR
5. Submit as part of your PR

## References

- [ADR Template](template.md)
- [Documentation on ADRs](https://adr.github.io/)
- [Michael Nygard's original ADR format](http://thinkrelevance.com/blog/2011/11/15/documenting-architecture-decisions)
