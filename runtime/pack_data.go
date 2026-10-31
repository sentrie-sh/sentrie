// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"
	"fmt"
	"maps"
	"slices"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/index"
)

// packDataValue returns the value a `data` declaration reads: the entry the
// loader decoded once at pack load. The value is shared by every evaluation.
// Policy code has no operation that mutates a value, and every value handed
// to a TypeScript module is copied at the boundary (box.TryToBoundaryAny and
// normalizeBoundaryForJS), so nothing can change it.
func packDataValue(exec Executor, d *ast.DataStatement) (box.Value, error) {
	idx := exec.Index()
	if idx != nil && idx.Pack != nil {
		if v, ok := idx.Pack.DataValues[d.Name]; ok {
			return v, nil
		}
	}
	return box.Undefined(), fmt.Errorf("pack data %q is not loaded at %s", d.Name, d.Span())
}

// validatePackData checks every typed `data <name>: T` declaration in the
// index against its loaded value, once, when the executor is built. Data does
// not change between evaluations, so a mismatch is a pack error, reported
// before any request runs rather than on every evaluation that reads it.
// Constraint arguments are literals, so no lets or facts are bound here.
func (e *executorImpl) validatePackData(ctx context.Context) error {
	if e.index == nil {
		return nil
	}
	for _, nsName := range slices.Sorted(maps.Keys(e.index.Namespaces)) {
		ns := e.index.Namespaces[nsName]
		for _, polName := range slices.Sorted(maps.Keys(ns.Policies)) {
			p := ns.Policies[polName]
			for _, alias := range slices.Sorted(maps.Keys(p.Data)) {
				d := p.Data[alias]
				if d.Type == nil {
					continue
				}
				if err := e.validateDataDeclaration(ctx, p, d); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (e *executorImpl) validateDataDeclaration(ctx context.Context, p *index.Policy, d *ast.DataStatement) error {
	v, err := packDataValue(e, d)
	if err != nil {
		return err
	}
	ec := NewExecutionContext(p, e)
	defer ec.Dispose()
	if err := validateValueAgainstTypeRef(ctx, ec, e, p, v, d.Type, d.Span()); err != nil {
		return fmt.Errorf("pack data %q in policy %s does not match its declared type: %w", d.Name, p.FQN, err)
	}
	return nil
}
