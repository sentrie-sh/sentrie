# ADR-0012: Default Values in Rules

**Status:** Accepted  
**Date:** 2025-11-09  
**Deciders:** [@binaek](https://github.com/binaek)  
**Tags:** language-design, evaluation, safety

## Context

Rules in policies may not always yield a value. This can happen when:

- Conditional logic doesn't match any branch
- Data is missing (Unknown values)
- Rule conditions are not satisfied

The system needs to decide what value a rule should have when it doesn't explicitly yield one. Options include: no value (error), implicit defaults, or explicit defaults.

## Decision

Rules in Sentrie **must specify explicit default values**. The syntax is:

```sentrie
rule name = default <value> {
  // rule logic
}
```

If a rule doesn't yield a value during evaluation, it returns its default value. Defaults can be any type: trinary values, numbers, strings, collections, etc.

## Rationale

1. **Predictability**: Explicit defaults make it clear what happens when conditions aren't met, preventing surprises.

2. **Safety**: Defaults allow policies to handle edge cases gracefully rather than failing or returning Unknown.

3. **Fail-safe defaults**: Policies can use fail-safe defaults (e.g., `default false` for access control) to ensure security.

4. **Explicit intent**: Requiring explicit defaults forces policy authors to think about edge cases and make intentional decisions.

5. **Type safety**: Defaults must match the rule's return type, providing type checking.

6. **No implicit behavior**: There's no guessing about what happens when a rule doesn't yield - the default is always explicit.

7. **Composability**: Rules with defaults can be safely composed, as their behavior is predictable.

## Consequences

### Positive

- Predictable rule behavior
- Explicit handling of edge cases
- Type-safe defaults
- Forces intentional design decisions
- Supports fail-safe security defaults
- No implicit behavior to remember

### Negative

- Must specify defaults for all rules
- More verbose syntax
- Must think about edge cases upfront
- Defaults must be appropriate for the use case

### Neutral

- Defaults are evaluated if rule doesn't yield
- Defaults can be complex expressions
- Defaults are type-checked

## Alternatives Considered

### Alternative A: No Defaults (Error on No Yield)

**Description:** If a rule doesn't yield a value, evaluation fails with an error.

**Pros:**

- Forces explicit handling of all cases
- No implicit behavior
- Clear when logic is incomplete

**Cons:**

- Policies fail when edge cases occur
- Less graceful error handling
- Harder to write defensive policies
- Can't use fail-safe defaults

**Why not chosen:** Explicit defaults provide the same benefits (forcing intentional decisions) while allowing graceful handling of edge cases and fail-safe defaults.

### Alternative B: Implicit Defaults

**Description:** Use implicit defaults based on type (false for boolean, 0 for numbers, empty string, etc.).

**Pros:**

- Less verbose
- Familiar (like many languages)
- Quick to write

**Cons:**

- Implicit behavior can be surprising
- May not be appropriate for all use cases
- Harder to reason about
- Can hide bugs (wrong default assumed)

**Why not chosen:** Explicit defaults are clearer and force intentional decisions. The verbosity is worth the clarity and safety.

### Alternative C: Optional Return Types

**Description:** Rules can return optional values, and callers must handle the None case.

**Pros:**

- Type system enforces handling
- Very explicit about possibility of no value
- Familiar pattern (like Rust Option, Haskell Maybe)

**Cons:**

- More complex type system
- Requires unwrapping/pattern matching everywhere
- More verbose for simple cases
- Doesn't solve the "what should the default be" question

**Why not chosen:** Explicit defaults provide similar benefits (forcing intentional decisions) with simpler semantics. Optional types add complexity without clear benefits for policy evaluation.

### Alternative D: Default to Unknown

**Description:** Rules that don't yield default to Unknown (trinary).

**Pros:**

- Consistent with trinary logic
- Propagates uncertainty
- Simple rule

**Cons:**

- Unknown may not be appropriate for all rule types
- Can't use fail-safe defaults
- Less flexible
- Forces all rules to handle Unknown

**Why not chosen:** Explicit defaults provide more flexibility. Rules can still return Unknown if that's the appropriate default, but they can also use other defaults when appropriate.

## Implementation Notes

- Default values are part of rule syntax: `rule name = default value { ... }`
- Defaults are type-checked against the rule's return type
- Defaults are evaluated if the rule body doesn't yield a value
- Defaults can be any expression of the appropriate type
- Defaults are evaluated in the same context as rule bodies

## References

- Language documentation: [Rules](https://sentrie.sh/reference/rules/)
- Implementation: `parser/rule.go`, `ast/rule.go`
- Related: ADR-0001 (The Sentrie Language), ADR-0008 (Rule Evaluation Model), ADR-0009 (Trinary Logic System)
