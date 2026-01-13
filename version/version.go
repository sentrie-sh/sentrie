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

package version

import (
	"fmt"
	"runtime/debug"
	"strings"
	"text/tabwriter"
)

// Info holds version information for the application.
// Fields are exported to match the caarlos0/go-version API.
type Info struct {
	Name         string
	Description  string
	Website      string
	GitVersion   string
	GitCommit    string
	GitTreeState string
	BuildDate    string
	BuiltBy      string
	asciiName    string
}

// Option is a function that configures an Info struct.
type Option func(*Info)

// WithAppDetails sets the application name, description, and website.
func WithAppDetails(name, description, website string) Option {
	return func(i *Info) {
		i.Name = name
		i.Description = description
		i.Website = website
	}
}

// WithASCIIName sets the ASCII art name.
func WithASCIIName(asciiName string) Option {
	return func(i *Info) {
		i.asciiName = asciiName
	}
}

// GetVersionInfo creates a new Info struct pre-filled with build information from debug.BuildInfo,
// then applies the given options.
func GetVersionInfo(opts ...Option) Info {
	info := Info{}

	// Pre-fill from debug.BuildInfo
	bi, _ := debug.ReadBuildInfo()
	if bi != nil {
		// Extract build settings
		for _, setting := range bi.Settings {
			switch setting.Key {
			case "vcs.revision":
				info.GitCommit = setting.Value
			case "vcs.time":
				info.BuildDate = setting.Value
			case "vcs.modified":
				if setting.Value == "true" {
					info.GitTreeState = "dirty"
				} else {
					info.GitTreeState = "clean"
				}
			}
		}

		// Use Main.Version if available (for development builds it might be "(devel)")
		if bi.Main.Version != "" && bi.Main.Version != "(devel)" {
			info.GitVersion = bi.Main.Version
		}
	}

	// Process options (these can override the pre-filled values)
	for _, opt := range opts {
		opt(&info)
	}

	return info
}

// String returns a formatted version string with ASCII art and build information.
func (i Info) String() string {
	var b strings.Builder

	// ASCII art
	if i.asciiName != "" {
		b.WriteString(i.asciiName)
		b.WriteString("\n")
	}

	// App details
	if i.Name != "" {
		if i.GitVersion != "" {
			b.WriteString(fmt.Sprintf("%s v%s\n", i.Name, i.GitVersion))
		} else {
			b.WriteString(fmt.Sprintf("%s\n", i.Name))
		}
	}
	if i.Description != "" {
		b.WriteString(fmt.Sprintf("\n%s\n", i.Description))
	}
	if i.Website != "" {
		b.WriteString(fmt.Sprintf("\n%s\n", i.Website))
	}
	b.WriteString("\n")

	// Build info using tabwriter for aligned output
	w := tabwriter.NewWriter(&b, 0, 0, 1, ' ', 0)
	if i.GitCommit != "" {
		fmt.Fprintf(w, "Git Commit:\t%s\n", i.GitCommit)
	}
	if i.GitTreeState != "" {
		fmt.Fprintf(w, "Git Tree:\t%s\n", i.GitTreeState)
	}
	if i.BuildDate != "" {
		fmt.Fprintf(w, "Build Date:\t%s\n", i.BuildDate)
	}
	if i.BuiltBy != "" {
		fmt.Fprintf(w, "Built By:\t%s\n", i.BuiltBy)
	}
	w.Flush()

	b.WriteString("\n")

	return b.String()
}
