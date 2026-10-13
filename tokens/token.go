// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package tokens

import (
	"fmt"
	"maps"
	"slices"
)

type Instance struct {
	Kind  Kind
	Value string
	Range Range
}

func EofInstance(file string, pos Pos) Instance {
	return Instance{
		Kind:  EOF,
		Value: "",
		Range: NewRangeFromPos(file, pos),
	}
}

func New(kind Kind, value string, r Range) Instance {
	return Instance{
		Kind:  kind,
		Value: value,
		Range: r,
	}
}

func Err(r Range, message string) Instance {
	return Instance{
		Kind:  Error,
		Value: message,
		Range: r,
	}
}

func (t Instance) IsOfKind(kinds ...Kind) bool {
	return slices.Contains(kinds, t.Kind)
}

func (t Instance) String() string {
	keywordTokens := slices.Collect(maps.Values(keywords))
	if t.IsOfKind(EOF) {
		return "<EOF>"
	} else if t.IsOfKind(KeywordNull) {
		return fmt.Sprintf("%s(null)", t.Kind)
	} else if slices.Contains(keywordTokens, t.Kind) {
		return fmt.Sprintf("%s()", t.Kind)
	}
	return fmt.Sprintf("%s(%q)", t.Kind, t.Value)
}
