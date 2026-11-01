// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"sync"

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

// shapeScopeDeepCatalog exports, from com/example/catalog, shapes that reach
// a sibling Price through a second shape, a list, a dict and a nullable type.
const shapeScopeDeepCatalog = `namespace com/example/catalog
shape Price {
  amount: number
}
shape Item {
  price: Price
}
shape Order {
  item: Item
}
shape Bag {
  prices: list[Price]
  byName: dict[Price]
  maybe: Price?
}
shape Base {
  price: Price
}
export shape Order
export shape Bag
export shape Base
`

// shapeScopeDeepPricing is a third namespace: an alias of the catalog's
// exported Bag, and a shape that takes the catalog's Base in through `with`
// and adds a field naming pricing's own Money.
const shapeScopeDeepPricing = `namespace com/example/pricing
shape Money {
  value: number
}
shape Wrapped com/example/catalog/Bag
shape Mid with com/example/catalog/Base {
  money: Money
}
export shape Wrapped
export shape Mid
`

// TestShapeScopeNestedAndContainerFieldTypes checks that a Price reached
// through a field of a field, a list element, a dict element and a nullable
// field resolves in the catalog, not in the shop, with and without a shop
// Price of the same name.
func (s *RuntimeTestSuite) TestShapeScopeNestedAndContainerFieldTypes() {
	good := func(p any) map[string]any {
		return map[string]any{
			"prices": []any{p, p},
			"byName": map[string]any{"a": p},
			"maybe":  p,
		}
	}
	catalogPrice := map[string]any{"amount": 1}
	shopPrice := map[string]any{"cents": 1}

	for name, shopShapes := range map[string]string{
		"caller without Price":      "",
		"caller with its own Price": "shape Price {\n  cents: number\n}\n",
	} {
		s.Run(name, func() {
			programs := []deriveTestProgram{
				{name: "catalog.sentrie", src: shapeScopeDeepCatalog},
				{name: "shop.sentrie", src: shapeScopeShop(shopShapes,
					[2]string{"order", "line is com/example/catalog/Order"},
					[2]string{"bag", "line is com/example/catalog/Bag"},
				)},
			}
			order := func(p any) map[string]any {
				return map[string]any{"item": map[string]any{"price": p}}
			}
			bad := func(label string, rule string, line any) {
				got, err := s.runShapeScopeRule(programs, rule, line)
				s.Require().NoError(err, label)
				s.Equal(box.Bool(false), got, label)
			}

			got, err := s.runShapeScopeRule(programs, "order", order(catalogPrice))
			s.Require().NoError(err)
			s.Equal(box.Bool(true), got, "two-level field")
			bad("two-level field, shop-shaped Price", "order", order(shopPrice))

			got, err = s.runShapeScopeRule(programs, "bag", good(catalogPrice))
			s.Require().NoError(err)
			s.Equal(box.Bool(true), got, "list, dict and nullable fields")

			nilMaybe := good(catalogPrice)
			nilMaybe["maybe"] = nil
			got, err = s.runShapeScopeRule(programs, "bag", nilMaybe)
			s.Require().NoError(err)
			s.Equal(box.Bool(true), got, "null nullable field")

			// Each checked container position fails on its own when only its
			// Price is wrong. A dict's element type is not enforced at runtime,
			// so byName is only checked for the dict itself.
			for field, wrong := range map[string]any{
				"prices": []any{catalogPrice, shopPrice},
				"maybe":  shopPrice,
			} {
				line := good(catalogPrice)
				line[field] = wrong
				bad("wrong "+field, "bag", line)
			}
		})
	}
}

// TestShapeScopeAliasFromThirdNamespaceAndComposedFields checks an alias of
// an exported shape from a third namespace, and shapes that take fields in
// through `with` from another namespace: each field's type resolves where it
// was written (catalog's Price, pricing's Money), never in the shop.
func (s *RuntimeTestSuite) TestShapeScopeAliasFromThirdNamespaceAndComposedFields() {
	bag := func(p any) map[string]any {
		return map[string]any{"prices": []any{p}, "byName": map[string]any{}, "maybe": nil}
	}
	mid := func(price, money any) map[string]any {
		return map[string]any{"price": price, "money": money}
	}
	catalogPrice := map[string]any{"amount": 1}
	shopPrice := map[string]any{"cents": 1}
	pricingMoney := map[string]any{"value": 1}
	shopMoney := map[string]any{"units": 1}

	for name, shopShapes := range map[string]string{
		"caller without siblings":      "",
		"caller with its own siblings": "shape Price {\n  cents: number\n}\nshape Money {\n  units: number\n}\n",
	} {
		s.Run(name, func() {
			programs := []deriveTestProgram{
				{name: "catalog.sentrie", src: shapeScopeDeepCatalog},
				{name: "pricing.sentrie", src: shapeScopeDeepPricing},
				{name: "shop.sentrie", src: shapeScopeShop(shopShapes,
					[2]string{"wrapped", "line is com/example/pricing/Wrapped"},
					[2]string{"mid", "line is com/example/pricing/Mid"},
				)},
			}
			check := func(label, rule string, v any, want bool) {
				got, err := s.runShapeScopeRule(programs, rule, v)
				s.Require().NoError(err, label)
				s.Equal(box.Bool(want), got, label)
			}

			check("alias, catalog Price", "wrapped", bag(catalogPrice), true)
			check("alias, shop Price", "wrapped", bag(shopPrice), false)

			// Mid declares money itself and takes price in from catalog's Base.
			check("with, both from their own namespace", "mid", mid(catalogPrice, pricingMoney), true)
			check("with, Price from the shop", "mid", mid(shopPrice, pricingMoney), false)
			check("with, Money from the shop", "mid", mid(catalogPrice, shopMoney), false)
		})
	}
}

// TestShapeScopeConcurrentEvaluationFromTwoPolicies evaluates one exported
// shape at the same time from two policies, one of which declares its own
// Price: neither sees the other's, and the shared shape is only read. Run it
// with -race to check the latter.
func (s *RuntimeTestSuite) TestShapeScopeConcurrentEvaluationFromTwoPolicies() {
	shop := `namespace com/example/shop
shape Price {
  cents: number
}
policy plain {
  fact line: document as line
  rule ok = { yield line is com/example/catalog/Order }
  export decision of ok
}
policy other {
  fact line: document as line
  shape Price {
    nope: string
  }
  rule ok = { yield line is com/example/catalog/Order }
  export decision of ok
}
`
	ctx := s.T().Context()
	_, exec := s.mustBuildDeriveExecutor(ctx,
		deriveTestProgram{name: "catalog.sentrie", src: shapeScopeDeepCatalog},
		deriveTestProgram{name: "shop.sentrie", src: shop},
	)
	order := func(p any) map[string]any {
		return map[string]any{"line": map[string]any{"item": map[string]any{"price": p}}}
	}
	const rounds = 16
	var wg sync.WaitGroup
	results := make([]box.Value, 4*rounds)
	errs := make([]error, 4*rounds)
	for i := range rounds {
		for j, c := range []struct {
			policy string
			price  any
		}{
			{"plain", map[string]any{"amount": 1}},
			{"other", map[string]any{"amount": 1}},
			{"plain", map[string]any{"cents": 1}},
			{"other", map[string]any{"nope": "x"}},
		} {
			k := 4*i + j
			wg.Go(func() {
				res, err := exec.ExecRule(ctx, "com/example/shop", c.policy, "ok", order(c.price))
				errs[k] = err
				if err == nil {
					results[k] = res.Decision.Value
				}
			})
		}
	}
	wg.Wait()
	for k := range results {
		s.Require().NoError(errs[k])
		s.Equal(box.Bool(k%4 < 2), results[k], "call %d", k)
	}
}
