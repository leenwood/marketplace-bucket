package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"

	"github.com/marketplace/marketplace-bucket/internal/platform/logger"
	"github.com/marketplace/marketplace-bucket/internal/platform/metrics"
)

// Chain applies middlewares to h in outermost-to-innermost order.
func Chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Recover catches panics, logs the stack trace, and returns a 500.
func Recover(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger.FromContext(r.Context(), log).Error("panic recovered",
						"panic", rec,
						"stack", string(debug.Stack()),
					)
					http.Error(w, "internal server error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// Logger logs each request and records Prometheus RED metrics.
// It also injects the OTel trace_id into the context so all handler logs carry it.
func Logger(log *slog.Logger, m *metrics.Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			span := trace.SpanFromContext(r.Context())
			traceID := span.SpanContext().TraceID().String()
			ctx := logger.WithTraceID(r.Context(), traceID)
			r = r.WithContext(ctx)

			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(rw, r)

			duration := time.Since(start)
			status := strconv.Itoa(rw.statusCode)

			l := logger.FromContext(r.Context(), log)
			if rid := w.Header().Get("X-Request-ID"); rid != "" {
				l = l.With("request_id", rid)
			}
			l.Info("HTTP request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", status,
				"duration_ms", duration.Milliseconds(),
			)

			m.HTTPRequestsTotal.WithLabelValues(r.Method, r.URL.Path, status).Inc()
			m.HTTPRequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration.Seconds())
		})
	}
}

// RequestID reads X-Request-ID from the incoming request or generates a new UUID,
// writes it to the response header, and injects it into the request context.
func RequestID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := r.Header.Get("X-Request-ID")
			if id == "" {
				id = uuid.New().String()
			}
			w.Header().Set("X-Request-ID", id)
			ctx := logger.WithRequestID(r.Context(), id)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// MaxBodySize rejects requests whose body exceeds n bytes with a 413.
func MaxBodySize(n int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, n)
			next.ServeHTTP(w, r)
		})
	}
}
