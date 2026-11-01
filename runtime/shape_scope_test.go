// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"github.com/sentrie-sh/sentrie/box"
)

// shapeScopeCatalog exports, from com/example/catalog, a shape whose field
// names a sibling shape (Item), an alias of a sibling shape (Priced), and a
// shape that takes a field naming a sibling shape in through `with` (Base).
const shapeScopeCatalog = `namespace com/example/catalog
shape Price {
  amount: number
}
shape Item {
  sku: string
  price: Price
}
shape Priced Price
shape Base {
  price: Price
}
export shape Item
export shape Priced
export shape Base
`

// shapeScopeShop builds com/example/shop's source: its own shapes, then
// policy cart with fact line and one exported rule per "name = expr" pair.
func shapeScopeShop(shopShapes string, rules ...[2]string) string {
	src := "namespace com/example/shop\n" + shopShapes + "policy cart {\n  fact line: document as line\n"
	for _, r := range rules {
		src += "  rule " + r[0] + " = { yield " + r[1] + " }\n  export decision of " + r[0] + "\n"
	}
	return src + "}\n"
}

// runShapeScopeRule builds the programs and runs one rule of
// com/example/shop's policy cart with line set to the given fact.
func (s *RuntimeTestSuite) runShapeScopeRule(programs []deriveTestProgram, rule string, line any) (box.Value, error) {
	ctx := s.T().Context()
	_, exec := s.mustBuildDeriveExecutor(ctx, programs...)
	res, err := exec.ExecRule(ctx, "com/example/shop", "cart", rule, map[string]any{"line": line})
	if err != nil {
		return box.Undefined(), err
	}
	return res.Decision.Value, nil
}

// TestShapeFieldAndAliasTypesUseShapeScope checks that an exported shape's
// field type, alias and `with` field, each naming a sibling shape, resolve in
// the shape's namespace when the shape is used from another namespace: one
// without the sibling, and one that declares a different Price (#281).
func (s *RuntimeTestSuite) TestShapeFieldAndAliasTypesUseShapeScope() {
	for name, shopShapes := range map[string]string{
		"caller without Price":      "",
		"caller with its own Price": "shape Price {\n  cents: number\n}\n",
	} {
		s.Run(name, func() {
			shop := shapeScopeShop(shopShapes+"shape Line with com/example/catalog/Base {\n  sku: string\n}\n",
				[2]string{"item", "line is com/example/catalog/Item"},
				[2]string{"priced", "line.price is com/example/catalog/Priced"},
				[2]string{"composed", "line is Line"},
			)
			programs := []deriveTestProgram{
				{name: "catalog.sentrie", src: shapeScopeCatalog},
				{name: "shop.sentrie", src: shop},
			}
			good := map[string]any{"sku": "a", "price": map[string]any{"amount": 1}}
			// Only the shop's Price: the catalog's Price wants amount.
			callers := map[string]any{"sku": "a", "price": map[string]any{"cents": 1}}

			for _, rule := range []string{"item", "priced", "composed"} {
				got, err := s.runShapeScopeRule(programs, rule, good)
				s.Require().NoError(err, rule)
				s.Equal(box.Bool(true), got, rule)

				got, err = s.runShapeScopeRule(programs, rule, callers)
				s.Require().NoError(err, rule)
				s.Equal(box.Bool(false), got, rule)
			}
		})
	}
}

// TestShapeFieldTypeErrorNamesShapeScope checks that a shape's field type
// that names a sibling shape is checked, not reported as missing, when the
// value is cast from another namespace: the mismatch is the sibling's field.
func (s *RuntimeTestSuite) TestShapeFieldTypeErrorNamesShapeScope() {
	shop := shapeScopeShop("",
		[2]string{"cast", `(line as com/example/catalog/Item).sku`},
	)
	programs := []deriveTestProgram{
		{name: "catalog.sentrie", src: shapeScopeCatalog},
		{name: "shop.sentrie", src: shop},
	}
	got, err := s.runShapeScopeRule(programs, "cast", map[string]any{"sku": "a", "price": map[string]any{"amount": 1}})
	s.Require().NoError(err)
	s.Equal(box.String("a"), got)

	_, err = s.runShapeScopeRule(programs, "cast", map[string]any{"sku": "a", "price": map[string]any{"cents": 1}})
	s.Require().Error(err)
	s.Contains(err.Error(), "field amount is required")
	s.NotContains(err.Error(), "not found")
}

// TestPolicyShapeFieldTypesUsePolicyScope checks that a policy-scoped shape's
// field and alias types still resolve in its policy first: the policy's Inner,
// not the namespace's Inner of the same name.
func (s *RuntimeTestSuite) TestPolicyShapeFieldTypesUsePolicyScope() {
	src := `namespace com/example/shop
shape Inner {
  b: string
}
policy cart {
  fact line: document as line
  shape Inner {
    a: number
  }
  shape Outer {
    i: Inner
  }
  shape InnerAlias Inner
  rule outer = { yield line is Outer }
  rule alias = { yield line.i is InnerAlias }
  export decision of outer
  export decision of alias
}
`
	programs := []deriveTestProgram{{name: "shop.sentrie", src: src}}
	for _, rule := range []string{"outer", "alias"} {
		got, err := s.runShapeScopeRule(programs, rule, map[string]any{"i": map[string]any{"a": 1}})
		s.Require().NoError(err, rule)
		s.Equal(box.Bool(true), got, rule)

		got, err = s.runShapeScopeRule(programs, rule, map[string]any{"i": map[string]any{"b": "x"}})
		s.Require().NoError(err, rule)
		s.Equal(box.Bool(false), got, rule)
	}
}
