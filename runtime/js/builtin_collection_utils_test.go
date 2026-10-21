// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package js

// TestCollectionListUnique pins #212: list_unique deduplicates scalars,
// arrays and objects by value, keeps the first occurrence in order, and no
// longer panics on unhashable elements.
func (s *JSTestSuite) TestCollectionListUnique() {
	cases := []struct {
		name string
		expr string
		want string
	}{
		{"scalars", `m.list_unique([3, 1, 3, 2, 1])`, `[3,1,2]`},
		{"strings", `m.list_unique(["b", "a", "b"])`, `["b","a"]`},
		{"number vs string", `m.list_unique([1, "1", 1, "1"])`, `[1,"1"]`},
		{"int and float of one value", `m.list_unique([1, 1.0, 1.5, 1.5])`, `[1,1.5]`},
		{"booleans and null", `m.list_unique([true, null, false, true, null])`, `[true,null,false]`},
		{"arrays", `m.list_unique([[1, 2], [2, 1], [1, 2], []])`, `[[1,2],[2,1],[]]`},
		{"objects", `m.list_unique([{a: 1, b: 2}, {b: 2, a: 1}, {a: 1}, {a: "1"}])`, `[{"a":1,"b":2},{"a":1},{"a":"1"}]`},
		{"nested", `m.list_unique([{a: [1, {b: 2}]}, {a: [1, {b: 2}]}, {a: [1, {b: 3}]}])`, `[{"a":[1,{"b":2}]},{"a":[1,{"b":3}]}]`},
		{"mixed", `m.list_unique([1, [1], {x: 1}, "1", [1], {x: 1}, 1])`, `[1,[1],{"x":1},"1"]`},
		{"empty", `m.list_unique([])`, `[]`},
	}
	for _, tc := range cases {
		s.Run(tc.name, func() {
			got, err := s.runBuiltinScript(BuiltinCollectionGo, `JSON.stringify(`+tc.expr+`)`)
			s.Require().NoError(err)
			s.Equal(tc.want, got.String())
		})
	}
}

// TestCollectionListUniqueNonFinite checks NaN and the infinities, which JSON
// cannot encode, deduplicate like a JavaScript Set instead of throwing.
func (s *JSTestSuite) TestCollectionListUniqueNonFinite() {
	got, err := s.runBuiltinScript(BuiltinCollectionGo, `(function () {
		const r = m.list_unique([NaN, Infinity, NaN, -Infinity, Infinity, 0, -0]);
		return r.length + ":" + Number.isNaN(r[0]) + ":" + r.slice(1).join(",");
	})()`)
	s.Require().NoError(err)
	s.Equal("4:true:Infinity,-Infinity,0", got.String())
}

// TestCollectionListUniqueThrowsOnUncomparable keeps the #111 convention:
// an element with no canonical form throws a catchable error.
func (s *JSTestSuite) TestCollectionListUniqueThrowsOnUncomparable() {
	_, err := s.runBuiltinScript(BuiltinCollectionGo, `m.list_unique([1, function () {}])`)
	s.Require().Error(err)
	s.Contains(err.Error(), "list_unique: element 1 cannot be compared")
}

// TestCollectionListUniqueKeepsOriginals checks the result is a real array
// holding the caller's own first-occurrence objects.
func (s *JSTestSuite) TestCollectionListUniqueKeepsOriginals() {
	got, err := s.runBuiltinScript(BuiltinCollectionGo, `(function () {
		const a = {id: 1}, b = {id: 1}, c = {id: 2};
		const r = m.list_unique([a, b, c]);
		return Array.isArray(r) + ":" + r.length + ":" + (r[0] === a) + ":" + (r[1] === c);
	})()`)
	s.Require().NoError(err)
	s.Equal("true:2:true:true", got.String())
}
