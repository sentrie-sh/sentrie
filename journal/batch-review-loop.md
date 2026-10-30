# batch-review-loop

## Binaek Sarkar (binaek89@gmail.com)

### 2026-10-30 23:04 - Write down the implement, test and review loop as a rule
- Asked for after the batch of #236, #237, #134, #135, #136, #150, #151 and #117 (PRs #249 to #256): record the procedure that batch followed so later batches repeat it.
- New `.agentrules/implement-review-test-loop.md` covers the parts no rule held yet:
  - asking the user for open design choices before implementers start, and recording the answer in the pickup journal entry;
  - three separate agents per issue (implementer, test and CI agent, reviewer) and the loop back to the same implementer until `VERDICT: APPROVE` and the local CI gate pass;
  - the test agent proving each test bites against `main`'s production file;
  - merging the batch together in a throwaway clone, keeping combination-only fixes out of the branches as a recipe for whichever PR merges second;
  - queueing commits as an exact script when signing fails or the user holds commits, then one push and `gh pr edit` from the PR description files.
- Pickup, origin ownership, model choice and the two-attempt limit stay in their own rules; the new rule links to them instead of restating them.
- Wrote down the one exception from the batch: the throwaway clone's merge commits skipped signing, as they are never pushed. Stated as allowed only there, so it is visible for review.
- Branched off `main` with no issue, as a standalone conventions change ([PR conventions](../.agentrules/pr-conventions.md)).
- _Model: claude-opus-5-5_
