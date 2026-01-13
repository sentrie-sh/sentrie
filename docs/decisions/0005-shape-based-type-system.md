# ADR-0005: Shape-Based Type System

**Status:** Accepted  
**Date:** 2025-09-28  
**Deciders:** [@binaek](https://github.com/binaek)  
**Tags:** language-design, type-system

## Context

Policies need to work with structured data. The type system must:
- Define the structure of input data (facts)
- Provide type safety and validation
- Be easy to use and understand
- Support common data structures (records, lists, maps, primitives)
- Enable type checking at policy definition time

Different approaches exist: structural types, nominal types, schema languages, or dynamic typing.

## Decision

Sentrie uses a **shape-based type system**. Shapes define the structure of data using a record-like syntax. Shapes are nominal types (identified by name) that can be referenced throughout policies. The type system includes:
- Primitive types: `string`, `number`, `boolean`, `trinary`
- Collection types: `list<T>`, `map<K, V>`
- Shape types: user-defined structures
- Type references with constraints

## Rationale

1. **Type safety**: Shapes provide compile-time type checking, catching errors before policy execution.

2. **Clear data contracts**: Shapes explicitly define what data a policy expects, making policies self-documenting.

3. **Validation**: Shapes can validate that input facts match expected structure, preventing runtime errors.

4. **Familiar syntax**: Record-like syntax is familiar to developers from many languages (structs, classes, interfaces).

5. **Reusability**: Shapes can be defined once and reused across multiple policies.

6. **Namespace scoping**: Shapes belong to namespaces, avoiding naming conflicts.

7. **Extensibility**: The shape system can be extended with constraints and validation rules.

## Consequences

### Positive

- Type safety at policy definition time
- Self-documenting policies (shapes show expected data)
- Early error detection
- Reusable type definitions
- Clear data contracts
- Validation of input facts

### Negative

- Must define shapes before using them
- More verbose than dynamic typing
- Learning curve for developers unfamiliar with type systems
- Type definitions add to policy file size

### Neutral

- Shapes are nominal (name-based), not structural
- Shapes can be exported for use in other namespaces
- Type checking happens during indexing/validation

## Alternatives Considered

### Alternative A: Dynamic Typing

**Description:** Use dynamic typing like JavaScript or Python, with runtime type checking.

**Pros:**
- Flexible and easy to use
- No type definitions needed
- Familiar to many developers

**Cons:**
- Errors only discovered at runtime
- No compile-time safety
- Harder to understand what data is expected
- Poor tooling support (no autocomplete, etc.)

**Why not chosen:** Type safety is critical for policy correctness. Runtime errors in policies can lead to security issues or incorrect decisions.

### Alternative B: Structural Types

**Description:** Use structural typing where types are defined by their structure, not name.

**Pros:**
- More flexible (any structure matching the shape is valid)
- No need to explicitly name types
- Can be more concise

**Cons:**
- Harder to understand type relationships
- Less explicit about intent
- Can lead to accidental type matches
- More complex type checking

**Why not chosen:** Nominal types (shapes) are clearer about intent and provide better error messages. The explicit naming helps with documentation and understanding.

### Alternative C: JSON Schema

**Description:** Use JSON Schema for type definitions instead of a custom shape syntax.

**Pros:**
- Standard format
- Rich validation capabilities
- Tooling support
- Familiar to API developers

**Cons:**
- Verbose syntax
- Not integrated into the language
- Harder to reference in code
- Less readable in policy files

**Why not chosen:** Shapes provide a more integrated, readable syntax that's part of the language. JSON Schema is better suited for external API documentation.

### Alternative D: Interface/Protocol Types

**Description:** Use interface-like types that define required fields without specifying exact structure.

**Pros:**
- Flexible (any type with required fields matches)
- Can be more permissive
- Supports polymorphism

**Cons:**
- Less type safety
- Harder to validate input
- Can be confusing (what fields are actually present?)
- More complex semantics

**Why not chosen:** Shapes provide better type safety and validation. The explicit structure is clearer for policy evaluation.

## Implementation Notes

- Shapes are defined using the `shape` keyword
- Shapes can contain fields with types and optional constraints
- Shapes are parsed in `parser/shape.go`
- Type checking happens in the index system (`index/validate.go`)
- Shapes can reference other shapes
- Type references support constraints (e.g., `list<string>`, `map<string, number>`)

## References

- Language documentation: [Shapes](https://sentrie.sh/reference/shapes/)
- Implementation: `parser/shape.go`, `ast/shape.go`, `index/shape.go`
- Related: ADR-0001 (The Sentrie Language), ADR-0004 (Namespace-Based Organization), ADR-0007 (Fact-Based Input Model)
