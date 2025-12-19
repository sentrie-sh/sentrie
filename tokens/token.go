// SPDX-License-Identifier: Apache-2.0

// Copyright 2025 Binaek Sarkar
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tokens

import (
	"fmt"
	"slices"

	"golang.org/x/exp/maps"
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
	keywordTokens := maps.Values(keywords)
	if t.IsOfKind(EOF) {
		return "<EOF>"
	} else if t.IsOfKind(KeywordNull) {
		return fmt.Sprintf("%s(null)", t.Kind)
	} else if slices.Contains(keywordTokens, t.Kind) {
		return fmt.Sprintf("%s()", t.Kind)
	}
	return fmt.Sprintf("%s(%q)", t.Kind, t.Value)
}
