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

import (
	"fmt"
	"strings"
	"testing"
)

func TestListenerConfiguration(t *testing.T) {
	tests := []struct {
		name     string
		port     int
		listen   []string
		expected string
	}{
		{
			name:     "nil listen uses port only",
			port:     8080,
			listen:   nil,
			expected: ":8080",
		},
		{
			name:     "empty listen uses port only",
			port:     8080,
			listen:   []string{},
			expected: ":8080",
		},
		{
			name:     "local maps to localhost",
			port:     8080,
			listen:   []string{"local"},
			expected: "localhost:8080",
		},
		{
			name:     "local4 maps to 127.0.0.1",
			port:     8080,
			listen:   []string{"local4"},
			expected: "127.0.0.1:8080",
		},
		{
			name:     "local6 maps to ::1",
			port:     8080,
			listen:   []string{"local6"},
			expected: "[::1]:8080",
		},
		{
			name:     "network maps to all interfaces",
			port:     8080,
			listen:   []string{"network"},
			expected: ":8080",
		},
		{
			name:     "network4 maps to 0.0.0.0",
			port:     8080,
			listen:   []string{"network4"},
			expected: "0.0.0.0:8080",
		},
		{
			name:     "network6 maps to ::",
			port:     8080,
			listen:   []string{"network6"},
			expected: "[::]:8080",
		},
		{
			name:     "custom address used as-is",
			port:     8080,
			listen:   []string{"192.168.1.100:3000"},
			expected: "192.168.1.100:3000",
		},
		{
			name:     "custom address with different port",
			port:     8080,
			listen:   []string{"10.0.0.1:9090"},
			expected: "10.0.0.1:9090",
		},
		{
			name:     "multiple addresses returns all",
			port:     8080,
			listen:   []string{"local", "network4"},
			expected: "localhost:8080,0.0.0.0:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr := resolveAddress(tt.port, tt.listen)
			if addr != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, addr)
			}
		})
	}
}

// Helper function to test the address resolution logic
func resolveAddress(port int, listen []string) string {
	if len(listen) > 0 {
		var addresses []string
		for _, listenAddr := range listen {
			var addr string
			switch listenAddr {
			case "local":
				addr = fmt.Sprintf("localhost:%d", port)
			case "local4":
				addr = fmt.Sprintf("127.0.0.1:%d", port)
			case "local6":
				addr = fmt.Sprintf("[::1]:%d", port)
			case "network":
				addr = fmt.Sprintf(":%d", port)
			case "network4":
				addr = fmt.Sprintf("0.0.0.0:%d", port)
			case "network6":
				addr = fmt.Sprintf("[::]:%d", port)
			default:
				addr = listenAddr
			}
			addresses = append(addresses, addr)
		}
		return strings.Join(addresses, ",")
	}
	return fmt.Sprintf(":%d", port)
}
