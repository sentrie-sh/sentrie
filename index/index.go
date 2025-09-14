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
	"context"
	"sync"

	"github.com/binaek/sentra/ast"
	"github.com/binaek/sentra/pack"
)

type Index struct {
	theLock    *sync.RWMutex
	Pack       *pack.PackFile
	Namespaces map[string]*Namespace
	Programs   map[string]*Program
}

func CreateIndex() *Index {
	return &Index{
		theLock:    &sync.RWMutex{},
		Namespaces: make(map[string]*Namespace),
		Programs:   make(map[string]*Program),
	}
}

func (idx *Index) SetPack(ctx context.Context, p *pack.PackFile) error {
	idx.theLock.Lock()
	defer idx.theLock.Unlock()

	idx.Pack = p
	return nil
}

func (idx *Index) AddProgram(ctx context.Context, astProgram *ast.Program) error {
	idx.theLock.Lock()
	defer idx.theLock.Unlock()

	// bail out if the context is done
	if ctx.Err() != nil {
		return ctx.Err()
	}

	program := createProgram(astProgram)

	ns := idx.ensureNamespace(ctx, program.Namespace)

	for _, shape := range program.Shapes {
		shape, err := createShape(ns, nil, shape)
		if err != nil {
			return err
		}

		if err := ns.addShape(shape); err != nil {
			return err
		}
	}

	for _, policy := range program.Policies {
		p, err := createPolicy(ns, policy, astProgram)
		if err != nil {
			return err
		}

		if err := ns.addPolicy(p); err != nil {
			return err
		}
	}

	for _, export := range program.ShapeExports {
		if err := ns.addShapeExport(&ExportedShape{Name: export.Name, Statement: export}); err != nil {
			return err
		}
	}

	idx.Programs[astProgram.Reference] = program

	return nil
}

func (idx *Index) ensureNamespace(_ context.Context, namespace *ast.NamespaceStatement) *Namespace {
	if ns, ok := idx.Namespaces[namespace.String()]; ok {
		return ns
	}

	theNew := createNamespace(namespace)

	// now iterate through all known namespaces and resolve the parent/child relationships
	for _, indexed := range idx.Namespaces {
		if theNew.IsChildOf(indexed) {
			theNew.Parent = indexed
			indexed.Children = append(indexed.Children, theNew)
		}

		if theNew.IsParentOf(indexed) {
			indexed.Parent = theNew
			theNew.Children = append(theNew.Children, indexed)
		}
	}

	idx.Namespaces[namespace.String()] = theNew

	return theNew
}
