package main

import (
	"context"
	"encoding/json"
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
	"github.com/fms/fms/shared/database"
	"github.com/fms/fms/shared/database/sqlc"
	"github.com/fms/fms/shared/logger"
	"github.com/fms/fms/shared/metrics"
	"github.com/fms/fms/services/vehicle"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	cfg            *config.Config
	logger         *logger.Logger
	metrics        *metrics.Metrics
	router         *chi.Mux
	pool           *pgxpool.Pool
	vehicleService vehicle.Service
}

func main() {
	ctx := context.Background()

	app, err := NewApp(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize app: %v\n", err)
		os.Exit(1)
	}

	if err := app.Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to run app: %v\n", err)
		os.Exit(1)
	}
}

func NewApp(ctx context.Context) (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	l := logger.Default()
	l.Info("Starting FMS API", "port", cfg.App.Port)

	m := metrics.NewMetrics()

	pool, err := database.NewPoolFactory(ctx, cfg.Database.URL, int32(cfg.Database.MaxOpenConns))
	if err != nil {
		return nil, fmt.Errorf("failed to create database pool: %w", err)
	}

	l.Info("Connected to database")

	queries := sqlc.New(pool)

	vehicleSvc, err := vehicle.NewService(&VehicleRepoAdapter{queries: queries})
	if err != nil {
		return nil, fmt.Errorf("failed to create vehicle service: %w", err)
	}

	router := chi.NewRouter()
	app := &App{
		cfg:            cfg,
		logger:         l,
		metrics:        m,
		router:         router,
		pool:           pool,
		vehicleService: vehicleSvc,
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
	a.router.Use(a.errorHandler)

	a.router.Get("/health", a.healthHandler)
	a.router.Handle("/metrics", promhttp.Handler())

	a.router.Route("/vehicles", func(r chi.Router) {
		r.Post("/", a.createVehicleHandler)
		r.Get("/", a.listVehiclesHandler)
		r.Get("/{id}", a.getVehicleHandler)
	})
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

func (a *App) errorHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				a.logger.Error("panic recovered", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (a *App) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (a *App) createVehicleHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Plate string `json:"plate"`
		Model string `json:"model"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	v, err := a.vehicleService.RegisterVehicle(r.Context(), req.Plate, req.Model)
	if err != nil {
		a.handleServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(v)
}

func (a *App) getVehicleHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	v, err := a.vehicleService.GetVehicle(r.Context(), id)
	if err != nil {
		a.handleServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func (a *App) listVehiclesHandler(w http.ResponseWriter, r *http.Request) {
	vehicles, err := a.vehicleService.ListVehicles(r.Context())
	if err != nil {
		a.handleServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(vehicles)
}

func (a *App) handleServiceError(w http.ResponseWriter, err error) {
	if database.IsNotFound(err) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
		return
	}

	if database.IsConstraintViolation(err) {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{"error": "constraint violation"})
		return
	}

	a.logger.Error("service error", "error", err)
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
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

	a.pool.Close()
	a.logger.Info("Database pool closed")

	a.logger.Info("Server stopped")
	return nil
}

// VehicleRepoAdapter adapts the SQLc Querier to the vehicle.Repository interface.
type VehicleRepoAdapter struct {
	queries *sqlc.Queries
}

func (a *VehicleRepoAdapter) Create(ctx context.Context, v *vehicle.Vehicle) error {
	params := sqlc.CreateVehicleParams{
		Plate:     v.Plate,
		Model:     v.Model,
		CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}

	err := a.queries.CreateVehicle(ctx, params)
	if err != nil {
		return database.WrapError(err)
	}

	return nil
}

func (a *VehicleRepoAdapter) GetByID(ctx context.Context, id string) (*vehicle.Vehicle, error) {
	uuid := pgtype.UUID{}
	if err := uuid.Scan(id); err != nil {
		return nil, &database.ErrNotFound{Err: fmt.Errorf("invalid UUID: %s", id)}
	}

	sqlcVehicle, err := a.queries.GetVehicleByID(ctx, uuid)
	if err != nil {
		return nil, database.WrapError(err)
	}

	return &vehicle.Vehicle{
		ID:        uuidToHex(sqlcVehicle.ID),
		Plate:     sqlcVehicle.Plate,
		Model:     sqlcVehicle.Model,
		CreatedAt: sqlcVehicle.CreatedAt.Time.Format(time.RFC3339),
	}, nil
}

func (a *VehicleRepoAdapter) GetAll(ctx context.Context) ([]*vehicle.Vehicle, error) {
	sqlcVehicles, err := a.queries.GetAllVehicles(ctx)
	if err != nil {
		return nil, database.WrapError(err)
	}

	vehicles := make([]*vehicle.Vehicle, len(sqlcVehicles))
	for i, v := range sqlcVehicles {
		vehicles[i] = &vehicle.Vehicle{
			ID:        uuidToHex(v.ID),
			Plate:     v.Plate,
			Model:     v.Model,
			CreatedAt: v.CreatedAt.Time.Format(time.RFC3339),
		}
	}

	return vehicles, nil
}

func uuidToHex(uuid pgtype.UUID) string {
	if !uuid.Valid {
		return ""
	}
	b := uuid.Bytes
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%12x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
