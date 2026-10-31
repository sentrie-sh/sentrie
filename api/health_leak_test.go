// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package api

import "net/http"

// The deepcheck body must not name the pack's filesystem path, but the
// validation and commit errors embed source spans that start with it.
func (s *APITestSuite) TestDeepHealthErrorsDoNotLeakPackPath() {
	for name, rules := range map[string]string{
		"validation": "  shape Alias string\n  shape Composed with Alias { a: string }\n" + healthyRules,
		"commit":     commitFailRules,
	} {
		api, dir := s.healthPackAPI(rules, true)
		rec := s.getHealth(api, "/health/deep")
		s.Equal(http.StatusServiceUnavailable, rec.Code, name)
		s.NotContains(rec.Body.String(), dir, name)
	}
}
