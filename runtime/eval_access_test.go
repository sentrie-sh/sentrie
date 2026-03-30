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
	"context"
	"testing"

	"github.com/sentrie-sh/sentrie/box"
	"github.com/stretchr/testify/require"
)

func TestAccessFieldPreservesBoxedUndefined(t *testing.T) {
	obj := box.Map(map[string]box.Value{
		"nested": box.Undefined(),
	})
	out, err := accessField(context.Background(), obj, "nested")
	require.NoError(t, err)
	require.True(t, out.IsUndefined())
}

func TestAccessIndexPreservesBoxedUndefined(t *testing.T) {
	col := box.List([]box.Value{box.Undefined()})
	out, err := accessIndex(context.Background(), col, box.Number(0))
	require.NoError(t, err)
	require.True(t, out.IsUndefined())
}

func TestAccessIndexMapAnyMissingKeyReturnsUndefined(t *testing.T) {
	col := box.Object(map[string]any{
		"present": 1,
	})
	out, err := accessIndex(context.Background(), col, box.String("missing"))
	require.NoError(t, err)
	require.True(t, out.IsUndefined())
}
