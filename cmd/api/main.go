package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/fms/fms/shared/config"
	"github.com/fms/fms/shared/logger"
	"github.com/fms/fms/shared/metrics"
)

type App struct {
	cfg    *config.Config
	logger *logger.Logger
	metrics *metrics.Metrics
	router *chi.Mux
}

func main() {
	ctx := context.Background()

	app, err := NewApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize app: %v\n", err)
		os.Exit(1)
	}

	if err := app.Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to run app: %v\n", err)
		os.Exit(1)
	}
}

func NewApp() (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	l := logger.Default()
	l.Info("Starting FMS API", "port", cfg.App.Port)

	m := metrics.NewMetrics()

	router := chi.NewRouter()
	app := &App{
		cfg:    cfg,
		logger: l,
		metrics: m,
		router: router,
	}

	app.setupRouter()
	return app, nil
}

func (a *App) setupRouter() {
	a.router.Use(middleware.RequestID)
	a.router.Use(middleware.RealIP)
	a.router.Use(a.metricsMiddleware)
	a.router.Use(middleware.Timeout(60 * time.Second))
	a.router.Use(middleware.Logger)
	a.router.Use(middleware.Recoverer)

	a.router.Get("/health", a.healthHandler)
	a.router.Handle("/metrics", promhttp.Handler())
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (a *App) metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rec := &statusRecorder{ResponseWriter: w, status: 200}

		next.ServeHTTP(rec, r)

		duration := time.Since(start).Seconds()
		a.metrics.HTTPRequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)

		statusStr := fmt.Sprintf("%dxx", rec.status/100)
		a.metrics.HTTPRequestsTotal.WithLabelValues(r.Method, r.URL.Path, statusStr).Inc()
	})
}

func (a *App) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (a *App) Run(ctx context.Context) error {
	server := &http.Server{
		Addr:    ":" + a.cfg.App.Port,
		Handler: a.router,
	}

	go func() {
		a.logger.Info("Server starting", "address", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Error("Server error", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	a.logger.Info("Shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	a.logger.Info("Server stopped")
	return nil
}