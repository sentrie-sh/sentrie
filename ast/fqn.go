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
)

const FQNSeparator = "/"

type FQN []string

func (f FQN) String() string {
	if len(f) == 0 {
		return ""
	}
	return strings.Join(f, FQNSeparator)
}

func CreateFQN(base FQN, lastSegment string) FQN {
	if len(base) == 0 {
		return FQN{lastSegment}
	}
	return append(base, lastSegment)
}

func (f FQN) IsChildOf(another FQN) bool {
	// ["com","example","foo"] child of ["com","example"]
	// ["com","example","foo"] not child of ["com","example","bar"]
	// ["com","example","foo"] not child of ["com","example2","foo"]
	// ["com","example","foo","bar"] not child of ["com","example"]
	if len(f)-1 != len(another) {
		return false
	}

	fqn := f.String()
	supposedToBeChildFQN := another.String()
	return strings.HasPrefix(fqn, fmt.Sprintf("%s%s", supposedToBeChildFQN, FQNSeparator))
}

func (f FQN) IsParentOf(another FQN) bool {
	return another.IsChildOf(f)
}
