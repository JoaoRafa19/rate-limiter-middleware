package api

import (
	"fmt"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

func RateLimiter(limiter *rate.Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.Allow() {
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code //capture
	s.ResponseWriter.WriteHeader(code)
}

func Logger() func(http.Handler) http.Handler {

	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			h.ServeHTTP(rec, r) // default status 200
			duration := time.Since(start)
			fmt.Printf("[%d] [%s] %s %s took %s\n", rec.status, r.Method, r.URL.Path, r.Proto, duration)
		})
	}
}
