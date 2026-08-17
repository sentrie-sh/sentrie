// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0
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
	"fmt"
	"sync"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/dag"
	"github.com/sentrie-sh/sentrie/pack"
)

type Index struct {
	theLock    *sync.RWMutex
	Pack       *pack.PackFile
	Namespaces map[string]*Namespace
	Programs   map[string]*Program

	DerivesByFQN map[string]*Derive

	ruleDag  dag.G[*Rule]
	shapeDag dag.G[*Shape]

	validated       uint32 // 0 = not validated, 1 = validated
	validationError error
	validationOnce  *sync.Once

	committed   uint32 // 0 = not committed, 1 = committed
	commitError error
	commitOnce  *sync.Once
}

func CreateIndex() *Index {
	return &Index{
		theLock:        &sync.RWMutex{},
		Namespaces:     make(map[string]*Namespace),
		Programs:       make(map[string]*Program),
		DerivesByFQN:   make(map[string]*Derive),
		validated:      0,
		validationOnce: &sync.Once{},
		committed:      0,
		commitOnce:     &sync.Once{},
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

	var extraNamespace ast.Statement
	nsCount := 0
	for _, stmt := range astProgram.Statements {
		if _, ok := stmt.(*ast.NamespaceStatement); !ok {
			continue
		}
		nsCount++
		if nsCount > 1 {
			extraNamespace = stmt
			break
		}
	}
	if extraNamespace != nil {
		return fmt.Errorf("duplicate namespace statement at %s", extraNamespace.Span())
	}

	program := createProgram(astProgram)

	ns, err := idx.ensureNamespace(ctx, program.Namespace)
	if err != nil {
		return err
	}

	for _, stmt := range astProgram.Statements {
		switch s := stmt.(type) {
		case *ast.CommentStatement:
			continue

		case *ast.NamespaceStatement:
			// The first (and only) namespace is consumed by createProgram/ensureNamespace.
			continue

		case *ast.ShapeStatement:
			shape, err := createShape(ns, nil, s)
			if err != nil {
				return err
			}
			if err := ns.addShape(shape); err != nil {
				return err
			}

		case *ast.PolicyStatement:
			p, err := createPolicy(ns, s, astProgram, idx, cloneDeriveMap(ns.Derives))
			if err != nil {
				return err
			}
			if err := ns.addPolicy(p); err != nil {
				return err
			}

		case *ast.DeriveStatement:
			if _, err := ns.addDerive(idx, s, cloneDeriveMap(ns.Derives)); err != nil {
				return err
			}

		case *ast.ShapeExportStatement:
			if err := ns.addShapeExport(&ExportedShape{Name: s.Name, Statement: s}); err != nil {
				return err
			}

		case *ast.ExportDeriveStatement:
			if _, ok := ns.Derives[s.Name]; !ok {
				return fmt.Errorf("cannot export unknown derive %q at %s", s.Name, s.Span())
			}
			if err := ns.addDeriveExport(&ExportedDerive{Name: s.Name, Statement: s}); err != nil {
				return err
			}

		default:
			return fmt.Errorf("unsupported top-level statement %T at %s", stmt, stmt.Span())
		}
	}

	idx.Programs[astProgram.Reference] = program

	return nil
}

func (idx *Index) ensureNamespace(_ context.Context, namespace *ast.NamespaceStatement) (*Namespace, error) {
	if ns, ok := idx.Namespaces[namespace.String()]; ok {
		return ns, nil
	}

	theNew := createNamespace(namespace)

	// now iterate through all known namespaces and resolve the parent/child relationships
	for _, indexed := range idx.Namespaces {
		if theNew.IsChildOf(indexed) {
			if err := indexed.addChild(theNew); err != nil {
				return nil, err
			}
		}

		if theNew.IsParentOf(indexed) {
			if err := theNew.addChild(indexed); err != nil {
				return nil, err
			}
		}
	}

	idx.Namespaces[namespace.String()] = theNew

	return theNew, nil
}
