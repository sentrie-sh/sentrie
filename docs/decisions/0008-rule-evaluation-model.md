# ADR-0008: Rule Evaluation Model

**Status:** Accepted  
**Date:** 2025-09-14  
**Deciders:** [@binaek](https://github.com/binaek)  
**Tags:** language-design, evaluation, runtime

## Context

Policies need to evaluate rules and produce decisions. The evaluation model must:

- Support conditional logic and data transformations
- Handle trinary logic (True/False/Unknown)
- Allow rules to reference other rules
- Support default values
- Export decisions for external consumption
- Be predictable and auditable

Different evaluation models exist: imperative execution, declarative evaluation, lazy evaluation, or eager evaluation.

## Decision

Sentrie uses a **declarative rule evaluation model** with:

- Rules that yield values (trinary, numbers, strings, collections, etc.)
- Rules can reference other rules (creating a dependency graph)
- Default values for rules when conditions aren't met
- Explicit decision exports that specify which rule values to return
- Eager evaluation (rules are evaluated when needed, but dependencies are resolved first)

## Rationale

1. **Declarative nature**: Rules declare what should be true, not how to compute it. This aligns with policy-as-code philosophy.

2. **Dependency resolution**: Rules can depend on other rules, and the system resolves dependencies automatically.

3. **Default values**: Rules can specify default values when conditions aren't met, providing predictable behavior.

4. **Explicit exports**: Policies explicitly declare which rule values to export as decisions, making the output clear.

5. **Trinary logic**: Rules yield trinary values, allowing Unknown to propagate when data is missing.

6. **Auditability**: The declarative model makes it easier to trace how decisions were reached.

7. **Composability**: Rules can be composed from other rules, enabling reusable policy components.

## Consequences

### Positive

- Clear, declarative policy syntax
- Automatic dependency resolution
- Predictable evaluation with defaults
- Explicit decision outputs
- Supports complex rule dependencies
- Easy to reason about and audit

### Negative

- Must understand rule evaluation semantics
- Dependency cycles must be detected and prevented
- Default values must be carefully considered
- Evaluation order depends on dependencies

### Neutral

- Rules are evaluated eagerly when needed
- Dependencies form a directed acyclic graph (DAG)
- Exports can reference any rule in the policy

## Alternatives Considered

### Alternative A: Imperative Execution Model

**Description:** Use imperative statements that execute in order, like a traditional programming language.

**Pros:**

- Familiar to developers
- Explicit control flow
- Easy to understand execution order

**Cons:**

- Less declarative
- Harder to reason about dependencies
- More verbose for policy logic
- Doesn't fit policy-as-code philosophy

**Why not chosen:** The declarative model better fits policy evaluation. Imperative execution is less suitable for expressing business rules.

### Alternative B: Lazy Evaluation

**Description:** Evaluate rules only when their values are actually needed.

**Pros:**

- Can be more efficient (don't evaluate unused rules)
- Supports infinite data structures (though not applicable here)

**Cons:**

- More complex to implement
- Harder to debug (evaluation happens later)
- Less predictable performance
- Can make errors appear later than expected

**Why not chosen:** Eager evaluation provides better predictability and debugging experience. The performance benefits of lazy evaluation are not significant for policy evaluation.

### Alternative C: Single Expression Evaluation

**Description:** Each rule is a single expression, no dependencies between rules.

**Pros:**

- Simple model
- No dependency resolution needed
- Easy to understand

**Cons:**

- Less composable
- Can't reuse rule logic
- More verbose (must repeat logic)
- Less powerful

**Why not chosen:** Rule dependencies enable composition and reuse, which are important for maintainable policies.

### Alternative D: Constraint-Based Evaluation

**Description:** Use constraint satisfaction where rules are constraints that must be satisfied.

**Pros:**

- Very declarative
- Good for complex constraint scenarios
- Can find solutions to constraint sets

**Cons:**

- More complex to implement
- Less intuitive for simple policies
- Overkill for most policy scenarios
- Harder to understand for developers

**Why not chosen:** The rule-based model provides sufficient declarative power while being more intuitive and easier to implement.

## Implementation Notes

- Rules are defined with `rule name = default value { ... }` syntax
- Rules can reference other rules by name
- The index system builds a dependency graph (DAG) of rules
- Evaluation resolves dependencies before evaluating dependent rules
- Default values are used when rule conditions don't yield a value
- Exports use `export decision of ruleName` syntax
- Rule evaluation happens in the runtime package

## References

- Language documentation: [Rules](https://sentrie.sh/reference/rules/)
- Implementation: `parser/rule.go`, `runtime/`
- Related: ADR-0001 (The Sentrie Language), ADR-0009 (Trinary Logic System), ADR-0012 (Default Values in Rules)
