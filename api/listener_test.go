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

package api

func (s *APITestSuite) TestListenerConfiguration() {
	s.Run("nil listen uses port only", func() {
		s.Equal(":8080", resolveAddress(8080, nil))
	})
	s.Run("empty listen uses port only", func() {
		s.Equal(":8080", resolveAddress(8080, []string{}))
	})
	s.Run("local maps to localhost", func() {
		s.Equal("localhost:8080", resolveAddress(8080, []string{"local"}))
	})
	s.Run("local4 maps to 127.0.0.1", func() {
		s.Equal("127.0.0.1:8080", resolveAddress(8080, []string{"local4"}))
	})
	s.Run("local6 maps to ::1", func() {
		s.Equal("[::1]:8080", resolveAddress(8080, []string{"local6"}))
	})
	s.Run("network maps to all interfaces", func() {
		s.Equal(":8080", resolveAddress(8080, []string{"network"}))
	})
	s.Run("network4 maps to 0.0.0.0", func() {
		s.Equal("0.0.0.0:8080", resolveAddress(8080, []string{"network4"}))
	})
	s.Run("network6 maps to ::", func() {
		s.Equal("[::]:8080", resolveAddress(8080, []string{"network6"}))
	})
	s.Run("custom address used as-is", func() {
		s.Equal("192.168.1.100:3000", resolveAddress(8080, []string{"192.168.1.100:3000"}))
	})
	s.Run("custom address with different port", func() {
		s.Equal("10.0.0.1:9090", resolveAddress(8080, []string{"10.0.0.1:9090"}))
	})
	s.Run("multiple addresses returns all", func() {
		s.Equal("localhost:8080,0.0.0.0:8080", resolveAddress(8080, []string{"local", "network4"}))
	})
}
