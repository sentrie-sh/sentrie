// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"fmt"
	"sync"
)

// Many evaluations read the same shared data value at once, including through
// a module call that receives a copy. Run under -race this proves a read does
// not write to the shared value.
func (s *RuntimeTestSuite) TestPackDataConcurrentReads() {
	ctx := s.T().Context()
	exec, err := s.buildPackDataExecutor(`roles = "roles.json"`+"\n", map[string]string{
		"roles.json": packDataRolesJSON,
		"mut.ts": `export function mutate(x: any): number {
  x.admin[0] = "changed";
  x.extra = 1;
  return Object.keys(x).length;
}
`,
		"policy.sentrie": `namespace data/test

policy p {
  fact user: string
  data roles
  use { mutate } from "./mut" as m

  rule canWrite = "write" in roles[user]
  rule copy = m.mutate(roles) == 3 and roles["admin"][0] == "read" and roles["extra"] is not defined
  export decision of canWrite
  export decision of copy
}
`,
	})
	s.Require().NoError(err)
	defer func() { _ = exec.Close() }()

	const workers, rounds = 16, 25
	errs := make(chan error, workers*rounds*2)
	var wg sync.WaitGroup
	for w := range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			user := "admin"
			if w%2 == 1 {
				user = "viewer"
			}
			for range rounds {
				out, err := exec.ExecRule(ctx, "data/test", "p", "canWrite", map[string]any{"user": user})
				if err != nil {
					errs <- err
					continue
				}
				if want := user == "admin"; out.Decision.Value.Any() != want {
					errs <- fmt.Errorf("canWrite(%s) = %v, want %v", user, out.Decision.Value.Any(), want)
				}
				out, err = exec.ExecRule(ctx, "data/test", "p", "copy", map[string]any{"user": user})
				if err != nil {
					errs <- err
					continue
				}
				if out.Decision.Value.Any() != true {
					errs <- fmt.Errorf("copy = %v, want true", out.Decision.Value.Any())
				}
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		s.NoError(err)
	}
}
