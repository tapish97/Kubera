package requestlog

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Attributes func(*http.Request) []any

type contextKey string

const metadataKey contextKey = "request-log-metadata"

type metadata struct {
	requestID string
	mu        sync.Mutex
	fields    []any
}

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
	slowThreshold := SlowThreshold()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		requestID := strings.TrimSpace(r.Header.Get("X-Request-ID"))
		if requestID == "" {
			requestID = newRequestID()
		}
		w.Header().Set("X-Request-ID", requestID)
		wrapped := &responseWriter{ResponseWriter: w}
		meta := &metadata{requestID: requestID}
		r = r.WithContext(context.WithValue(r.Context(), metadataKey, meta))

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
			duration := time.Since(started)
			fields := []any{"request_id", requestID, "method", r.Method, "path", r.URL.Path, "status", status, "duration_ms", duration.Milliseconds(), "response_bytes", wrapped.bytes}
			meta.mu.Lock()
			fields = append(fields, meta.fields...)
			meta.mu.Unlock()
			if attributes != nil {
				fields = append(fields, attributes(r)...)
			}
			if duration >= slowThreshold {
				logger.Warn("slow http request", fields...)
			} else {
				logger.Info("http request", fields...)
			}
		}()

		next.ServeHTTP(wrapped, r)
	})
}

// AddAttributes adds fields to the final request log entry. It is safe to call
// from middleware and handlers while the request is active.
func AddAttributes(ctx context.Context, fields ...any) {
	meta, ok := ctx.Value(metadataKey).(*metadata)
	if !ok {
		return
	}
	meta.mu.Lock()
	meta.fields = append(meta.fields, fields...)
	meta.mu.Unlock()
}

func RequestIDFromContext(ctx context.Context) string {
	meta, ok := ctx.Value(metadataKey).(*metadata)
	if !ok {
		return ""
	}
	return meta.requestID
}

func SlowThreshold() time.Duration {
	const fallback = 750 * time.Millisecond
	value := strings.TrimSpace(os.Getenv("SLOW_REQUEST_MS"))
	if value == "" {
		return fallback
	}
	milliseconds, err := strconv.Atoi(value)
	if err != nil || milliseconds <= 0 {
		return fallback
	}
	return time.Duration(milliseconds) * time.Millisecond
}

func newRequestID() string {
	var value [8]byte
	if _, err := rand.Read(value[:]); err != nil {
		return hex.EncodeToString([]byte(time.Now().Format("150405.000000")))
	}
	return hex.EncodeToString(value[:])
}
