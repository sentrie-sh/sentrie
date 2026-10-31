// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/loader"
	runtimepkg "github.com/sentrie-sh/sentrie/runtime"
)

// healthMux routes the two health endpoints the way Setup does.
func healthMux(api *HTTPAPI) http.Handler {
	mux := http.NewServeMux()
	api.registerHealthRoutes(mux)
	return mux
}

func (s *APITestSuite) getHealth(api *HTTPAPI, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	healthMux(api).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

// decodeJSON decodes a response body into a generic map, so the tests check the
// wire shape (field names, nulls) and not the Go structs.
func (s *APITestSuite) decodeJSON(rec *httptest.ResponseRecorder) map[string]any {
	var body map[string]any
	s.Require().NoError(json.Unmarshal(rec.Body.Bytes(), &body), rec.Body.String())
	return body
}

const (
	healthyRules = "  rule r = { yield true }\n  export decision of r"
	cyclicRules  = "  rule ra = { yield rb }\n  rule rb = { yield ra }\n  export decision of ra"
)

// healthPackAPI builds an API over a one-policy pack in a temp dir. With
// validate false the index is left unvalidated; with validate true a pack that
// fails validation yields an API whose index recorded the error.
func (s *APITestSuite) healthPackAPI(body string, validate bool) (*HTTPAPI, string) {
	ctx := s.T().Context()
	dir := s.T().TempDir()
	files := map[string]string{
		loader.PackFileName: "[schema]\nversion = 1\n\n[pack]\nname = \"health-pack\"\nversion = \"1.2.3\"\n",
		"policy.sentrie":    "namespace health/test\n\npolicy p {\n" + body + "\n}\n",
	}
	for name, content := range files {
		s.Require().NoError(os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644))
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
	if validate {
		_ = idx.Validate(ctx)
	}
	exec, err := runtimepkg.NewExecutor(idx)
	s.Require().NoError(err)
	s.T().Cleanup(func() { _ = exec.Close() })
	return NewHTTPAPI(exec), dir
}

func (s *APITestSuite) TestHealthSchema() {
	rec := s.getHealth(NewHTTPAPI(nil), "/health")
	s.Equal(http.StatusOK, rec.Code)
	s.Equal("application/json", rec.Header().Get("Content-Type"))
	s.Equal("no-store", rec.Header().Get("Cache-Control"))

	body := s.decodeJSON(rec)
	s.Len(body, 3, "exactly check_id, status and time")
	s.Equal("healthy", body["status"], "status is kept for existing callers")
	checkID, ok := body["check_id"].(string)
	s.Require().True(ok)
	s.NotEmpty(checkID)
	ts, ok := body["time"].(string)
	s.Require().True(ok)
	parsed, err := time.Parse(time.RFC3339, ts)
	s.Require().NoError(err)
	s.WithinDuration(time.Now(), parsed, time.Minute)
}

func (s *APITestSuite) TestHealthCheckIDIsFreshPerResponse() {
	api := NewHTTPAPI(nil)
	seen := map[string]bool{}
	for range 20 {
		id := s.decodeJSON(s.getHealth(api, "/health"))["check_id"].(string)
		s.False(seen[id], "check_id repeated")
		seen[id] = true
	}
}

func (s *APITestSuite) TestHealthIsLivenessOnly() {
	// No executor at all: the process is still alive and answers.
	s.Equal(http.StatusOK, s.getHealth(NewHTTPAPI(nil), "/health").Code)
}

func (s *APITestSuite) TestHealthRejectsOtherMethods() {
	rec := httptest.NewRecorder()
	healthMux(NewHTTPAPI(nil)).ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/health/deep", nil))
	s.Equal(http.StatusMethodNotAllowed, rec.Code)
}

func (s *APITestSuite) TestDeepHealthHealthyPack() {
	rec := s.getHealth(s.examplePackHTTPAPI(), "/health/deep")
	s.Equal(http.StatusOK, rec.Code, rec.Body.String())
	s.Equal("application/json", rec.Header().Get("Content-Type"))
	s.Equal("no-store", rec.Header().Get("Cache-Control"))

	body := s.decodeJSON(rec)
	s.Equal("healthy", body["status"])
	_, err := time.Parse(time.RFC3339, body["time"].(string))
	s.Require().NoError(err)
	s.Equal([]any{}, body["errors"], "errors is an empty array, not null")

	comps := body["components"].(map[string]any)
	s.ElementsMatch([]string{"index", "executor", "js_runtime", "pack"}, keysOf(comps),
		"only components the process can observe")

	idx := comps["index"].(map[string]any)
	s.Equal("healthy", idx["status"])
	s.Equal(true, idx["validated"])
	s.Equal(true, idx["committed"])
	s.Contains(idx, "validation_error")
	s.Nil(idx["validation_error"])
	s.Contains(idx, "commit_error")
	s.Nil(idx["commit_error"])
	stats := idx["statistics"].(map[string]any)
	s.ElementsMatch([]string{"namespaces", "policies", "programs", "rules", "shapes"}, keysOf(stats))
	for name, v := range stats {
		s.GreaterOrEqual(v.(float64), float64(0), name)
	}
	s.GreaterOrEqual(stats["namespaces"].(float64), float64(1))
	s.GreaterOrEqual(stats["policies"].(float64), float64(1))
	s.GreaterOrEqual(stats["programs"].(float64), float64(1))
	s.GreaterOrEqual(stats["rules"].(float64), float64(1))

	exec := comps["executor"].(map[string]any)
	s.Equal("healthy", exec["status"])
	s.Equal(true, exec["initialized"])
	s.NotContains(exec, "caches", "no observable cache state to report")

	js := comps["js_runtime"].(map[string]any)
	s.Equal("healthy", js["status"])
	s.Equal(true, js["registry_initialized"])
	s.NotContains(js, "instance_pools")
	mods := js["builtin_modules"].(map[string]any)
	names := toStrings(mods["modules"].([]any))
	s.Equal(float64(len(names)), mods["registered"])
	s.Contains(names, "@sentrie/uuid")
	s.Contains(names, "@sentrie/math")
	s.Contains(names, "@sentrie/js", "TypeScript built-ins are listed with the Go ones")
	s.IsIncreasing(names, "module names are sorted")

	pack := comps["pack"].(map[string]any)
	s.Equal("healthy", pack["status"])
	s.Equal(true, pack["loaded"])
	s.Equal(true, pack["accessible"])
	s.NotContains(pack, "location", "the filesystem path is not exposed")
	s.NotContains(comps, "operational")
}

func (s *APITestSuite) TestDeepHealthDoesNotLeakPackPath() {
	api := s.examplePackHTTPAPI()
	rec := s.getHealth(api, "/health/deep")
	s.NotContains(rec.Body.String(), api.executor.Index().Pack.Location)
}

func (s *APITestSuite) TestDeepHealthIndexNotValidated() {
	api, _ := s.healthPackAPI(healthyRules, false)
	rec := s.getHealth(api, "/health/deep")
	s.Equal(http.StatusServiceUnavailable, rec.Code)

	body := s.decodeJSON(rec)
	s.Equal("unhealthy", body["status"])
	idx := body["components"].(map[string]any)["index"].(map[string]any)
	s.Equal("unhealthy", idx["status"])
	s.Equal(false, idx["validated"])
	s.Equal(false, idx["committed"])
	s.Equal([]any{map[string]any{"component": "index", "message": "index has not been validated"}}, body["errors"])
	// The other components are still reported and healthy.
	comps := body["components"].(map[string]any)
	s.Equal("healthy", comps["executor"].(map[string]any)["status"])
	s.Equal("healthy", comps["pack"].(map[string]any)["status"])
}

func (s *APITestSuite) TestDeepHealthValidationError() {
	api, _ := s.healthPackAPI(cyclicRules, true)
	rec := s.getHealth(api, "/health/deep")
	s.Equal(http.StatusServiceUnavailable, rec.Code)

	body := s.decodeJSON(rec)
	s.Equal("unhealthy", body["status"])
	idx := body["components"].(map[string]any)["index"].(map[string]any)
	s.Equal("unhealthy", idx["status"])
	s.Equal(false, idx["validated"])
	msg, ok := idx["validation_error"].(string)
	s.Require().True(ok, "validation_error says validation failed")
	s.Equal(msgIndexValidationFailed, msg, "the index's own text carries the pack path and stays in the log")
	s.Nil(idx["commit_error"])
	errs := body["errors"].([]any)
	s.Require().Len(errs, 1)
	s.Equal(map[string]any{"component": "index", "message": msg}, errs[0])
}

func (s *APITestSuite) TestDeepHealthPackDirectoryGone() {
	api, dir := s.healthPackAPI(healthyRules, true)
	s.Equal(http.StatusOK, s.getHealth(api, "/health/deep").Code)

	s.Require().NoError(os.RemoveAll(dir))
	rec := s.getHealth(api, "/health/deep")
	s.Equal(http.StatusServiceUnavailable, rec.Code)
	body := s.decodeJSON(rec)
	pack := body["components"].(map[string]any)["pack"].(map[string]any)
	s.Equal("unhealthy", pack["status"])
	s.Equal(true, pack["loaded"])
	s.Equal(false, pack["accessible"])
	s.Equal("pack directory is not accessible", pack["error"])
	s.NotContains(rec.Body.String(), dir)
	s.Equal([]any{map[string]any{"component": "pack", "message": "pack directory is not accessible"}}, body["errors"])
}

func (s *APITestSuite) TestDeepHealthWithoutExecutor() {
	rec := s.getHealth(NewHTTPAPI(nil), "/health/deep")
	s.Equal(http.StatusServiceUnavailable, rec.Code)
	body := s.decodeJSON(rec)
	s.Equal("unhealthy", body["status"])
	comps := body["components"].(map[string]any)
	s.Equal(false, comps["executor"].(map[string]any)["initialized"])
	s.Equal(false, comps["pack"].(map[string]any)["loaded"])
	s.Equal(false, comps["js_runtime"].(map[string]any)["registry_initialized"])
	var components []string
	for _, e := range body["errors"].([]any) {
		components = append(components, e.(map[string]any)["component"].(string))
	}
	s.ElementsMatch([]string{"executor", "index", "js_runtime", "pack"}, components)
}

func (s *APITestSuite) TestSetupServesHealthRoutes() {
	api := s.examplePackHTTPAPI()
	s.Require().NoError(api.Setup(s.T().Context(), 0, []string{"127.0.0.1"}))
	s.T().Cleanup(func() { _ = api.StopServer(s.T().Context()) })
	go api.StartServer(s.T().Context(), 0, nil)

	base := "http://" + api.listeners[0].Listener.Addr().String()
	for _, path := range []string{"/health", "/health/deep"} {
		var resp *http.Response
		var err error
		s.Require().Eventually(func() bool {
			resp, err = http.Get(base + path)
			return err == nil
		}, 5*time.Second, 20*time.Millisecond, path)
		s.Equal(http.StatusOK, resp.StatusCode, path)
		_ = resp.Body.Close()
	}
}

// commitFailRules is a pack whose validation passes but whose commit fails: a
// shape cannot compose with an alias of another shape.
const commitFailRules = healthyRules + "\n}\nshape Alias string\nshape Composed with Alias { a: string }\npolicy q {\n" + healthyRules

func (s *APITestSuite) TestDeepHealthCommitError() {
	api, _ := s.healthPackAPI(commitFailRules, true)
	rec := s.getHealth(api, "/health/deep")
	s.Equal(http.StatusServiceUnavailable, rec.Code)

	body := s.decodeJSON(rec)
	s.Equal("unhealthy", body["status"])
	idx := body["components"].(map[string]any)["index"].(map[string]any)
	s.Equal("unhealthy", idx["status"])
	s.Equal(true, idx["validated"], "validation itself passed")
	s.Equal(false, idx["committed"])
	s.Nil(idx["validation_error"], "a commit failure is reported once, under commit_error")
	msg, ok := idx["commit_error"].(string)
	s.Require().True(ok, "commit_error says the commit failed")
	s.Equal(msgIndexCommitFailed, msg, "the index's own text carries the pack path and stays in the log")
	s.Equal([]any{map[string]any{"component": "index", "message": msg}}, body["errors"])
}

func (s *APITestSuite) TestHealthRoutesRejectPost() {
	mux := healthMux(NewHTTPAPI(nil))
	for _, path := range []string{"/health", "/health/deep"} {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, path, nil))
		s.Equal(http.StatusMethodNotAllowed, rec.Code, path)
		s.Contains(rec.Header().Get("Allow"), http.MethodGet, path)
	}
}

func (s *APITestSuite) TestDeepHealthConcurrentWithValidate() {
	api, _ := s.healthPackAPI(healthyRules, false)
	idx := api.executor.Index()
	ctx := s.T().Context()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = idx.Validate(ctx)
	}()
	for range 4 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 25 {
				rec := httptest.NewRecorder()
				healthMux(api).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health/deep", nil))
				// Either side of the Validate: unvalidated (503) or healthy (200).
				if rec.Code != http.StatusOK && rec.Code != http.StatusServiceUnavailable {
					s.Failf("unexpected status", "%d", rec.Code)
				}
			}
		}()
	}
	wg.Wait()
	s.Equal(http.StatusOK, s.getHealth(api, "/health/deep").Code, "healthy once Validate has finished")
}

func keysOf(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func toStrings(in []any) []string {
	out := make([]string, len(in))
	for i, v := range in {
		out[i] = v.(string)
	}
	return out
}
