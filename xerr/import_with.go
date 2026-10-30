// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package xerr

import (
	"fmt"
	"strings"

	"github.com/sentrie-sh/sentrie/tokens"
)

// ImportUnknownFactError is a validate-time `with` binding whose name matches
// no fact of the imported policy.
type ImportUnknownFactError struct {
	At      tokens.Range
	Message string
}

func (e ImportUnknownFactError) Error() string {
	return fmt.Sprintf("%s: %s: %s", e.At.String(), e.Message, ErrIndex.Error())
}

func (e ImportUnknownFactError) Unwrap() error { return ErrIndex }

// ErrImportUnknownFact reports, anchored at the with clause, a `with` name that
// the imported policy does not declare. facts must already be sorted.
func ErrImportUnknownFact(at tokens.Range, with, policy string, facts []string) error {
	return ImportUnknownFactError{At: at, Message: ImportUnknownFactMessage(with, policy, facts)}
}

// ImportUnknownFactMessage is the message shared by the index check and the
// runtime fallback, so the two always read the same. facts must already be sorted.
func ImportUnknownFactMessage(with, policy string, facts []string) string {
	declared := "declares no facts"
	if len(facts) > 0 {
		declared = "declares facts: " + strings.Join(facts, ", ")
	}
	return fmt.Sprintf("import binds unknown fact %q: policy %s %s", with, policy, declared)
}
