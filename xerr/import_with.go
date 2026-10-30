// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package xerr

import (
	"fmt"
	"maps"
	"slices"
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

// ImportUnknownFactMessage is the message shared by the index check and the
// runtime fallback. facts maps each fact alias of the imported policy (the
// name a `with` clause must use, because Policy.Facts is keyed by alias) to the
// fact's declared name. When a fact's declared name equals the `with` name but
// its alias differs, the message says so.
func ImportUnknownFactMessage(with, policy string, facts map[string]string) string {
	aliases := slices.Sorted(maps.Keys(facts))
	msg := fmt.Sprintf("import binds unknown fact %q: policy %s ", with, policy)
	if len(aliases) == 0 {
		return msg + "declares no facts"
	}
	msg += "accepts fact aliases: " + strings.Join(aliases, ", ")
	for _, alias := range aliases {
		if facts[alias] == with {
			msg += fmt.Sprintf("; fact %q is bound by its alias %q", with, alias)
		}
	}
	return msg
}
