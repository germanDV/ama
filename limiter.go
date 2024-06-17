package main

import (
	"fmt"
	"net/http"
)

func globalLimiter(limit int, countGetter func() (int, error)) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			current, err := countGetter()
			if err != nil {
				fmt.Printf("error getting current count: %s\n", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			if current >= limit {
				fmt.Printf("reached limit of %d for %s %s\n", current, r.Method, r.URL.Path)
				http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
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
				http.Error(w, "no id provided", http.StatusBadRequest)
				return
			}

			current, err := countGetter(questionnaireID)
			if err != nil {
				fmt.Printf("error getting current count: %s\n", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			if current >= limit {
				fmt.Printf("reached limit of %d for %s %s\n", current, r.Method, r.URL.Path)
				http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
				return
			}

			fmt.Printf("limit: %s %s %d/%d\n", r.Method, r.URL.Path, current+1, limit)
			next.ServeHTTP(w, r)
		})
	}
}
