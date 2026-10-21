## Binaek Sarkar (binaek89@gmail.com)

### 22:05 - Pick up #211
- Branch cut from `210-bound-regex-cache`. #111 rewrites the error sites in `builtin_net.go`, and #213 (two layers up) changes the same file, so this stays in the chain 208 to 213 rather than branching from `main`.
- Plan: `cidrExpand` throws above a named 65,536-host limit, checked from the prefix length before any allocation.
- _Model: claude-opus-5-5_

### 22:06 - Cap `cidrExpand` at 65,536 addresses
- Chose a limit over removing `cidrExpand`: removal breaks every caller, while a limit only refuses ranges that could never have been useful in a policy (a /15 is already 131k strings).
- `cidrExpandMaxHostBits = 16`, `cidrExpandMaxHosts = 1 << 16`, documented at the constant. The check is `bits - ones` from `ipNet.Mask.Size()`, so it runs before the loop and before any allocation, and the error reports the size as `2^n` (`::/0` is 2^128, past any integer type).
- The limit is the same address count for both families: a /16 in IPv4, a /112 in IPv6.
- Since the count is now known and bounded, the slice is preallocated to `1 << hostBits`.
- Tests in a new `builtin_net_test.go`: `10.0.0.0/8`, `10.0.0.0/15`, `::/0`, `/64` and `/111` IPv6 throw in under a second; `/30`, `/32`, IPv6 `/127`, and exactly-at-limit `/16` and `/112` still enumerate.
- Breaking only for ranges above the limit, which previously exhausted memory or ran for minutes. `.d.ts` and the `runtime.js.builtins` graph node document the limit.
- _Model: claude-opus-5-5_
