// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2026 Binaek Sarkar
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

package box

import "github.com/sentrie-sh/sentrie/trinary"

func (s *BoxTestSuite) TestTrinaryFrom_matchesFromAny() {
	cases := []Value{
		Undefined(),
		Null(),
		Bool(true),
		Bool(false),
		Number(0),
		Number(3.14),
		String(""),
		String("hello"),
		String("true"),
		Trinary(trinary.True),
		Trinary(trinary.False),
		Trinary(trinary.Unknown),
		List(nil),
		List([]Value{}),
		List([]Value{Number(0)}),
		Map(nil),
		Map(map[string]Value{}),
		Map(map[string]Value{"a": Number(1)}),
		Object(struct{ X int }{1}),
		Object(map[string]any{"k": 1}),
	}
	for _, b := range cases {
		s.Run(b.String(), func() {
			s.Equal(trinary.From(b.Any()), TrinaryFrom(b))
		})
	}
}
