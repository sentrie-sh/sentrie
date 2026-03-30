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

package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/sentrie-sh/sentrie/box"
	"github.com/stretchr/testify/require"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	defer func() { require.NoError(t, r.Close()) }()
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	fn()

	require.NoError(t, w.Close())
	var buf bytes.Buffer
	_, err = buf.ReadFrom(r)
	require.NoError(t, err)
	return buf.String()
}

func TestFormatAttachmentRecursesBoxedContainers(t *testing.T) {
	value := box.Map(map[string]box.Value{
		"items": box.List([]box.Value{box.Number(1), box.Number(2)}),
	})
	out := captureStdout(t, func() {
		formatAttachment("root", value, 0)
	})
	require.Contains(t, out, "root:")
	require.Contains(t, out, "items:")
	require.Contains(t, out, "- 1")
	require.Contains(t, out, "- 2")
}
