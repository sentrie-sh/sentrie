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
	"os/signal"
	"runtime/debug"
	"time"

	"github.com/sentrie-sh/sentrie/cmd"
	"github.com/sentrie-sh/sentrie/constants"
)

// version is overridden at build time using -ldflags
// Example: -ldflags "-X main.version=v1.0.0"
// commit, date, and dirty are extracted from debug.ReadBuildInfo() at runtime
var version = constants.APPVERSION
var builtWithMakefile = "false"

func main() {
	ctx := context.Background()

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, os.Kill)
	defer stop()

	// set an exit code
	exitCode := 0

	cli := cmd.Setup(ctx, getVersionString())
	if err := cmd.Execute(ctx, cli, os.Args); err != nil {
		// pretty print the error in the forn <red>Error</red>: <error>
		fmt.Printf("Error: %s\n", err)
		exitCode = 1
	}
	os.Exit(exitCode)
}

// if dirty, return the version string of the following format:
// <version> (<commit>) <date> <os>/<arch> <go-version>
// else return the version string as is
func getVersionString() string {
	trueVersion := version
	commit, date, dirty := getBuildInfo()
	if dirty || builtWithMakefile == "true" {
		buildNumber := date.Format("20060102150405")
		if len(commit) > 7 {
			commit = commit[:7]
		}
		trueVersion = fmt.Sprintf("%s-dirty.%s+%s", version, buildNumber, commit)
		if builtWithMakefile == "true" {
			trueVersion = fmt.Sprintf("%s(makefile)", trueVersion)
		}
	}
	return trueVersion
}

func getBuildInfo() (commit string, date time.Time, dirty bool) {
	commit = "none"
	date = time.Time{}
	dirty = false

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return commit, date, dirty
	}

	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			commit = setting.Value
		case "vcs.time":
			date, _ = time.Parse(time.RFC3339, setting.Value)
		case "vcs.modified":
			dirty = setting.Value == "true"
		}
	}

	return commit, date, dirty
}
