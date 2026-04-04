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

package loader

import (
	"os"
	"path/filepath"
)

func (s *LoaderTestSuite) writePackDir(content string) string {
	dir := s.T().TempDir()
	packFile := filepath.Join(dir, PackFileName)
	s.Require().NoError(os.WriteFile(packFile, []byte(content), 0o644))
	return dir
}
