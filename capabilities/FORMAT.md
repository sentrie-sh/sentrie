# Capabilities file format

A capabilities file is a deployment-level JSON document that limits what a policy
pack may use. The operator passes it to `sentrie exec`, `sentrie serve` and
`sentrie validate` with `--capabilities <path>`. OPA's `capabilities.json` served
as the reference.

A pack's `[permissions]` table in `sentrie.pack.toml` is the pack's **request**. A
capabilities file is the operator's **grant**. Without `--capabilities`, everything
is granted, which is how Sentrie behaved before the file existed.

## Example

```json
{
  "version": 1,
  "builtins": ["all", "any", "count", "filter", "merge"],
  "modules": ["@sentrie/hash", "@sentrie/jwt", "@sentrie/net"],
  "local_modules": false
}
```

This pack may call five builtins and load three builtin modules, and it may not
load any TypeScript or JavaScript files of its own.

## Fields

| Field | Type | Absent | Meaning |
|-------|------|--------|---------|
| `version` | integer | error | File format version. Must be `1`. |
| `builtins` | array of strings | every builtin is granted | Builtins the pack may call, by name (`"count"`, `"now"`). `[]` grants none. |
| `modules` | array of strings | every builtin module is granted | Builtin modules the pack may load, by full specifier (`"@sentrie/hash"`). `[]` grants none. |
| `local_modules` | boolean | `true` | Whether the pack may load its own TypeScript or JavaScript files. |

`null` counts as absent. These are errors, and the command refuses to start: an
empty file, a top-level value that is not an object, a missing `version` or any
other `version`, an unknown key, a key spelled with different case (`"Builtins"`),
a key given twice, data after the object, an empty builtin name, and a module
entry without the `@sentrie/` prefix. Syntax, type, unknown-key and duplicate-key
errors name the line and column, for example
`caps.json: capabilities: 3:21: invalid character ']' looking for beginning of value`.

Names are not checked against the builtins and modules this build ships. A file
written for a newer Sentrie still loads, and a name this build does not know grants
nothing.

## What is checked, and when

After the pack is parsed, indexed and validated, and before it is executed or
served, every use is checked against the grant. Each denial is reported with its
source span, and the command fails:

```
policy.sentrie:4:21-26: builtin "count": not granted by the capabilities file
```

A `use` that is not granted is reported the same way, at the `use` statement,
as `module "@sentrie/hash"` or `local module "./helper"`.

- **Builtins:** every call to a native builtin, in rules, policy `let`s, fact
  defaults, export attachments and derive bodies. A name shadowed by a `let`, a
  lambda parameter or a derive is not a builtin call.
- **Builtin modules:** every `use ... from @sentrie/<name>` statement.
- **Local modules:** every `use ... from "./path"` and `use ... from @local/...`
  statement, when `local_modules` is `false`.
- **`require()` in TypeScript:** every module a `use` loads is compiled, and each
  static `require("...")` (every `import` compiles to one) in pack-local code is
  checked against `modules` and `local_modules`, following pack-local requires
  transitively. So `serve` refuses to start instead of failing the first request
  that reaches the module.

The runtime enforces the same grant again:

- Builtin dispatch refuses a builtin that is not granted.
- Module resolution refuses a module that is not granted, including a dynamic
  `require(expr)` that the load-time scan cannot see.

Runtime denials wrap the same `not granted by the capabilities file` error.

## Nondeterministic sources

`sentrie validate` prints a warning for every nondeterministic source a pack uses,
whether or not a capabilities file is given. These sources are the `now` builtin,
`@sentrie/time.now`, `@sentrie/uuid.v4`, `v6` and `v7`, and `@sentrie/js.now` and
`random`. To forbid them, leave them out of `builtins` and `modules`, and set
`local_modules` to `false`: a pack's own TypeScript can call `Date.now()` and
`Math.random()` directly.
