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

package js

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuiltinJSTS_Embedded(t *testing.T) {
	// Verify that the embedded TypeScript source is not empty
	assert.NotEmpty(t, BuiltinJSTS, "BuiltinJSTS should not be empty")
	assert.Greater(t, len(BuiltinJSTS), 100, "BuiltinJSTS should contain substantial content")
	
	// Verify it contains expected exports
	source := string(BuiltinJSTS)
	assert.Contains(t, source, "export const round", "Should export Math.round")
	assert.Contains(t, source, "export const length", "Should export String.length")
	assert.Contains(t, source, "export const parse", "Should export JSON.parse")
	assert.Contains(t, source, "export const stringify", "Should export JSON.stringify")
}

