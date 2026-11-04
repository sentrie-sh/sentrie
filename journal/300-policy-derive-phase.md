## Binaek Sarkar (binaek89@gmail.com)

### 23:36 - Pick up #300
- A `derive` that is the first body statement of a policy is rejected as a late header (`'derive' must appear before rules, exports, lets, and shapes`), so a policy-level derive is only accepted after a `let`, `rule` or `shape`.
- Picked up with #296 to #300 in one batch (follow-ups from #279 to #285); one pickup push and a draft PR per branch before implementation.
- Plan: a policy-level `derive` starts the body like a `let` does; facts and uses after it are still late headers.
- _Model: claude-opus-5-5_

### 2026-11-03 21:09 - Start the policy body at a derive
- Changed the `*ast.DeriveStatement` case in `createPolicy` (index/policy.go) from rejecting a derive in the meta, facts and uses phases to latching the phase to `body`, the same as let, rule, export and shape. A fact, use or metadata statement after a derive is now a late header, as it is after a let.
- The late-header message listed "rules, exports, lets, and shapes" only; with derives now starting the body it would mislead for `derive` then `fact`, so it now reads "rules, exports, lets, derives, and shapes". Updated the three existing assertions on the old text (policy_test.go, index_test.go).
- Deleted `TestPolicyDeriveLateHeaderRejected` in derive_coverage_test.go: it pinned the bug (a derive directly after a fact must be rejected). New tests in index/policy_derive_phase_test.go cover derive first, after meta, after facts, after uses, and fact/use/title after a derive; all seven fail on the old code.
- Checked namespace-level ordering (`AddProgram` in index/index.go): it has no phase state, top-level derive/shape/policy/export statements are accepted in any order, so there is no matching asymmetry.
- Rewrote the legacy headers of index/policy.go, policy_test.go, index_test.go and derive_coverage_test.go (already new form) as touched files. Updated the `index.policy` graph node (exceptions list and phase-ordering gotcha).
- Verified end to end with the repro pack: `sentrie validate` passes and `exec t/a/p/ok` returns true.
- _Model: claude-sonnet-5-5_

### 2026-11-03 21:12 - End-to-end tests for a derive-first policy
- Added runtime/policy_derive_phase_test.go: the #300 repro through the executor (derive first, rules call it, ExecRule true/false and ExecPolicy) and a derive after a fact called with the fact value.
- Added two cmd tests in cmd/validate_test.go: `sentrie validate` accepts the derive-first policy, and rejects `fact` after a derive with the new "lets, derives, and shapes" message.
- All four fail on main's index/policy.go (verified by swapping the file and restoring it); the index-level tests already in policy_derive_phase_test.go were not duplicated.
- Ruled out: other callers of the old message (rg finds none outside index tests already updated); the many `let _seed = 0` workarounds in existing derive tests are left alone (would be a separate cleanup, they still pass).
- _Model: claude-sonnet-5-5_
