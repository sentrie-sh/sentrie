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

package runtime

import (
	"context"

	"github.com/sentrie-sh/sentrie/box"
)

func (s *RuntimeTestSuite) TestAccessFieldPreservesBoxedUndefined() {
	obj := box.Map(map[string]box.Value{
		"nested": box.Undefined(),
	})
	out, err := accessField(context.Background(), obj, "nested")
	s.Require().NoError(err)
	s.Require().True(out.IsUndefined())
}

func (s *RuntimeTestSuite) TestAccessIndexPreservesBoxedUndefined() {
	col := box.List([]box.Value{box.Undefined()})
	out, err := accessIndex(context.Background(), col, box.Number(0))
	s.Require().NoError(err)
	s.Require().True(out.IsUndefined())
}

func (s *RuntimeTestSuite) TestAccessIndexMapAnyMissingKeyReturnsUndefined() {
	col := box.Object(map[string]any{
		"present": 1,
	})
	out, err := accessIndex(context.Background(), col, box.String("missing"))
	s.Require().NoError(err)
	s.Require().True(out.IsUndefined())
}
