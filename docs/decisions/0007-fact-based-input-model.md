# ADR-0007: Fact-Based Input Model

**Status:** Accepted  
**Date:** 2025-10-05  
**Deciders:** [@binaek](https://github.com/binaek)  
**Tags:** language-design, input-model, evaluation

## Context

Policies need to receive external data to make decisions. The input model must:

- Provide data to policies in a structured way
- Be type-safe and validated
- Support multiple data items
- Be easy to use from APIs and CLIs
- Clearly separate policy logic from input data

Different input models exist: function parameters, global variables, dependency injection, or fact-based models.

## Decision

Sentrie uses a **fact-based input model**. Policies declare the facts they need using `fact` statements with type annotations. Facts are provided as JSON data when evaluating policies. Each fact has a name and a type (shape), and facts are scoped to the policy.

## Rationale

1. **Explicit data requirements**: Fact declarations make it clear what data a policy needs, improving readability and documentation.

2. **Type safety**: Facts are typed, enabling validation that provided data matches expected types.

3. **Separation of concerns**: Facts clearly separate input data from policy logic, making policies more testable and reusable.

4. **JSON compatibility**: Facts map naturally to JSON objects, making integration with APIs and CLIs straightforward.

5. **Multiple facts**: Policies can declare multiple facts, allowing complex data structures without nesting everything in a single object.

6. **Scoped to policy**: Facts are scoped to individual policies, preventing accidental data sharing between policies.

7. **Familiar pattern**: The fact-based model is similar to database facts in logic programming, making it familiar to some developers.

## Consequences

### Positive

- Clear declaration of data requirements
- Type-safe input validation
- Easy to test (provide test facts)
- Natural JSON integration
- Self-documenting (facts show what data is needed)
- Supports multiple independent facts

### Negative

- Must declare facts before using them
- Facts are policy-scoped (cannot share across policies easily)
- JSON structure must match fact declarations
- More verbose than global variables

### Neutral

- Facts are provided at evaluation time, not compile time
- Facts can be optional (with default handling)
- Fact names are identifiers in the policy scope

## Alternatives Considered

### Alternative A: Function Parameters

**Description:** Treat policies as functions that take parameters, similar to function calls.

**Pros:**

- Familiar to developers (like function calls)
- Clear parameter list
- Type-safe parameters

**Cons:**

- Less flexible for JSON input
- Harder to provide partial data
- Parameter order matters
- Less intuitive for policy evaluation

**Why not chosen:** Facts provide better JSON integration and are more flexible for providing data. The fact-based model is more natural for policy evaluation.

### Alternative B: Global Variables

**Description:** Use global variables that are set before policy evaluation.

**Pros:**

- Simple to use
- No declarations needed
- Familiar pattern

**Cons:**

- No type safety
- Easy to have naming conflicts
- Hard to track what data is used
- Poor testability (global state)
- No clear data contract

**Why not chosen:** Facts provide type safety, clear contracts, and better testability without the downsides of global state.

### Alternative C: Dependency Injection

**Description:** Use dependency injection to provide data to policies.

**Pros:**

- Flexible and powerful
- Can inject services, not just data
- Supports complex scenarios

**Cons:**

- More complex to implement
- Overkill for simple data input
- Harder to understand
- Requires framework support

**Why not chosen:** Facts provide sufficient functionality with much simpler semantics. Dependency injection adds complexity without clear benefits for policy evaluation.

### Alternative D: Single Input Object

**Description:** Provide a single JSON object containing all data, accessed via property access.

**Pros:**

- Simple structure
- One object to pass around
- Familiar (like function parameters)

**Cons:**

- No type declarations
- Hard to validate structure
- Less clear what data is needed
- All data must be nested in one object

**Why not chosen:** Facts provide type safety and clearer data contracts. Multiple facts are more flexible than a single nested object.

## Implementation Notes

- Facts are declared in policies using `fact name:Type` syntax
- Facts are provided as a JSON object mapping fact names to values
- Type checking validates that provided facts match declared types
- Facts are available as variables in rule evaluation
- Missing facts can result in Unknown values (trinary logic)
- Facts are scoped to the policy where they're declared

## References

- Language documentation: [Facts](https://sentrie.sh/reference/facts/)
- Implementation: `parser/fact.go`, `ast/fact.go`
- Related: ADR-0001 (The Sentrie Language), ADR-0005 (Shape-Based Type System), ADR-0009 (Trinary Logic System)
