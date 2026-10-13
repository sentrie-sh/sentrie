// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"fmt"
	"net"
	"slices"

	"github.com/binaek/gocoll/collection"
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

	var addresses = make([]string, 0, len(listen))
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
