package middleware

import (
	"net/http"
	"time"

	"github.com/dsnikitin/gophermart/internal/pkg/logger"
)

type loggedResponseData struct {
	status int
	size   int
}

type loggedResponseWriter struct {
	http.ResponseWriter
	data *loggedResponseData
}

func (w *loggedResponseWriter) Write(b []byte) (int, error) {
	size, err := w.ResponseWriter.Write(b)
	w.data.size += size
	return size, err
}

func (w *loggedResponseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.data.status = statusCode
}

func Logger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		lw := loggedResponseWriter{
			ResponseWriter: w,
			data:           &loggedResponseData{},
		}

		h.ServeHTTP(&lw, r)

		logger.Log.Infow("Request handled",
			"uri", r.RequestURI,
			"method", r.Method,
			"status", lw.data.status,
			"duration", time.Since(start),
			"size", lw.data.size,
		)
	})
}
