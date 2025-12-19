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

package index

import (
	"fmt"
	"sync/atomic"

	"github.com/pkg/errors"
	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
	"github.com/sentrie-sh/sentrie/xerr"
)

type Shape struct {
	Statement *ast.ShapeStatement
	Namespace *Namespace
	Policy    *Policy
	Name      string
	FQN       ast.FQN
	Model     *ShapeModel
	AliasOf   ast.TypeRef
	FilePath  string

	hydrated uint32 // 0 = not hydrated, 1 = hydrated
}

type ShapeModel struct {
	WithFQN *ast.FQN
	Fields  map[string]*ShapeModelField
}

type ExportedShape struct {
	Statement *ast.ShapeExportStatement
	Name      string
}

type ShapeModelField struct {
	Node        *ast.ShapeField
	Name        string
	NotNullable bool
	Required    bool
	TypeRef     ast.TypeRef
}

func (s *Shape) String() string {
	return s.FQN.String()
}

func (s *Shape) resolveDependency(idx *Index, inPolicy *Policy) error {
	if atomic.LoadUint32(&s.hydrated) == 1 {
		return nil
	}

	defer func() {
		atomic.StoreUint32(&s.hydrated, 1)
	}()

	if s.Model == nil {
		// nothing to do
		return nil
	}

	if s.Model.WithFQN == nil || s.Model.WithFQN.IsEmpty() {
		// nothing to do
		return nil
	}

	var withShape *Shape
	withName := s.Model.WithFQN.LastSegment()

	// if we have a policy, look for it in the policy's shapes
	if inPolicy != nil {
		// check the policy's shapes
		if shape, ok := inPolicy.Shapes[withName]; ok {
			withShape = shape
		}
	}

	// check if we have the shape in the containing namespace
	if shape, ok := s.Namespace.Shapes[withName]; ok {
		withShape = shape
	}

	if withShape == nil {
		// now we need to check whether this is exported by some other namespaces in the index
		for _, ns := range idx.Namespaces {
			// check in exported shapes
			s, err := idx.ResolveShape(ns.FQN.String(), withName)
			if errors.Is(err, xerr.ErrShapeNotFound(withName)) {
				continue
			}

			if s != nil {
				if ns.FQN.String() != s.Namespace.FQN.String() {
					// we have the shape, but we need to verify it's exported
					if err := ns.VerifyShapeExported(withName); err != nil {
						return errors.Wrapf(ErrIndex, "shape '%s' not exported at %s", withName, ns.Statement.Span())
					}
				}

				withShape = s
				break
			}
		}
	}

	// if by this point we don't have a shape, we need to error
	if withShape == nil {
		return errors.Wrapf(ErrIndex, "shape '%s' not found at %s", s.Model.WithFQN.String(), s.Statement.Span())
	}

	if withShape.AliasOf != nil {
		return errors.Wrapf(ErrIndex, "cannot compose '%s' with alias of shape '%s' at %s", s.FQN.String(), withShape.FQN.String(), withShape.Statement.Span())
	}

	// at this point we have the shape, we are going to assume it's hydrated
	// the assumption is not unfounded, since we traverse the shapes in a topological order

	// now we bring in the fields
	for name, field := range withShape.Model.Fields {
		if _, ok := s.Model.Fields[name]; ok {
			return errors.Wrapf(ErrIndex, "cannot compose with duplicate shape field '%s' at %s and %s", name, field.Node.Range, s.Model.Fields[name].Node.Range)
		}
		s.Model.Fields[name] = field
	}

	return nil
}

func (s *Shape) Span() tokens.Range {
	return s.Statement.Span()
}

func createShape(ns *Namespace, p *Policy, stmt *ast.ShapeStatement) (*Shape, error) {
	var fqn ast.FQN
	if p != nil {
		fqn = ast.CreateFQN(p.FQN, stmt.Name)
	} else {
		fqn = ast.CreateFQN(ns.FQN, stmt.Name)
	}
	shape := &Shape{
		Statement: stmt,
		Namespace: ns,
		Policy:    p,
		Name:      stmt.Name,
		FQN:       fqn,
		FilePath:  stmt.Rnge.File,
	}

	if stmt.Complex != nil {
		shape.Model = &ShapeModel{Fields: make(map[string]*ShapeModelField)}
		if stmt.Complex.With != nil {
			shape.Model.WithFQN = stmt.Complex.With
		}
		for _, field := range stmt.Complex.Fields {
			if field.Name == "" {
				continue
			}

			// if we already have the field, we need to error
			if _, ok := shape.Model.Fields[field.Name]; ok {
				return nil, fmt.Errorf("duplicate shape field '%s' at %s", field.Name, field.Range)
			}

			shape.Model.Fields[field.Name] = &ShapeModelField{
				Node:        field,
				Name:        field.Name,
				NotNullable: field.NotNullable,
				Required:    field.Required,
				TypeRef:     field.Type,
			}
		}
	} else {
		shape.AliasOf = stmt.Simple
	}

	return shape, nil
}
