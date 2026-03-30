// SPDX-License-Identifier: Apache-2.0
//
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

package box_test

import (
	"testing"

	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/trinary"
	"github.com/stretchr/testify/require"
)

func TestTrinaryFrom_matchesFromAny(t *testing.T) {
	cases := []box.Value{
		box.Undefined(),
		box.Null(),
		box.Bool(true),
		box.Bool(false),
		box.Number(0),
		box.Number(3.14),
		box.String(""),
		box.String("hello"),
		box.String("true"),
		box.Trinary(trinary.True),
		box.Trinary(trinary.False),
		box.Trinary(trinary.Unknown),
		box.List(nil),
		box.List([]box.Value{}),
		box.List([]box.Value{box.Number(0)}),
		box.Map(nil),
		box.Map(map[string]box.Value{}),
		box.Map(map[string]box.Value{"a": box.Number(1)}),
		box.Object(struct{ X int }{1}),
		box.Object(map[string]any{"k": 1}),
	}
	for _, b := range cases {
		t.Run(b.String(), func(t *testing.T) {
			require.Equal(t, trinary.From(b.Any()), box.TrinaryFrom(b))
		})
	}
}
