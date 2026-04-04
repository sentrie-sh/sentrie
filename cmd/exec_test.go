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

import "github.com/sentrie-sh/sentrie/box"

func (s *CmdTestSuite) TestFormatAttachmentRecursesBoxedContainers() {
	value := box.Map(map[string]box.Value{
		"items": box.List([]box.Value{box.Number(1), box.Number(2)}),
	})
	out := s.captureStdout(func() {
		formatAttachment("root", value, 0)
	})
	s.Contains(out, "root:")
	s.Contains(out, "items:")
	s.Contains(out, "- 1")
	s.Contains(out, "- 2")
}
