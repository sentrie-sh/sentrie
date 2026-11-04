// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"github.com/sentrie-sh/sentrie/trinary"
)

// TestKeywordSegmentNamespaceResolvesAcrossNamespaces checks a namespace whose
// path has keyword segments (#297) end to end: it validates, its shape is a
// base and a fact type in another namespace, its decision is imported there,
// and its own rules run under the keyword path.
func (s *RuntimeTestSuite) TestKeywordSegmentNamespaceResolvesAcrossNamespaces() {
	ctx := s.T().Context()
	lib := deriveTestProgram{name: "lib.sentrie", src: `namespace text/string

shape Item {
  name: string
}
export shape Item

policy p {
  fact item: Item
  rule ok = { yield item.name == "a" }
  export decision of ok
}
`}
	app := deriveTestProgram{name: "app.sentrie", src: `namespace app

shape Tagged with text/string/Item {
  label: string
}

policy q {
  fact t: Tagged
  fact i: text/string/Item
  rule imported = import decision ok from text/string/p with item as i
  rule both = { yield imported and t.label == "x" }
  export decision of both
}
`}
	_, exec := s.mustBuildDeriveExecutor(ctx, lib, app)

	out, err := exec.ExecRule(ctx, "text/string", "p", "ok", map[string]any{"item": map[string]any{"name": "a"}})
	s.Require().NoError(err)
	s.Equal(trinary.True, out.Decision.State)

	facts := func(name string) map[string]any {
		return map[string]any{
			"t": map[string]any{"name": "a", "label": "x"},
			"i": map[string]any{"name": name},
		}
	}
	out, err = exec.ExecRule(ctx, "app", "q", "both", facts("a"))
	s.Require().NoError(err)
	s.Equal(trinary.True, out.Decision.State)

	out, err = exec.ExecRule(ctx, "app", "q", "both", facts("b"))
	s.Require().NoError(err)
	s.Equal(trinary.False, out.Decision.State)

	// The base shape from the keyword namespace is enforced on the fact.
	_, err = exec.ExecRule(ctx, "app", "q", "both", map[string]any{
		"t": map[string]any{"label": "x"},
		"i": map[string]any{"name": "a"},
	})
	s.Require().Error(err)
	s.Contains(err.Error(), "name")
}

// TestKeywordSegmentNamespaceInExpressions checks the expression positions end
// to end: a derive exported from `lib/in` is called as `lib/in/double(2)`, and
// `x is lib/in/Item` checks against its exported shape (#297).
func (s *RuntimeTestSuite) TestKeywordSegmentNamespaceInExpressions() {
	ctx := s.T().Context()
	lib := deriveTestProgram{name: "lib.sentrie", src: `namespace lib/in

shape Item {
  name: string
}
export shape Item

derive double = (n: number): number => { yield n * 2 }
export derive double
`}
	app := deriveTestProgram{name: "app.sentrie", src: `namespace app

policy q {
  fact v: document
  rule doubled = { yield lib/in/double(2) == 4 }
  rule isItem = { yield v is lib/in/Item }
  export decision of doubled
  export decision of isItem
}
`}
	_, exec := s.mustBuildDeriveExecutor(ctx, lib, app)

	out, err := exec.ExecRule(ctx, "app", "q", "doubled", map[string]any{"v": map[string]any{}})
	s.Require().NoError(err)
	s.Equal(trinary.True, out.Decision.State)

	out, err = exec.ExecRule(ctx, "app", "q", "isItem", map[string]any{"v": map[string]any{"name": "a"}})
	s.Require().NoError(err)
	s.Equal(trinary.True, out.Decision.State)

	out, err = exec.ExecRule(ctx, "app", "q", "isItem", map[string]any{"v": map[string]any{"name": 1}})
	s.Require().NoError(err)
	s.Equal(trinary.False, out.Decision.State)
}
