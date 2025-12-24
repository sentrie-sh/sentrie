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

import (
	"fmt"
	"strings"
	"testing"
)

func TestListenerConfiguration(t *testing.T) {
	// nil listen uses port only
	t.Run("nil listen uses port only", func(t *testing.T) {
		addr := resolveAddress(8080, nil)
		if addr != ":8080" {
			t.Errorf("Expected %s, got %s", ":8080", addr)
		}
	})

	// empty listen uses port only
	t.Run("empty listen uses port only", func(t *testing.T) {
		addr := resolveAddress(8080, []string{})
		if addr != ":8080" {
			t.Errorf("Expected %s, got %s", ":8080", addr)
		}
	})

	// local maps to localhost
	t.Run("local maps to localhost", func(t *testing.T) {
		addr := resolveAddress(8080, []string{"local"})
		if addr != "localhost:8080" {
			t.Errorf("Expected %s, got %s", "localhost:8080", addr)
		}
	})

	// local4 maps to 127.0.0.1
	t.Run("local4 maps to 127.0.0.1", func(t *testing.T) {
		addr := resolveAddress(8080, []string{"local4"})
		if addr != "127.0.0.1:8080" {
			t.Errorf("Expected %s, got %s", "127.0.0.1:8080", addr)
		}
	})

	// local6 maps to ::1
	t.Run("local6 maps to ::1", func(t *testing.T) {
		addr := resolveAddress(8080, []string{"local6"})
		if addr != "[::1]:8080" {
			t.Errorf("Expected %s, got %s", "[::1]:8080", addr)
		}
	})

	// network maps to all interfaces
	t.Run("network maps to all interfaces", func(t *testing.T) {
		addr := resolveAddress(8080, []string{"network"})
		if addr != ":8080" {
			t.Errorf("Expected %s, got %s", ":8080", addr)
		}
	})

	// network4 maps to 0.0.0.0
	t.Run("network4 maps to 0.0.0.0", func(t *testing.T) {
		addr := resolveAddress(8080, []string{"network4"})
		if addr != "0.0.0.0:8080" {
			t.Errorf("Expected %s, got %s", "0.0.0.0:8080", addr)
		}
	})

	// network6 maps to ::
	t.Run("network6 maps to ::", func(t *testing.T) {
		addr := resolveAddress(8080, []string{"network6"})
		if addr != "[::]:8080" {
			t.Errorf("Expected %s, got %s", "[::]:8080", addr)
		}
	})

	// custom address used as-is
	t.Run("custom address used as-is", func(t *testing.T) {
		addr := resolveAddress(8080, []string{"192.168.1.100:3000"})
		if addr != "192.168.1.100:3000" {
			t.Errorf("Expected %s, got %s", "192.168.1.100:3000", addr)
		}
	})

	// custom address with different port
	t.Run("custom address with different port", func(t *testing.T) {
		addr := resolveAddress(8080, []string{"10.0.0.1:9090"})
		if addr != "10.0.0.1:9090" {
			t.Errorf("Expected %s, got %s", "10.0.0.1:9090", addr)
		}
	})

	// multiple addresses returns all
	t.Run("multiple addresses returns all", func(t *testing.T) {
		addr := resolveAddress(8080, []string{"local", "network4"})
		if addr != "localhost:8080,0.0.0.0:8080" {
			t.Errorf("Expected %s, got %s", "localhost:8080,0.0.0.0:8080", addr)
		}
	})
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
