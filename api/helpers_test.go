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
)

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
