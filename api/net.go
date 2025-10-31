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
	"net"

	"github.com/binaek/gocoll/collection"
	"golang.org/x/exp/slices"
)

func resolveBindings(port int, listen []string) ([]string, error) {
	predefined := [...]string{"local", "local4", "local6", "network", "network4", "network6"}

	// if any of the listen addresses is in the predefined list - then there MUST be exactly one address
	for _, listenAddr := range listen {
		if slices.Contains(predefined[:], listenAddr) {
			if len(listen) != 1 {
				return nil, fmt.Errorf("when using predefined listen addresses, there must be exactly one address")
			}
		}
	}

	var addresses []string = make([]string, 0, len(listen))
	if slices.Contains(predefined[:], listen[0]) {
		switch listen[0] {
		case "local":
			addresses = []string{net.JoinHostPort("localhost", fmt.Sprintf("%d", port))}
		case "local4":
			addresses = []string{net.JoinHostPort("127.0.0.1", fmt.Sprintf("%d", port))}
		case "local6":
			addresses = []string{net.JoinHostPort("[::1]", fmt.Sprintf("%d", port))}
		case "network":
			addresses = []string{net.JoinHostPort("", fmt.Sprintf("%d", port))}
		case "network4":
			addresses = []string{net.JoinHostPort("0.0.0.0", fmt.Sprintf("%d", port))}
		case "network6":
			addresses = []string{net.JoinHostPort("[::]", fmt.Sprintf("%d", port))}
		}
	} else {
		addresses = collection.Map(
			collection.From(listen...),
			func(listenAddr string) string {
				return net.JoinHostPort(listenAddr, fmt.Sprintf("%d", port))
			},
		).Elements()
	}

	return addresses, nil
}
