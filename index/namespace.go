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
	"github.com/pkg/errors"
	"github.com/sentrie-sh/sentrie/ast"
)

// Namespace is an index of policies and shapes visible within (namespace & sub-namespaces).
type Namespace struct {
	Statement    *ast.NamespaceStatement
	FQN          ast.FQN // this is always the FQN
	Parent       *Namespace
	Children     []*Namespace
	Policies     map[string]*Policy
	Shapes       map[string]*Shape // namespace-level shapes
	ShapeExports map[string]*ExportedShape
}

func (ns *Namespace) addChild(child *Namespace) error {
	baseName := child.FQN.LastSegment()

	if err := ns.checkNameAvailable(baseName); err != nil {
		return err
	}

	ns.Children = append(ns.Children, child)
	child.Parent = ns
	return nil
}

func (ns *Namespace) checkNameAvailable(name string) error {
	if _, ok := ns.Policies[name]; ok {
		return errors.Wrapf(ErrIndex, "name conflict: '%s' at %s", name, ns.Statement.Span())
	}
	if _, ok := ns.Shapes[name]; ok {
		return errors.Wrapf(ErrIndex, "name conflict: '%s' at %s", name, ns.Statement.Span())
	}
	// there shouldn't be a child namespace
	for _, child := range ns.Children {
		cName := child.FQN.LastSegment()
		if cName == name {
			return errors.Wrapf(ErrIndex, "namespace conflict: '%s' at %s", cName, ns.Statement.Span())
		}
	}
	return nil
}

func createNamespace(node *ast.NamespaceStatement) *Namespace {
	return &Namespace{
		Statement:    node,
		FQN:          node.Name,
		Parent:       nil,
		Children:     make([]*Namespace, 0),
		Policies:     make(map[string]*Policy),
		Shapes:       make(map[string]*Shape),
		ShapeExports: make(map[string]*ExportedShape),
	}
}

func (n *Namespace) addPolicy(policy *Policy) error {
	baseName := policy.FQN.LastSegment()
	if err := n.checkNameAvailable(baseName); err != nil {
		return err
	}

	if _, ok := n.Policies[policy.Name]; ok {
		return errors.Wrapf(ErrIndex, "policy name conflict: '%s' at %s", policy.Name, policy.Statement.Span())
	}

	n.Policies[policy.Name] = policy
	return nil
}

func (n *Namespace) addShape(shape *Shape) error {
	baseName := shape.FQN.LastSegment()
	if err := n.checkNameAvailable(baseName); err != nil {
		return err
	}

	if _, ok := n.Shapes[shape.Name]; ok {
		return errors.Wrapf(ErrIndex, "shape name conflict: '%s' at %s", shape.Name, shape.Statement.Span())
	}

	n.Shapes[shape.Name] = shape
	return nil
}

func (n *Namespace) addShapeExport(export *ExportedShape) error {
	if _, ok := n.ShapeExports[export.Name]; ok {
		return errors.Wrapf(ErrIndex, "shape export conflict: '%s' at %s", export.Name, export.Statement.Span())
	}

	n.ShapeExports[export.Name] = export
	return nil
}

func (n Namespace) IsChildOf(another *Namespace) bool {
	return n.FQN.IsChildOf(another.FQN)
}

func (n Namespace) IsParentOf(another *Namespace) bool {
	return n.FQN.IsParentOf(another.FQN)
}
