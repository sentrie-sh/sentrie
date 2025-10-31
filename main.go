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
	"os/signal"

	"github.com/sentrie-sh/sentrie/cmd"
	"github.com/sentrie-sh/sentrie/constants"
)

var version = constants.APPVERSION

func main() {
	ctx := context.Background()

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, os.Kill)
	defer stop()

	// set an exit code
	exitCode := 0

	cli := cmd.Setup(ctx, version)
	if err := cmd.Execute(ctx, cli, os.Args); err != nil {
		// pretty print the error in the forn <red>Error</red>: <error>
		fmt.Printf("Error: %s\n", err)
		exitCode = 1
	}
	os.Exit(exitCode)
}
