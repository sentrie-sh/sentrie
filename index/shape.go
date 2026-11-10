// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"errors"
	"fmt"
	"sync/atomic"

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

	shapeScope *Policy // see ShapeScope
	hydrated   uint32  // 0 = not hydrated, 1 = hydrated
}

type ShapeModel struct {
	WithFQN *ast.FQN
	Fields  map[string]*ShapeModelField
	// Open is true when the shape's own body ends in `...`: a value may then
	// carry fields the shape does not declare. It comes from this shape's
	// declaration only. A `with` base being open does not open the shape
	// that composes it, so a composed shape is closed unless it writes `...`
	// itself.
	Open bool
}

type ExportedShape struct {
	Statement *ast.ShapeExportStatement
	Name      string
}

type ShapeModelField struct {
	Node     *ast.ShapeField
	Name     string
	Optional bool
	TypeRef  ast.TypeRef

	// DeclaredIn is the shape whose declaration holds the field. A field a
	// shape takes in through `with` keeps the shape it came from, so its type
	// resolves where it was written. Nil only for a hand-built model.
	DeclaredIn *Shape
}

func (s *Shape) String() string {
	return s.FQN.String()
}

// ShapeScope is the policy whose shapes the shape's alias and field types
// resolve against: the shape's own policy, or for a namespace-level shape a
// Policy that holds only the shape's namespace (no policy shapes, no rules).
// It is built once when the shape is indexed and never written afterwards, so
// concurrent evaluations read it without a lock. It is nil for a hand-built
// shape with neither a policy nor a namespace.
func (s *Shape) ShapeScope() *Policy {
	if s.shapeScope != nil {
		return s.shapeScope
	}
	if s.Policy != nil {
		return s.Policy
	}
	if s.Namespace != nil {
		return &Policy{Namespace: s.Namespace}
	}
	return nil
}

// TypeScope is the policy whose shapes the field's type resolves against: the
// declaring shape's scope, or fallback when the declaring shape has none.
func (f *ShapeModelField) TypeScope(fallback *Policy) *Policy {
	if f.DeclaredIn != nil {
		if scope := f.DeclaredIn.ShapeScope(); scope != nil {
			return scope
		}
	}
	return fallback
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
			if err != nil {
				if isShapeDependencyNamespaceMiss(err) {
					continue
				}
				return err
			}

			if s != nil {
				if ns.FQN.String() != s.Namespace.FQN.String() {
					// we have the shape, but we need to verify it's exported
					if err := ns.VerifyShapeExported(withName); err != nil {
						return fmt.Errorf("shape '%s' not exported at %s: %w", withName, ns.Statement.Span(), xerr.ErrIndex)
					}
				}

				withShape = s
				break
			}
		}
	}

	// if by this point we don't have a shape, we need to error
	if withShape == nil {
		return fmt.Errorf("shape '%s' not found at %s: %w", s.Model.WithFQN.String(), s.Statement.Span(), xerr.ErrIndex)
	}

	if withShape.AliasOf != nil {
		return fmt.Errorf("cannot compose '%s' with alias of shape '%s' at %s: %w", s.FQN.String(), withShape.FQN.String(), withShape.Statement.Span(), xerr.ErrIndex)
	}

	// at this point we have the shape, we are going to assume it's hydrated
	// the assumption is not unfounded, since we traverse the shapes in a topological order

	// now we bring in the fields
	for name, field := range withShape.Model.Fields {
		if _, ok := s.Model.Fields[name]; ok {
			return fmt.Errorf("cannot compose with duplicate shape field '%s' at %s and %s: %w", name, field.Node.Range, s.Model.Fields[name].Node.Range, xerr.ErrIndex)
		}
		s.Model.Fields[name] = field
	}

	return nil
}

func isShapeDependencyNamespaceMiss(err error) bool {
	var notFoundErr xerr.NotFoundError
	return errors.As(err, &notFoundErr)
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
	shape.shapeScope = p
	if shape.shapeScope == nil {
		shape.shapeScope = &Policy{Namespace: ns}
	}

	if stmt.Complex != nil {
		shape.Model = &ShapeModel{Fields: make(map[string]*ShapeModelField)}
		if stmt.Complex.With != nil {
			shape.Model.WithFQN = stmt.Complex.With
		}
		shape.Model.Open = stmt.Complex.Open
		for _, field := range stmt.Complex.Fields {
			if field.Name == "" {
				continue
			}

			// if we already have the field, we need to error
			if _, ok := shape.Model.Fields[field.Name]; ok {
				return nil, fmt.Errorf("duplicate shape field '%s' at %s", field.Name, field.Range)
			}

			shape.Model.Fields[field.Name] = &ShapeModelField{
				Node:       field,
				Name:       field.Name,
				Optional:   field.Optional,
				TypeRef:    field.Type,
				DeclaredIn: shape,
			}
		}
	} else {
		shape.AliasOf = stmt.Simple
	}

	return shape, nil
}
