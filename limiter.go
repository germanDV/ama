package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/germandv/ama/internal/web"
)

func globalLimiter(limit int, countGetter func() (int, error)) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			current, err := countGetter()
			if err != nil {
				web.InternalError(w, err)
				return
			}

			if current >= limit {
				web.TooManyRequests(w, fmt.Sprintf("reached limit of %d for %s %s", current, r.Method, r.URL.Path))
				return
			}

			fmt.Printf("limit: %s %s %d/%d\n", r.Method, r.URL.Path, current+1, limit)
			next.ServeHTTP(w, r)
		})
	}
}

func idLimiter(limit int, countGetter func(id string) (int, error)) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			questionnaireID := r.PathValue("id")
			if questionnaireID == "" {
				web.BadRequest(w, errors.New("no id provided"))
				return
			}

			current, err := countGetter(questionnaireID)
			if err != nil {
				web.InternalError(w, err)
				return
			}

			if current >= limit {
				web.TooManyRequests(w, fmt.Sprintf("reached limit of %d for %s %s", current, r.Method, r.URL.Path))
				return
			}

			fmt.Printf("limit: %s %s %d/%d\n", r.Method, r.URL.Path, current+1, limit)
			next.ServeHTTP(w, r)
		})
	}
}
