package requestlog

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type Attributes func(*http.Request) []any

type responseWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (w *responseWriter) WriteHeader(status int) {
	if w.status != 0 {
		return
	}
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *responseWriter) Write(body []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	n, err := w.ResponseWriter.Write(body)
	w.bytes += n
	return n, err
}

func Middleware(logger *slog.Logger, attributes Attributes, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		requestID := strings.TrimSpace(r.Header.Get("X-Request-ID"))
		if requestID == "" {
			requestID = newRequestID()
		}
		w.Header().Set("X-Request-ID", requestID)
		wrapped := &responseWriter{ResponseWriter: w}

		defer func() {
			if recovered := recover(); recovered != nil {
				logger.Error("request panic", "request_id", requestID, "method", r.Method, "path", r.URL.Path, "error", recovered)
				if wrapped.status == 0 {
					http.Error(wrapped, "internal server error", http.StatusInternalServerError)
				}
			}
			status := wrapped.status
			if status == 0 {
				status = http.StatusOK
			}
			fields := []any{"request_id", requestID, "method", r.Method, "path", r.URL.Path, "status", status, "duration_ms", time.Since(started).Milliseconds(), "response_bytes", wrapped.bytes}
			if attributes != nil {
				fields = append(fields, attributes(r)...)
			}
			logger.Info("http request", fields...)
		}()

		next.ServeHTTP(wrapped, r)
	})
}

func newRequestID() string {
	var value [8]byte
	if _, err := rand.Read(value[:]); err != nil {
		return hex.EncodeToString([]byte(time.Now().Format("150405.000000")))
	}
	return hex.EncodeToString(value[:])
}
