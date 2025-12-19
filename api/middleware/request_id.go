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

package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type requestIdCtxKeyType struct{}

var requestIdCtxKey = requestIdCtxKeyType{}

func GetRequestIDFromRequest(req *http.Request) string {
	return req.Context().Value(requestIdCtxKey).(string)
}

func HasRequestIDInRequest(req *http.Request) bool {
	return req.Context().Value(requestIdCtxKey) != nil
}

func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = ensureRequestIDInRequest(r)
		next.ServeHTTP(w, r)
	})
}

func ensureRequestIDInRequest(r *http.Request) *http.Request {
	if HasRequestIDInRequest(r) {
		return r
	}
	return r.WithContext(context.WithValue(r.Context(), requestIdCtxKey, uuid.New().String()))
}
