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
	"path/filepath"
	"strings"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/xerr"
)

func (idx *Index) ResolveNamespace(ns string) (*Namespace, error) {
	n := idx.Namespaces[ns]
	if n == nil {
		return nil, xerr.ErrNamespaceNotFound(ns)
	}
	return n, nil
}

// ResolvePolicy tries exact namespace match; it does not traverse parents.
func (idx *Index) ResolvePolicy(ns, policy string) (*Policy, error) {
	n, err := idx.ResolveNamespace(ns)
	if err != nil {
		return nil, err
	}
	if n == nil {
		return nil, xerr.ErrNamespaceNotFound(ns)
	}
	p := n.Policies[policy]
	if p == nil {
		return nil, xerr.ErrPolicyNotFound(filepath.Join(ns, policy))
	}
	return p, nil
}

func (idx *Index) ResolveShape(ns, shape string) (*Shape, error) {
	n, err := idx.ResolveNamespace(ns)
	if err != nil {
		return nil, err
	}
	if n == nil {
		return nil, xerr.ErrNamespaceNotFound(ns)
	}
	s, ok := n.Shapes[shape]
	if !ok {
		return nil, xerr.ErrShapeNotFound(filepath.Join(ns, shape))
	}
	return s, nil
}

// VerifyRuleExported verifies that a rule is exported in its policy. Returns an error if the rule is not exported.
func (p Policy) VerifyRuleExported(rule string) error {
	if _, ok := p.RuleExports[rule]; !ok {
		return xerr.ErrNotExported(RuleFQN(p.Namespace.FQN.String(), p.Name, rule))
	}
	return nil
}

func (ns Namespace) VerifyShapeExported(shape string) error {
	if _, ok := ns.ShapeExports[shape]; !ok {
		return xerr.ErrNotExported(ShapeFQN(ns.FQN.String(), shape))
	}
	return nil
}

// FQN utilities
func RuleFQN(ns, policy, rule string) string {
	return strings.Join([]string{ns, policy, rule}, ast.FQNSeparator)
}

func ShapeFQN(ns, shape string) string {
	return strings.Join([]string{ns, shape}, ast.FQNSeparator)
}
