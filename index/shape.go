// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"fmt"
	"slices"
	"strings"
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

// resolveDependency copies the fields of the shape s composes with into s.
// The base comes from resolveWithBase, the same resolution the shape cycle
// check uses, so hydration and the cycle check agree on every parent.
func (s *Shape) resolveDependency(idx *Index) error {
	if atomic.LoadUint32(&s.hydrated) == 1 {
		return nil
	}

	defer func() {
		atomic.StoreUint32(&s.hydrated, 1)
	}()

	withShape, err := idx.resolveWithBase(s)
	if err != nil {
		return err
	}
	if withShape == nil {
		// no model or no with clause: nothing to do
		return nil
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

// resolveWithBase returns the shape s composes with through its `with`
// clause, or nil and no error when s has no model or no `with` clause. It is
// the one resolution both hydration (resolveDependency) and the shape cycle
// check (detectShapeCycle) use (#307, #308).
//
// A qualified base (two or more segments) resolves only in the namespace its
// leading segments name, which must export it, unless that namespace is the
// shape's own, which is exempt from the export check. An unqualified base
// resolves in the shape's policy, then its namespace, and the first match
// wins. An unqualified base that only another namespace's exports hold is an
// error that asks for the qualified name.
func (idx *Index) resolveWithBase(s *Shape) (*Shape, error) {
	if s.Model == nil || s.Model.WithFQN == nil || s.Model.WithFQN.IsEmpty() {
		return nil, nil
	}
	ref := s.Model.WithFQN
	at := s.Statement.Span()
	name := ref.LastSegment()

	if len(ref.Parts) > 1 {
		target, ok := idx.Namespaces[ref.Parent().String()]
		if !ok || target == nil {
			return nil, fmt.Errorf("shape '%s' not found at %s: %w", ref.String(), at, xerr.ErrIndex)
		}
		base, ok := target.Shapes[name]
		if !ok {
			return nil, fmt.Errorf("shape '%s' not found at %s: %w", ref.String(), at, xerr.ErrIndex)
		}
		if target != s.Namespace {
			if err := target.VerifyShapeExported(name); err != nil {
				return nil, fmt.Errorf("shape '%s' not exported at %s: %w", ref.String(), at, xerr.ErrIndex)
			}
		}
		return base, nil
	}

	if s.Policy != nil {
		if base, ok := s.Policy.Shapes[name]; ok {
			return base, nil
		}
	}
	if s.Namespace != nil {
		if base, ok := s.Namespace.Shapes[name]; ok {
			return base, nil
		}
	}

	// Only another namespace's exports can still hold the name. Those are
	// reachable by their qualified name alone, so name them in the error.
	var qualified []string
	for nsName, ns := range idx.Namespaces {
		if ns == nil || ns == s.Namespace {
			continue
		}
		if _, ok := ns.Shapes[name]; !ok {
			continue
		}
		if ns.VerifyShapeExported(name) == nil {
			qualified = append(qualified, ShapeFQN(nsName, name))
		}
	}
	if len(qualified) > 0 {
		slices.Sort(qualified)
		return nil, fmt.Errorf("shape '%s' at %s is exported by another namespace and must be qualified: use %s: %w", ref.String(), at, strings.Join(qualified, " or "), xerr.ErrIndex)
	}
	return nil, fmt.Errorf("shape '%s' not found at %s: %w", ref.String(), at, xerr.ErrIndex)
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
