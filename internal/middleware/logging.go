package middleware

import (
	"log"
	"net/http"
	"time"
)

// Custom ResponseWriter wrapper to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.written = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

// Middleware that logs HTTP requests
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Wrap the response writer
		wrapped := &responseWriter{ResponseWriter: w, statusCode: 200, written: false}

		// Record start time
		start := time.Now()

		// Call next handler
		next.ServeHTTP(wrapped, r)

		// Calculate duration and log
		duration := time.Since(start)
		requestID := RequestIDFromContext(r.Context())

		log.Printf("%s %s %d %v %s\n",
			r.Method, r.URL.Path, wrapped.statusCode, duration, requestID)
	})
}
