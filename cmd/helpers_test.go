// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2026 Binaek Sarkar
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
)

func (s *CmdTestSuite) captureStdout(fn func()) string {
	s.T().Helper()
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	s.Require().NoError(err)
	defer func() { s.Require().NoError(r.Close()) }()
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()
	fn()
	s.Require().NoError(w.Close())
	var buf bytes.Buffer
	_, err = buf.ReadFrom(r)
	s.Require().NoError(err)
	return buf.String()
}
