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

import "github.com/binaek/sentra/tokens"

type RuleStatement struct {
	Pos      tokens.Position
	RuleName string
	Default  Expression
	When     Expression
	Body     Expression
}

func (r RuleStatement) String() string {
	return r.RuleName
}

func (r RuleStatement) Position() tokens.Position {
	return r.Pos
}

func (r RuleStatement) statementNode() {}
