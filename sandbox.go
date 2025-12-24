//go:build sandbox
// +build sandbox

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
Run with: go run -tags sandbox sandbox.go

IMPORTANT: After testing is complete, RESET this file to its template state:
- Remove all test code from the main() function
- Remove any imports that were added for testing
- Restore the main() function to only contain:
    ctx := context.Background()
    // Your test code here
    fmt.Println("Sandbox test")
    os.Exit(0)
- Keep only the standard imports (context, fmt, os) unless they're needed for the template
*/

func main() {
	ctx := context.Background()

	// Your test code here
	fmt.Println("Sandbox test")

	os.Exit(0)
}
