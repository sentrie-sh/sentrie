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
