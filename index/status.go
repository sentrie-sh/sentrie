// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

// Status is a point-in-time, read-only report on an Index, for health checks.
// Taking it never validates or commits the index.
type Status struct {
	// Validated is true once validation has finished, whatever its outcome.
	Validated bool
	// ValidationError is the failure Validate recorded, nil on success or before
	// validation. When validation passed but the follow-up commit failed, this is
	// the commit failure too, which is also in CommitError.
	ValidationError error
	// Committed is true once the commit has finished, whatever its outcome.
	Committed   bool
	CommitError error

	Namespaces int
	Policies   int
	Programs   int
	Rules      int
	Shapes     int
}

// Status reports the index's validation and commit state and how many of each
// kind of definition it holds. It is safe to call while Validate or Commit runs
// on another goroutine.
func (idx *Index) Status() Status {
	idx.stateMu.Lock()
	st := Status{
		Validated:       idx.validated == 1,
		ValidationError: idx.validationError,
		Committed:       idx.committed == 1,
		CommitError:     idx.commitError,
	}
	idx.stateMu.Unlock()

	idx.theLock.RLock()
	defer idx.theLock.RUnlock()
	st.Namespaces = len(idx.Namespaces)
	st.Programs = len(idx.Programs)
	for _, ns := range idx.Namespaces {
		st.Shapes += len(ns.Shapes)
		st.Policies += len(ns.Policies)
		for _, p := range ns.Policies {
			st.Rules += len(p.Rules)
			st.Shapes += len(p.Shapes)
		}
	}
	return st
}
