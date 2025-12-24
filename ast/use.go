// SPDX-License-Identifier: Apache-2.0
//
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

type UseStatement struct {
	*baseNode
	Modules      []string // List of modules to use
	RelativeFrom string   //
	LibFrom      []string // Optional library information
	As           string
}

func NewUseStatement(modules []string, relativeFrom string, libFrom []string, as string, ssp tokens.Range) *UseStatement {
	return &UseStatement{
		baseNode: &baseNode{
			Rnge:  ssp,
			Kind_: "use",
		},
		Modules:      modules,
		RelativeFrom: relativeFrom,
		LibFrom:      libFrom,
		As:           as,
	}
}

func (s *UseStatement) String() string {
	from := s.RelativeFrom
	if len(s.LibFrom) > 0 {
		from = "@" + strings.Join(s.LibFrom, "/")
	}
	return fmt.Sprintf("use %s from %s as %s", strings.Join(s.Modules, ", "), from, s.As)
}

func (s *UseStatement) statementNode() {}

var _ Statement = &UseStatement{}
var _ Node = &UseStatement{}
