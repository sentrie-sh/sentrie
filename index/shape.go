package index

import (
	"fmt"
	"sync/atomic"

	"github.com/binaek/sentra/ast"
	"github.com/pkg/errors"
)

type Shape struct {
	Node      *ast.ShapeStatement
	Namespace *Namespace
	Policy    *Policy
	Name      string
	FQN       ast.FQN
	Complex   *Cmplx
	Simple    ast.TypeRef
	FilePath  string

	withHydrated uint32 // 0 = not hydrated, 1 = hydrated
}

type Cmplx struct {
	WithFQN ast.FQN
	Fields  map[string]*ShapeField
}

type ExportedShape struct {
	Node *ast.ShapeExportStatement
	Name string
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

func (s *Shape) HydrateShapeWith(idx *Index, inPolicy *Policy) error {
	if atomic.LoadUint32(&s.withHydrated) == 1 {
		return nil
	}

	defer func() {
		atomic.StoreUint32(&s.withHydrated, 1)
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

	// if we have a policy, look for it in the policy's shapes
	if inPolicy != nil {
		// check the policy's shapes
		if shape, ok := inPolicy.Shapes[s.Name]; ok {
			withShape = shape

			// hydrate the with shape
			if err := withShape.HydrateShapeWith(idx, inPolicy); err != nil {
				return err
			}
		}

		// check the policy's namespace's shapes
		if shape, ok := inPolicy.Namespace.Shapes[s.Name]; ok {
			withShape = shape
			// hydrate the with shape
			if err := withShape.HydrateShapeWith(idx, nil); err != nil {
				return err
			}
		}
	}

	// now we need to check whether this is exported by some other namespaces in the index
	for _, indexed := range idx.Namespaces {
		if shape, ok := indexed.Shapes[s.Complex.WithFQN.String()]; ok {
			withShape = shape

			// hydrate the with shape
			if err := withShape.HydrateShapeWith(idx, nil); err != nil {
				return err
			}
		}
	}

	// if by this point we don't have a shape, we need to error
	if withShape == nil {
		return errors.Wrapf(ErrIndex, "shape '%s' not found", s.Complex.WithFQN.String())
	}

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
		Node:      stmt,
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
