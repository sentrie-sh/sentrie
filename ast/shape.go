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

type ShapeStatement struct {
	Pos     tokens.Position
	Name    string
	Simple  TypeRef
	Complex *Cmplx
}

type Cmplx struct {
	Pos    tokens.Position
	With   FQN // optional
	Node   Node
	Fields map[string]*ShapeField
}

type ShapeField struct {
	Pos         tokens.Position
	Name        string
	NotNullable bool // true if field is not nullable
	Optional    bool // true if field is optional
	Type        TypeRef
	Node        Node
}

func (s *ShapeStatement) Position() tokens.Position {
	return s.Pos
}

func (s *ShapeStatement) String() string {
	return ""
}

func (s *ShapeStatement) statementNode() {}
