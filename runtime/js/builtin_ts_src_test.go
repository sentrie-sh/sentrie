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

package js

func (s *JSTestSuite) TestBuiltinJSTS_Embedded() {
	s.NotEmpty(BuiltinJSTS, "BuiltinJSTS should not be empty")
	s.Greater(len(BuiltinJSTS), 100, "BuiltinJSTS should contain substantial content")
	source := string(BuiltinJSTS)
	s.Contains(source, "export const round", "Should export Math.round")
	s.Contains(source, "export const length", "Should export String.length")
	s.Contains(source, "export const parse", "Should export JSON.parse")
	s.Contains(source, "export const stringify", "Should export JSON.stringify")
}
