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

package pack

import "github.com/binaek/sentra/ast"

type Pack struct {
	Pack     *PackFile
	Programs []*ast.Program
}

type PackFile struct {
	SchemaVersion string            `toml:"schema_version"`
	Name          string            `toml:"name"`
	Version       string            `toml:"version,omitempty"`
	Description   string            `toml:"description,omitempty"`
	License       string            `toml:"license,omitempty"`
	Repository    string            `toml:"repository,omitempty"`
	Engines       Engines           `toml:"engines"`
	Authors       map[string]string `toml:"authors,omitempty"`
	Permissions   Permissions       `toml:"permissions"`
	Metadata      map[string]any    `toml:"metadata,omitempty"`
	Location      string            `toml:"-"`
}

type Engines struct {
	Sentrie string `toml:"sentrie"`
}

type Permissions struct {
	FSRead []string `toml:"fs_read,omitempty"`
	Net    []string `toml:"net,omitempty"`
}
