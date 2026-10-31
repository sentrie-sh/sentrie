// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/loader"
	runtimepkg "github.com/sentrie-sh/sentrie/runtime"
)

// The decision endpoint evaluates a policy that reads pack data loaded once at
// startup, and every request sees the same data.
func (s *APITestSuite) TestHandleDecisionReadsPackData() {
	ctx := s.T().Context()
	dir := s.T().TempDir()
	files := map[string]string{
		loader.PackFileName: "[schema]\nversion = 1\n\n[pack]\nname = \"apidata\"\nversion = \"0.0.1\"\n\n[data]\nroles = \"roles.json\"\n",
		"roles.json":        `{"admin": ["read", "write"], "viewer": ["read"]}`,
		"policy.sentrie": `namespace data/api
policy p {
  fact user: string
  data roles
  rule canWrite = "write" in roles[user]
  export decision of canWrite
}
`,
	}
	for name, content := range files {
		s.Require().NoError(os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600))
	}

	packFile, err := loader.LoadPack(ctx, dir)
	s.Require().NoError(err)
	programs, err := loader.LoadPrograms(ctx, packFile)
	s.Require().NoError(err)
	idx := index.CreateIndex()
	s.Require().NoError(idx.SetPack(ctx, packFile))
	for _, program := range programs {
		s.Require().NoError(idx.AddProgram(ctx, program))
	}
	s.Require().NoError(idx.Validate(ctx))
	exec, err := runtimepkg.NewExecutor(idx)
	s.Require().NoError(err)
	defer func() { _ = exec.Close() }()
	api := NewHTTPAPI(exec)

	for _, tc := range []struct {
		user string
		want bool
	}{{"admin", true}, {"viewer", false}, {"admin", true}} {
		rec := s.postDecision(api, "/decision/data/api/p/canWrite", `{"facts": {"user": "`+tc.user+`"}}`)
		s.Require().Equal(http.StatusOK, rec.Code, rec.Body.String())
		var payload struct {
			Decisions []struct {
				Decision struct {
					Value any `json:"Value"`
				} `json:"Decision"`
			} `json:"decisions"`
		}
		s.Require().NoError(json.Unmarshal(rec.Body.Bytes(), &payload), rec.Body.String())
		s.Require().Len(payload.Decisions, 1, rec.Body.String())
		s.Equal(tc.want, payload.Decisions[0].Decision.Value, "user %s", tc.user)
	}
}
