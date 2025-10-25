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

	"github.com/sentrie-sh/sentrie/tokens"
)

// FloatLiteral represents a float literal
type FloatLiteral struct {
	Range tokens.Range
	Value float64
}

func (f *FloatLiteral) String() string {
	return fmt.Sprintf("%g", f.Value)
}

func (f *FloatLiteral) expressionNode() {}

func (f *FloatLiteral) Kind() string {
	return "float_literal"
}

// Span returns the span of the float literal in the source code
func (f *FloatLiteral) Span() tokens.Range {
	return f.Range
}

var _ Expression = &FloatLiteral{}
var _ Node = &FloatLiteral{}
