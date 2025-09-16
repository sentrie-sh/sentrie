package index

import (
	"fmt"
	"sync/atomic"

	"github.com/binaek/sentra/ast"
	"github.com/binaek/sentra/xerr"
	"github.com/pkg/errors"
)

type Shape struct {
	Statement *ast.ShapeStatement
	Namespace *Namespace
	Policy    *Policy
	Name      string
	FQN       ast.FQN
	Complex   *Cmplx
	Simple    ast.TypeRef
	FilePath  string

	hydrated uint32 // 0 = not hydrated, 1 = hydrated
}

type Cmplx struct {
	WithFQN ast.FQN
	Fields  map[string]*ShapeField
}

type ExportedShape struct {
	Statement *ast.ShapeExportStatement
	Name      string
}

type ShapeField struct {
	Node        *ast.ShapeField
	Name        string
	NotNullable bool
	Optional    bool
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

	if s.Complex == nil {
		// nothing to do
		return nil
	}

	if len(s.Complex.WithFQN) == 0 {
		// nothing to do
		return nil
	}

	var withShape *Shape
	withName := s.Complex.WithFQN.LastSegment()

	// if we have a policy, look for it in the policy's shapes
	if inPolicy != nil {
		// check the policy's shapes
		if shape, ok := inPolicy.Shapes[withName]; ok {
			withShape = shape
			// hydrate the with shape
			if err := withShape.resolveDependency(idx, inPolicy); err != nil {
				return err
			}
		}
	}

	// check that we have the shape in the containing namespace
	if shape, ok := s.Namespace.Shapes[withName]; ok {
		withShape = shape
		// hydrate the with shape
		if err := withShape.resolveDependency(idx, nil); err != nil {
			return err
		}
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
						return errors.Wrapf(ErrIndex, "shape '%s' not exported at %s", withName, ns.Statement.Position())
					}
				}

				withShape = s
				break
			}
		}
	}

	// if by this point we don't have a shape, we need to error
	if withShape == nil {
		return errors.Wrapf(ErrIndex, "shape '%s' not found at %s", s.Complex.WithFQN.String(), s.Statement.Position())
	}

	// at this point we have the shape, we are going to assume it's hydrated
	// the assumption is not unfounded, since we traverse the shapes in a topological order

	// now we bring in the fields
	for name, field := range withShape.Complex.Fields {
		if _, ok := s.Complex.Fields[name]; ok {
			return errors.Wrapf(ErrIndex, "cannot compose with duplicate shape field '%s' at %s and %s", name, field.Node.Pos, s.Complex.Fields[name].Node.Pos)
		}
		s.Complex.Fields[name] = field
	}

	return nil
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
		FilePath:  stmt.Pos.Filename,
	}

	if stmt.Complex != nil {
		shape.Complex = &Cmplx{WithFQN: stmt.Complex.With, Fields: make(map[string]*ShapeField)}
		for _, field := range stmt.Complex.Fields {
			if field.Name == "" {
				continue
			}

			// if we already have the field, we need to error
			if _, ok := shape.Complex.Fields[field.Name]; ok {
				return nil, fmt.Errorf("duplicate shape field '%s' at %s", field.Name, field.Node.Position())
			}

			shape.Complex.Fields[field.Name] = &ShapeField{
				Node:        field,
				Name:        field.Name,
				NotNullable: field.NotNullable,
				Optional:    field.Optional,
				TypeRef:     field.Type,
			}
		}
	} else {
		shape.Simple = stmt.Simple
	}

	return shape, nil
}
