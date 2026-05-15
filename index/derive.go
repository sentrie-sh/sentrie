// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
	"github.com/sentrie-sh/sentrie/xerr"
)

// Derive is an indexed pure function (namespace- or policy-scoped).
type Derive struct {
	Name      string
	FQN       ast.FQN
	Lambda    *ast.LambdaExpression
	Namespace *Namespace
	Policy    *Policy // nil when declared at namespace scope
	Statement ast.Statement

	// DefineShort and DefineFQN are bind-time snapshots (immutable after indexing): they record
	// which derives were visible when this derive was registered. Later programs that add derives
	// to the same namespace or policy do not retroactively update older entries; load dependent
	// programs after their helpers when using AddProgram, or keep helpers in the same file.
	DefineShort map[string]*Derive
	DefineFQN map[string]*Derive
}

func (d *Derive) String() string {
	return d.FQN.String()
}

func (d *Derive) Span() tokens.Range {
	if d.Statement != nil {
		return d.Statement.Span()
	}
	return d.Lambda.Span()
}

// ExportedDerive records a namespace-level export of a derive.
type ExportedDerive struct {
	Name      string
	Statement *ast.ExportDeriveStatement
}

func cloneDeriveMap(src map[string]*Derive) map[string]*Derive {
	out := make(map[string]*Derive, len(src)+1)
	for k, v := range src {
		out[k] = v
	}
	return out
}

func overlayDeriveMaps(base map[string]*Derive, over map[string]*Derive) map[string]*Derive {
	out := cloneDeriveMap(base)
	for k, v := range over {
		out[k] = v
	}
	return out
}

func buildDefineFQN(m map[string]*Derive) map[string]*Derive {
	out := make(map[string]*Derive, len(m))
	for _, d := range m {
		out[d.FQN.String()] = d
	}
	return out
}

func newDerive(name string, lam *ast.LambdaExpression, ns *Namespace, pol *Policy, stmt ast.Statement, visible map[string]*Derive) *Derive {
	var fqn ast.FQN
	if pol != nil {
		fqn = ast.CreateFQN(pol.FQN, name)
	} else {
		fqn = ast.CreateFQN(ns.FQN, name)
	}
	def := cloneDeriveMap(visible)
	return &Derive{
		Name:        name,
		FQN:         fqn,
		Lambda:      lam,
		Namespace:   ns,
		Policy:      pol,
		Statement:   stmt,
		DefineShort: def,
		DefineFQN:   buildDefineFQN(def),
	}
}

func (n *Namespace) addDerive(idx *Index, stmt *ast.DeriveStatement, visibleBefore map[string]*Derive) (*Derive, error) {
	name := stmt.Name
	if err := n.checkNameAvailable(name); err != nil {
		return nil, err
	}
	if _, ok := n.Derives[name]; ok {
		return nil, xerr.ErrConflict("derive declaration", stmt.Span(), n.Derives[name].Span())
	}
	d := newDerive(name, stmt.Value, n, nil, stmt, visibleBefore)
	if idx != nil {
		if err := idx.registerDeriveFQN(d); err != nil {
			return nil, err
		}
	}
	n.Derives[name] = d
	return d, nil
}

func (p *Policy) addDerive(idx *Index, stmt *ast.DeriveStatement, nsSnapshot map[string]*Derive, policySoFar map[string]*Derive) (*Derive, error) {
	name := stmt.Name
	if seen, ok := p.seenIdentifiers[name]; ok {
		return nil, xerr.ErrConflict("derive declaration", stmt.Span(), seen.Span())
	}
	if _, ok := p.Derives[name]; ok {
		return nil, xerr.ErrConflict("derive declaration", stmt.Span(), p.Derives[name].Span())
	}
	visible := overlayDeriveMaps(nsSnapshot, policySoFar)
	d := newDerive(name, stmt.Value, p.Namespace, p, stmt, visible)
	if idx != nil {
		if err := idx.registerDeriveFQN(d); err != nil {
			return nil, err
		}
	}
	p.Derives[name] = d
	p.seenIdentifiers[name] = stmt
	return d, nil
}

func (idx *Index) registerDeriveFQN(d *Derive) error {
	key := d.FQN.String()
	if other, ok := idx.DerivesByFQN[key]; ok {
		return xerr.ErrConflict("derive declaration", d.Span(), other.Span())
	}
	idx.DerivesByFQN[key] = d
	return nil
}
