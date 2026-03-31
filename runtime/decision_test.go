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

package runtime

import (
	"encoding/json"
	"testing"

	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/trinary"
	"github.com/stretchr/testify/require"
)

func TestDecisionOfUnknownInputs(t *testing.T) {
	for _, v := range []box.Value{box.Undefined(), box.Null()} {
		d := DecisionOf(v)
		require.Equal(t, trinary.Unknown, d.State)
		require.Equal(t, v, d.Value)
	}
}

func TestDecisionOfUsesTrinaryAndFallbackConversion(t *testing.T) {
	td := DecisionOf(box.Trinary(trinary.True))
	require.Equal(t, trinary.True, td.State)

	fd := DecisionOf(box.Bool(false))
	require.Equal(t, trinary.False, fd.State)
}

func TestDecisionMarshalJSONIncludesStateAndValue(t *testing.T) {
	raw, err := json.Marshal(Decision{
		State: trinary.True,
		Value: box.String("ok"),
	})
	require.NoError(t, err)
	require.JSONEq(t, `{"state":"true","value":"ok"}`, string(raw))
}
