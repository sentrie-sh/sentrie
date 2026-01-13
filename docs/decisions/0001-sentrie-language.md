# ADR-0001: The Sentrie Language

**Status:** Accepted  
**Date:** 2025-07-28  
**Deciders:** [@binaek](https://github.com/binaek)  
**Tags:** language-design, architecture, foundation

## Decision

Sentrie uses a **dedicated domain-specific language (DSL)** for policy evaluation. The language is:

- **Declarative**: Policies express business rules as declarative statements, not imperative code
- **Non-Turing complete**: Guaranteed termination and predictable performance
- **Type-safe**: Shape-based type system with compile-time checking
- **Trinary logic**: Three-valued logic (True/False/Unknown) for handling incomplete data
- **Namespace-organized**: Policies organized in namespaces to avoid naming conflicts
- **Fact-based input**: External data provided as typed facts
- **Rule-based evaluation**: Rules yield values and can reference other rules with explicit defaults
- **Extensible**: TypeScript modules provide escape hatch for complex computations

## Language Structure

The language consists of:

- **Namespaces**: Organize policies, rules, and shapes
- **Shapes**: Define data structures and types
- **Policies**: Contain rules and facts
- **Facts**: Declare typed input data
- **Rules**: Evaluate to values (trinary, numbers, strings, collections, etc.)
- **Exports**: Specify which rule values to return as decisions

## Keywords

### Declarations

- `namespace` - Declare a namespace
- `policy` - Define a policy
- `rule` - Define a rule
- `shape` - Define a type/shape
- `fact` - Declare input data
- `let` - Bind a value to a variable
- `export` - Export a decision
- `import` - Import from another policy
- `use` - Use a TypeScript module

### Rule Modifiers

- `default` - Specify default value for a rule
- `when` - Conditional rule evaluation
- `yield` - Return a value from a rule

### Type Keywords

- `string`, `number`, `boolean`, `trinary` - Primitive types
- `list`, `map`, `record`, `document` - Collection types
- `null` - Null value

### Logical Operators

- `and`, `or`, `xor`, `not` - Logical operations
- `true`, `false`, `unknown` - Trinary values

### Collection Operations

- `any` - Check if any element matches
- `all` - Check if all elements match
- `filter` - Filter collection elements
- `map` - Transform collection elements
- `reduce` - Aggregate collection elements
- `distinct` - Get unique elements
- `first` - Get first element
- `count` - Count elements

### Membership and Matching

- `in` - Check membership
- `contains` - Check if collection contains value
- `matches` - Pattern matching
- `is` - Type checking
- `defined` - Check if value is defined
- `empty` - Check if collection is empty

### Other Keywords

- `as` - Alias in collection operations
- `from` - Source in imports and reduces
- `with` - Parameter binding in imports
- `of` - Used in export statements
- `cast` - Type casting
- `attach` - Attach expressions to exports

## Language Constructs

### Namespace Declaration

```sentrie
namespace com.example.policy
```

### Shape Definition

```sentrie
shape User {
  name: string
  age?: number  // optional field
  role!: string // required field
  tags: list[string]
}
```

### Policy Definition

```sentrie
policy access_control {
  fact user: User

  rule allow = default false {
    yield user.role == "admin"
  }

  export decision of allow
}
```

### Fact Declaration

```sentrie
fact user: User
fact config?: Config  // optional fact
```

### Rule Definition

```sentrie
rule name = default value {
  yield expression
}

rule conditional = default false when condition {
  yield expression
}
```

### Let Bindings

```sentrie
let x = 10
let y = user.name
let result = x + y
```

### Collection Operations

```sentrie
let evens = filter numbers as n { yield n % 2 == 0 }
let exists = any items as item { yield item.active }
let allValid = all items as item { yield item.valid }
let doubled = map items as item { yield item * 2 }
let sum = reduce 0 from numbers as acc, n { yield acc + n }
let unique = distinct items as item { yield item }
let first = first items as item { yield item }
let total = count items
```

### Type System

- **Primitive types**: `string`, `number`, `boolean`, `trinary`
- **Collection types**: `list[T]`, `map[K, V]`, `record`, `document`
- **Shape types**: User-defined shapes
- **Optional fields**: `field?: Type` (nullable)
- **Required fields**: `field!: Type` (non-nullable)
- **Constraints**: `@min(0)`, `@max(100)`, `@length(10)`, etc.

### Operators

- **Arithmetic**: `+`, `-`, `*`, `/`, `%`
- **Comparison**: `==`, `!=`, `<`, `>`, `<=`, `>=`
- **Logical**: `and`, `or`, `xor`, `not`
- **Membership**: `in`, `contains`
- **Ternary**: `condition ? trueValue : falseValue`
- **Access**: `.` (dot notation), `[]` (index access)

### Expressions

- **Literals**: Strings, numbers, booleans, trinary values, null
- **Collections**: `[1, 2, 3]`, `{"key": "value"}`
- **Access**: `user.name`, `items[0]`
- **Function calls**: `count(items)`, `module.function(args)`
- **Type checks**: `value is Type`, `value is defined`, `value is empty`
- **Pattern matching**: `value matches pattern`

### Import and Use

```sentrie
rule check = import decision isValid from other_policy
             with user as subject
             with resource as obj

use crypto from "./crypto.ts"
let hash = crypto.sha256(data)
```

### Export

```sentrie
export decision of ruleName
export decision of ruleName attach expression
```

## Key Design Principles

1. **Safety first**: Non-Turing completeness ensures bounded execution
2. **Type safety**: All data is typed and validated at policy definition time
3. **Explicit handling**: Unknown values and defaults must be explicitly considered
4. **Portability**: Policies are independent of application implementation
5. **Simplicity**: Language focuses on expressing business rules clearly

## References

- Language documentation: [What is Sentrie?](https://sentrie.sh/getting-started/what-is-sentrie/)
- Language reference: [Reference Documentation](https://sentrie.sh/reference/)
- Related: ADR-0002 (Non-Turing Complete Design), ADR-0003 (Pratt Parser), ADR-0004 (Namespace-Based Organization), ADR-0005 (Shape-Based Type System), ADR-0007 (Fact-Based Input Model), ADR-0008 (Rule Evaluation Model), ADR-0009 (Trinary Logic System), ADR-0010 (TypeScript Module Support), ADR-0011 (Pack System), ADR-0012 (Default Values in Rules)
