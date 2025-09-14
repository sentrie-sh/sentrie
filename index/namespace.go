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
	"github.com/binaek/sentra/ast"
	"github.com/pkg/errors"
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
	if _, ok := n.Policies[policy.Name]; ok {
		return errors.Wrapf(ErrIndex, "policy name conflict: '%s' at %s", policy.Name, policy.Statement.Position())
	}

	n.Policies[policy.Name] = policy
	return nil
}

func (n *Namespace) addShape(shape *Shape) error {
	if _, ok := n.Shapes[shape.Name]; ok {
		return errors.Wrapf(ErrIndex, "shape name conflict: '%s' at %s", shape.Name, shape.Statement.Position())
	}

	n.Shapes[shape.Name] = shape
	return nil
}

func (n *Namespace) addShapeExport(export *ExportedShape) error {
	if _, ok := n.ShapeExports[export.Name]; ok {
		return errors.Wrapf(ErrIndex, "shape export conflict: '%s' at %s", export.Name, export.Statement.Position())
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
