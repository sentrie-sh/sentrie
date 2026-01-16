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

package pack

import (
	"encoding/json"
	"slices"

	"github.com/Masterminds/semver/v3"
	"github.com/sentrie-sh/sentrie/ast"
)

type Pack struct {
	Pack     *PackFile
	Programs []*ast.Program
}

func NewPackFile(name string) *PackFile {
	return &PackFile{
		SchemaVersion: &SentrieSchema{Version: 1},
		Pack: &PackInformation{
			Name:    name,
			Version: semver.MustParse("0.0.1"),
		},
	}
}

type PackFile struct {
	SchemaVersion *SentrieSchema   `toml:"schema" json:"schema"`
	Pack          *PackInformation `toml:"pack" json:"pack"`
	Permissions   *Permissions     `toml:"permissions,omitempty" json:"permissions"`
	Engine        *Engine          `toml:"engine,omitempty" json:"engine"`
	Metadata      map[string]any   `toml:"metadata,omitempty" json:"metadata"`
	Location      string           `toml:"-" json:"-"`
}

type SentrieSchema struct {
	Version uint64 `toml:"version" json:"version"`
}

type PackInformation struct {
	Name        string            `toml:"name" json:"name"`
	Version     *semver.Version   `toml:"version" json:"version"`
	Description string            `toml:"description,omitempty" json:"description"`
	License     string            `toml:"license,omitempty" json:"license"`
	Repository  string            `toml:"repository,omitempty" json:"repository"`
	Authors     map[string]string `toml:"authors,omitempty" json:"authors"`
}

type Engine struct {
	Sentrie *semver.Constraints `toml:"sentrie" json:"sentrie"`
}

// MarshalJSON implements json.Marshaler for Engine to serialize semver.Constraints as string
// this is necessary because semver.Constraints does not implement json.Marshaler
func (e *Engine) MarshalJSON() ([]byte, error) {
	type Alias Engine
	aux := &struct {
		Sentrie string `json:"sentrie"`
		*Alias
	}{
		Alias: (*Alias)(e),
	}
	if e.Sentrie != nil {
		aux.Sentrie = e.Sentrie.String()
	}
	return json.Marshal(aux)
}

// UnmarshalJSON implements json.Unmarshaler for Engine to deserialize semver.Constraints from string
func (e *Engine) UnmarshalJSON(data []byte) error {
	type Alias Engine
	aux := &struct {
		Sentrie string `json:"sentrie"`
		*Alias
	}{
		Alias: (*Alias)(e),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	if aux.Sentrie != "" {
		constraint, err := semver.NewConstraint(aux.Sentrie)
		if err != nil {
			return err
		}
		e.Sentrie = constraint
	}
	return nil
}

type Permissions struct {
	FSRead []string `toml:"fs_read,omitempty" json:"fs_read"`
	Net    []string `toml:"net,omitempty" json:"net"`
	Env    []string `toml:"env,omitempty" json:"env"`
}

func (p *Permissions) CheckEnvAccess(name string) bool {
	return slices.Contains(p.Env, name)
}
