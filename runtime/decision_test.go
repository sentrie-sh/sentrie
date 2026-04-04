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
	"encoding/json"

	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/trinary"
)

func (s *RuntimeTestSuite) TestDecisionOfUnknownInputs() {
	for _, v := range []box.Value{box.Undefined(), box.Null()} {
		d := DecisionOf(v)
		s.Require().Equal(trinary.Unknown, d.State)
		s.Require().Equal(v, d.Value)
	}
}

func (s *RuntimeTestSuite) TestDecisionOfUsesTrinaryAndFallbackConversion() {
	td := DecisionOf(box.Trinary(trinary.True))
	s.Require().Equal(trinary.True, td.State)

	fd := DecisionOf(box.Bool(false))
	s.Require().Equal(trinary.False, fd.State)
}

func (s *RuntimeTestSuite) TestDecisionMarshalJSONIncludesStateAndValue() {
	raw, err := json.Marshal(Decision{
		State: trinary.True,
		Value: box.String("ok"),
	})
	s.Require().NoError(err)
	s.Require().JSONEq(`{"state":"true","value":"ok"}`, string(raw))
}
