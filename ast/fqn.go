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

package ast

import (
	"fmt"
	"strings"

	"github.com/sentrie-sh/sentrie/tokens"
)

const FQNSeparator = "/"

type FQN struct {
	*baseNode
	Parts []string
}

func (f FQN) IsEmpty() bool {
	return len(f.Parts) == 0
}

func NewFQN(parts []string, ssp tokens.Range) FQN {
	return FQN{
		baseNode: &baseNode{
			Rnge:  ssp,
			Kind_: "fqn",
		},
		Parts: parts,
	}
}

func (f FQN) Ptr() *FQN {
	return &f
}

func (f FQN) String() string {
	if len(f.Parts) == 0 {
		return ""
	}
	return strings.Join(f.Parts, FQNSeparator)
}

func CreateFQN(base FQN, lastSegment string) FQN {
	if len(base.Parts) == 0 {
		return NewFQN([]string{lastSegment}, base.Rnge)
	}
	return NewFQN(append(base.Parts, lastSegment), base.Rnge)
}

// LastSegment returns the last segment of the FQN
func (f FQN) LastSegment() string {
	if len(f.Parts) == 0 {
		return ""
	}
	return f.Parts[len(f.Parts)-1]
}

// Parent returns the parent of the FQN
func (f FQN) Parent() FQN {
	if len(f.Parts) == 0 {
		return NewFQN([]string{}, f.Rnge)
	}
	return NewFQN(f.Parts[:len(f.Parts)-1], f.Rnge)
}

// IsChildOf returns true if the FQN is a child of another FQN
func (f FQN) IsChildOf(another FQN) bool {
	// ["com","example","foo"] child of ["com","example"]
	// ["com","example","foo"] not child of ["com","example","bar"]
	// ["com","example","foo"] not child of ["com","example2","foo"]
	// ["com","example","foo","bar"] not child of ["com","example"]
	if len(f.Parts)-1 != len(another.Parts) {
		return false
	}

	fqn := f.String()
	supposedToBeChildFQN := another.String()
	return strings.HasPrefix(fqn, fmt.Sprintf("%s%s", supposedToBeChildFQN, FQNSeparator))
}

func (f FQN) IsParentOf(another FQN) bool {
	return another.IsChildOf(f)
}

var _ Node = &FQN{}
