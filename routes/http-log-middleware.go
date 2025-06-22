package routes

import (
	"log"
	"net/http"
	"strings"
	"time"
)

// ResponseWriter wrapper to capture status code and response size
type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	responseSize int64
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.statusCode == 0 {
		rw.statusCode = 200
	}
	size, err := rw.ResponseWriter.Write(b)
	rw.responseSize += int64(size)
	return size, err
}

// Add Flush method to implement http.Flusher interface
func (rw *responseWriter) Flush() {
	if flusher, ok := rw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// HTTP logging middleware
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Get client IP (handle X-Forwarded-For header)
		clientIP := r.RemoteAddr
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			clientIP = strings.Split(forwarded, ",")[0]
		} else if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
			clientIP = realIP
		}

		// Wrap the response writer to capture status code and response size
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     0,
			responseSize:   0,
		}

		// Call the next handler
		next.ServeHTTP(wrapped, r)

		// Calculate duration
		duration := time.Since(start)

		// Log the request
		log.Printf("[HTTP] %s %s %d %d bytes %v %s \"%s\"",
			r.Method,
			r.URL.Path,
			wrapped.statusCode,
			wrapped.responseSize,
			duration,
			clientIP,
			r.UserAgent(),
		)
	})
}
