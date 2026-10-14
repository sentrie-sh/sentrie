//go:build sandbox
// +build sandbox

// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"os"
	// Add imports as needed for testing
	// Example:
	// "github.com/sentrie-sh/sentrie/index"
	// "github.com/sentrie-sh/sentrie/loader"
	// "github.com/sentrie-sh/sentrie/runtime"
)

/*
This file is used to test the sandboxed programs.
Run with: GOWORK=off go run -tags sandbox sandbox.go

IMPORTANT: After testing is complete, RESET this file to its template state:
- Remove all test code from the main() function
- Remove any imports that were added for testing
- Restore the main() function to only contain:
    ctx := context.Background()
    _ = ctx // remove once your test code uses ctx
    // Your test code here
    fmt.Println("Sandbox test")
    os.Exit(0)
- Keep only the standard imports (context, fmt, os) unless they're needed for the template
*/

func main() {
	ctx := context.Background()
	_ = ctx // remove once your test code uses ctx

	// Your test code here
	fmt.Println("Sandbox test")

	os.Exit(0)
}
