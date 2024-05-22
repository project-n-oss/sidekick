package api

import (
	"context"
	"net/http"
)

func (a *Api) healthMiddleware(handler http.Handler, f func(ctx context.Context) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			if err := f(r.Context()); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			return
		}
		handler.ServeHTTP(w, r)
	}
}
