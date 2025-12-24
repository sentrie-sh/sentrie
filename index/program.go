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

package index

import "github.com/sentrie-sh/sentrie/ast"

type Program struct {
	Reference    *ast.Program
	Namespace    *ast.NamespaceStatement
	Policies     []*ast.PolicyStatement
	Shapes       []*ast.ShapeStatement
	ShapeExports []*ast.ShapeExportStatement
}

func createProgram(astProgram *ast.Program) *Program {
	p := &Program{
		Reference:    astProgram,
		Namespace:    nil,
		Policies:     make([]*ast.PolicyStatement, 0),
		Shapes:       make([]*ast.ShapeStatement, 0),
		ShapeExports: make([]*ast.ShapeExportStatement, 0),
	}

	for _, stmt := range astProgram.Statements {
		switch stmt := stmt.(type) {
		case *ast.NamespaceStatement:
			p.Namespace = stmt
		case *ast.PolicyStatement:
			p.Policies = append(p.Policies, stmt)
		case *ast.ShapeStatement:
			p.Shapes = append(p.Shapes, stmt)
		case *ast.ShapeExportStatement:
			p.ShapeExports = append(p.ShapeExports, stmt)
		}
	}

	return p
}
