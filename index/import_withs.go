// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/xerr"
)

// checkImportWiths reports every `with` clause of an `import decision` whose
// name is not a fact of the imported policy (#151). An import is only parsed
// as a rule body, so rule.Body is the only place to look, the same place
// detectRuleCycle reads. It runs after detectRuleCycle, which has already
// resolved every imported policy.
func (idx *Index) checkImportWiths(ctx context.Context) error {
	var errs []xerr.ImportUnknownFactError
	for _, ns := range idx.Namespaces {
		for _, policy := range ns.Policies {
			if ctx.Err() != nil {
				return fmt.Errorf("validation cancelled: %w", xerr.ErrIndex)
			}
			for _, rule := range policy.Rules {
				imp, ok := rule.Body.(*ast.ImportClause)
				if !ok || imp == nil || len(imp.Withs) == 0 {
					continue
				}
				nsName, polName := importTarget(policy, imp)
				target, err := idx.ResolvePolicy(nsName, polName)
				if err != nil {
					return fmt.Errorf("error resolving policy: %s: %w", err, xerr.ErrIndex)
				}
				for _, w := range imp.Withs {
					if w == nil {
						continue
					}
					if _, ok := target.Facts[w.Name]; ok {
						continue
					}
					errs = append(errs, xerr.ImportUnknownFactError{
						At:      w.Span(),
						Message: xerr.ImportUnknownFactMessage(w.Name, nsName+ast.FQNSeparator+polName, target.FactNamesByAlias()),
					})
				}
			}
		}
	}

	slices.SortStableFunc(errs, func(a, b xerr.ImportUnknownFactError) int {
		return cmp.Or(
			cmp.Compare(a.At.File, b.At.File),
			cmp.Compare(a.At.From.Line, b.At.From.Line),
			cmp.Compare(a.At.From.Column, b.At.From.Column),
			cmp.Compare(a.Message, b.Message),
		)
	})
	joined := make([]error, 0, len(errs))
	for _, e := range errs {
		joined = append(joined, e)
	}
	return errors.Join(joined...)
}
