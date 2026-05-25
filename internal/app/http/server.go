package apphttp

import (
	"log/slog"
	"net/http"
	"net/http/pprof"

	"github.com/marketplace/marketplace-bucket/internal"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/marketplace/marketplace-bucket/internal/app/http/handler"
	"github.com/marketplace/marketplace-bucket/internal/app/http/middleware"
	"github.com/marketplace/marketplace-bucket/internal/platform/metrics"
)

type Deps struct {
	Log     *slog.Logger
	Metrics *metrics.Metrics
	Handler *handler.Handler
	Ping    func() error
}

func NewServer(cfg internal.HTTPConfig, deps Deps) *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	mux.HandleFunc("GET /ready", func(w http.ResponseWriter, r *http.Request) {
		if deps.Ping != nil {
			if err := deps.Ping(); err != nil {
				deps.Log.Error("readiness check failed", "error", err)
				http.Error(w, `{"status":"unavailable"}`, http.StatusServiceUnavailable)
				return
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ready"}`))
	})

	mux.Handle("GET /metrics", promhttp.HandlerFor(deps.Metrics.Registry(), promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	}))

	mux.HandleFunc("/swagger/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "docs/swagger.yaml")
	})

	if cfg.PprofEnabled {
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}

	deps.Handler.RegisterRoutes(mux)

	otelMiddleware := func(next http.Handler) http.Handler {
		return otelhttp.NewHandler(next, "server",
			otelhttp.WithFilter(func(r *http.Request) bool {
				return r.URL.Path != "/metrics" && r.URL.Path != "/health"
			}),
		)
	}

	chain := middleware.Chain(mux,
		otelMiddleware,
		middleware.Recover(deps.Log),
		middleware.Logger(deps.Log, deps.Metrics),
		middleware.RequestID(),
		middleware.MaxBodySize(1<<20),
	)

	return &http.Server{
		Addr:         cfg.Addr,
		Handler:      chain,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}
}
