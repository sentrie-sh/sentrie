# ADR-0001: Trinary Logic System (True/False/Unknown)

**Status:** Accepted  
**Date:** 2025-11-08  
**Deciders:** [@binaek](https://github.com/binaek)  
**Tags:** language-design, type-system, evaluation

## Context

Policy evaluation often deals with incomplete or missing data. In many scenarios, we need to distinguish between:

- A condition that is explicitly true
- A condition that is explicitly false
- A condition where the result cannot be determined (missing data, undefined values, etc.)

Traditional boolean logic only provides two states (true/false), which forces policy authors to make assumptions about missing data or handle undefined cases explicitly throughout their policies.

## Decision

Sentrie implements a **trinary logic system** with three values:

- `True`: The condition is satisfied
- `False`: The condition is not satisfied
- `Unknown`: The result cannot be determined (missing data, undefined values, etc.)

This trinary system is used throughout the language for all logical operations, comparisons, and rule evaluations.

## Rationale

1. **Real-world data is incomplete**: Policies often need to evaluate conditions on data that may be missing or undefined. Trinary logic provides a first-class way to handle this.

2. **Explicit handling of uncertainty**: Rather than defaulting to `false` (which could hide bugs) or `true` (which could be insecure), `Unknown` forces policy authors to explicitly consider what should happen when data is missing.

3. **Kleene logic compatibility**: The implementation follows Kleene's three-valued logic, which is well-established and mathematically sound. This provides predictable behavior:

   - `True AND Unknown = Unknown`
   - `False AND Unknown = False`
   - `True OR Unknown = True`
   - `False OR Unknown = Unknown`

4. **Auditability**: Unknown values make it clear when a policy decision cannot be made due to missing information, improving audit trails and debugging.

5. **Type safety**: The trinary system is integrated into the type system, preventing accidental misuse of boolean values where trinary logic is needed.

## Consequences

### Positive

- Policies can explicitly handle missing or undefined data
- More accurate representation of real-world policy evaluation scenarios
- Better audit trails showing when decisions couldn't be made
- Mathematically sound logic system (Kleene logic)
- Prevents silent failures from missing data

### Negative

- More complex than boolean logic (three states instead of two)
- Policy authors must understand trinary logic semantics
- Some operations may propagate Unknown when boolean would fail fast
- JSON serialization requires string representation ("true", "false", "unknown")

### Neutral

- All logical operations (AND, OR, NOT) must handle three states
- Comparison operations return trinary values
- Rule evaluation results are trinary

## Alternatives Considered

### Alternative A: Boolean with Null/Undefined Handling

**Description:** Use boolean logic but add explicit null/undefined checks in the language.

**Pros:**

- Simpler mental model (just true/false)
- Familiar to most developers
- Easier JSON serialization

**Cons:**

- Requires explicit null checks everywhere
- Easy to forget null handling, leading to bugs
- No first-class representation of uncertainty
- More verbose policy code

**Why not chosen:** The trinary system provides a more elegant and safer way to handle missing data without requiring explicit checks throughout policies.

### Alternative B: Optional Types

**Description:** Use optional/maybe types to represent missing values, with boolean logic.

**Pros:**

- Type system enforces handling of missing values
- Can be more expressive in some cases

**Cons:**

- More complex type system
- Requires unwrapping/pattern matching
- Doesn't solve the logical operation semantics question
- Less intuitive for policy authors

**Why not chosen:** Trinary logic provides the benefits of optional types for logical operations while being simpler and more intuitive for policy evaluation.

### Alternative C: Default to False for Missing Data

**Description:** Use boolean logic and default all missing/undefined values to false.

**Pros:**

- Simple boolean logic
- Familiar semantics
- Fast evaluation (fail-closed security model)

**Cons:**

- Hides missing data problems
- May lead to incorrect denials when data is simply missing
- No way to distinguish "explicitly denied" from "missing data"
- Poor auditability

**Why not chosen:** This approach hides important information about why a decision was made and can lead to incorrect policy evaluations.

## Implementation Notes

- The trinary `Value` type is defined in `trinary/tristate.go`
- All logical operations (AND, OR, NOT) implement Kleene logic
- The `From()` function provides automatic coercion from Go types to trinary values
- JSON serialization uses string values: "true", "false", "unknown"
- The language keywords `true`, `false`, and `unknown` map to trinary values

## References

- [Kleene's three-valued logic](https://en.wikipedia.org/wiki/Three-valued_logic#Kleene_logic)
- Implementation: `trinary/tristate.go`
- Language documentation: [Trinary Logic](https://sentrie.sh/reference/trinary/)
- Related: ADR-0001 (The Sentrie Language)
