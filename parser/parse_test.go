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

package parser

import (
	"testing"
)

func TestParser_Programs(t *testing.T) {
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := t.Context()
			p := NewParserFromString(tc.src, tc.name)
			prg, err := p.ParseProgram(ctx)
			if err != nil {
				t.Errorf("parse failed: %v", err)
			}
			_ = prg
		})
	}
}

var tests = []struct {
	name string
	src  string
}{
	{
		name: "01-minimal-namespace",
		src: `namespace minimal
`,
	},
	{
		name: "02-empty-policy-block",
		src: `namespace empty_policy
policy nothing {}
`,
	},
	{
		name: "03-comments-only",
		src: `namespace comments
-- top-level note
policy c1 {
  -- inside policy
}
`,
	},
	{
		name: "04-scalar-lets",
		src: `namespace scalars
policy scalar_vals {
  let s = "hello"
  let i = 42
  let f = 3.1415
  let b = true
  let n = null
}
`,
	},
	{
		name: "05-collection-literals",
		src: `namespace collections
policy lists_and_maps {
  let fruits = ["apple", "banana", "cherry"]
  let ages   = {"alice": 30, "bob": 25}
}
`,
	},
	{
		name: "06-arithmetic-precedence",
		src: `namespace arithmetic
policy arith {
  let result = (2 + 3) * 4 % 5 - 6 / 2
}
`,
	},
	{
		name: "07-logical-operators",
		src: `namespace logical
policy logic_ops {
  let ok = true and false or true xor false
}
`,
	},
	{
		name: "08-unary-not-bang",
		src: `namespace unary
policy neg {
  let flag = not false
  let bang = !true
}
`,
	},
	{
		name: "09-ternary-expression",
		src: `namespace ternary
policy choose {
  let x   = 10
  let y   = 20
  let max = x > y ? x : y
}
`,
	},
	{
		name: "10-equality-and-relational",
		src: `namespace compare
policy cmp {
  let eq  = 5 == 5
  let ne  = 5 != 6
  let lt  = 4 < 5
  let lte = 5 <= 5
  let gt  = 6 > 5
  let gte = 6 >= 6
}
`,
	},
	{
		name: "11-is-defined-empty",
		src: `namespace isops
policy defined_empty {
  let lst = []
  let m   = {"k": "v"}
  let a1  = lst is empty
  let a2  = lst is not empty
  let a3  = m   is defined
  let a4  = m   is not defined
}
`,
	},
	{
		name: "12-matches-contains-in",
		src: `namespace memops
policy match_contain_in {
  let regex    = "abc123" matches "[a-z]+[0-9]+"
  let nomatch  = "foo" not matches "[0-9]+"
  let has      = ["x", "y", "z"] contains "y"
  let nohas    = ["x", "y"] not contains "z"
  let inside   = 3 in [1, 2, 3, 4]
  let outside  = 9 not in [1, 2, 3]
}
`,
	},
	{
		name: "13-list-map-comparison-kv-in",
		src: `namespace listmapcmp
policy cmp2 {
  let ls   = ["a", "b"]
  let t1   = ls in  [ ["a", "b"], ["c"] ]
  let mp   = { "k": "v" }
  let t2   = mp not in [ { "x": 1 }, { "k": "v" } ]
  let kvok = { "x": 1 } in [ { "x": 1, "y": 2 }, { "a": 3 } ]
}
`,
	},
	{
		name: "14-comprehensions-aggregations",
		src: `namespace comprehensions
policy comp {
  let nums        = [1, 2, 3, 4, 5]
  let existsEven  = any nums as n { n % 2 == 0 }
  let allPos      = all nums as n { n > 0 }
  let evens       = filter nums as n { n % 2 == 0 }
  let doubled     = map nums as n { n * 2 }
  let uniq        = distinct nums as n { n }
  let sum         = reduce 0 from nums as acc, n { acc + n }
  let total       = count nums
}
`,
	},
	{
		name: "15-function-call-index-access",
		src: `namespace callindex
policy funcs {
  let upper = toUpper("hello")
  let first = ["a", "b", "c"][0]
}
`,
	},
	{
		name: "16-rule-inline-block",
		src: `namespace rules1
policy simple_rule {
  rule isAdult = { age >= 18 }
}
`,
	},
	{
		name: "17-rule-import-with-with",
		src: `namespace rules2
policy consumer {
  rule checkAccess = import decision isAllowed from auth
                     with user     as subject
                     with resource as obj
}
`,
	},
	{
		name: "18-fact-declaration",
		src: `namespace facts
policy employment {
  fact "employee_type" as role
}
`,
	},
	{
		name: "19-export-with-attach",
		src: `namespace exports
policy prod {
  rule passed = { score >= 60 }
  export decision of passed attach severity as "low"
}
`,
	},
	{
		name: "20-use-string-source",
		src: `namespace usestr
policy util {
  use jsonParse from "file://json.wasm" as json
}
`,
	},
	{
		name: "21-use-at-source",
		src: `namespace usemod
policy util2 {
  use add, sub from @math/core as m
}
`,
	},
	{
		name: "22-multi-policy-cross-import",
		src: `namespace multi

policy base {
  rule allowAll = { true }
}

policy consumer {
  rule allow = import decision allowAll from base
}
`,
	},
	{
		name: "23-nested-mixed-precedence",
		src: `namespace complex
policy big {
  let data   = { "ids": [1, 2, 3, 4], "active": true }
  let result =
    any data["ids"] as id {
      (id % 2 == 0 and not (id in [4, 6])) or id == 3
    } ? "ok" : "fail"
}
`,
	},
	{
		name: "24-ternary-top-level",
		src: `namespace ternary2
policy cond {
  let a = true ? 1 : 0
}
`,
	},
	{
		name: "25-comments-with-operators",
		src: `namespace commentmix
-- file header
policy mix {
  -- inner note
  let r = (5 + 3 * 2) xor (not false and true) ? -- trailing comment
          "yes" : "no"
}
`,
	},
}
