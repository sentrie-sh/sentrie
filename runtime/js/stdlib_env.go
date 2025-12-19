// SPDX-License-Identifier: Apache-2.0

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

package js

import (
	"context"
	"os"
	"strings"

	"github.com/sentrie-sh/sentrie/pack"
)

func (ar *AliasRuntime) setupEnvStdLib(_ context.Context, pack *pack.PackFile) error {
	env := ar.VM.NewObject()

	// range through the os.Environ() and add to env
	for _, kv := range os.Environ() {
		key, value, found := strings.Cut(kv, "=")
		if !found {
			continue // skip malformed environment variables
		}

		if pack != nil && pack.Permissions != nil && pack.Permissions.CheckEnvAccess(key) {
			if err := env.Set(key, value); err != nil {
				return err
			}
		}
	}

	if err := ar.VM.Set("env", env); err != nil {
		return err
	}
	return nil
}
