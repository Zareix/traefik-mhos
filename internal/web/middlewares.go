package web

import (
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

type wrappedWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *wrappedWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wrapped := &wrappedWriter{w, http.StatusOK}
		start := time.Now()
		next.ServeHTTP(wrapped, r)
		log.Info().Msgf("%s %s %s", r.Method, r.URL.Path, time.Since(start).String())
	})
}
