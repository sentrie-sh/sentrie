## Binaek Sarkar (binaek89@gmail.com)

### 22:09 - Pick up #213
- Branch cut from `212-list-unique-unhashable`, top of the chain 208 to 213. #111 set the "predicates answer false" convention in `builtin_net.go` and #211 changed the same file, so this stacks on both.
- Plan: classification predicates throw on input that is not an IP; validity predicates keep answering `false`.
- _Model: claude-opus-5-5_

### 22:10 - Throw from the classification predicates
- Audited every `ex.Set` in `builtin_net.go` for fail-open answers on junk. Only `isPrivate`, `isPublic`, `isLoopback` and `isMulticast` classify; `cidrContains` and `cidrIntersects` already throw on bad input (#111), `parseIP` returns `null`, and `isIPv4`, `isIPv6`, `cidrIsValid` are validity tests that keep answering `false`.
- The four classifiers now share `classifyIP(vm, name, classify)`, which keeps the argument-count check and throws `<name>: invalid IP address: "<input>"` when `net.ParseIP` fails. This also removed four copies of the same parse block.
- A CIDR string such as `10.0.0.0/8` is not an IP address, so it throws too.
- Updated `TestBuiltinPredicatesAndNoResult` (from #111): its `isPrivate`/`isPublic` junk-answers-false rows encoded exactly the behaviour this issue removes; replaced with `isIPv4`/`isIPv6` junk rows. This supersedes #111's "every `net` `is*` check answers false" convention for the classifiers.
- New tests in `builtin_net_test.go`: each classifier throws on junk, empty and CIDR input, in plain and negated form, and the throw is catchable; valid addresses still classify, and validity predicates still answer `false` for junk.
- Breaking: a policy that relied on `isPrivate("junk") === false` now fails the rule. `.d.ts` (`@throws` on each classifier, junk note on `isIPv4`/`isIPv6`) and the `runtime.js.builtins` graph node (conventions bullet and the old "predicates answer false" gotcha) document the split.
- _Model: claude-opus-5-5_
