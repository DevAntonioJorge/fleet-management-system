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
	"github.com/fms/fms/shared/messaging"
	"github.com/fms/fms/services/telemetry"
	"github.com/fms/fms/services/alert"
	"github.com/fms/fms/services/vehicle"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/google/uuid"
	"strconv"
)

type App struct {
	cfg            *config.Config
	logger         *logger.Logger
	metrics        *metrics.Metrics
	router         *chi.Mux
	pool           *pgxpool.Pool
	vehicleService vehicle.Service
	publisher      messaging.Publisher
	subscriber     messaging.Subscriber
	telemetryService telemetry.Service
	alertService     alert.Service
	consumerCancel    context.CancelFunc
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

	// messaging publisher & subscriber
	pubFactory := messaging.NewRabbitMQPublisherFactory()
	publisher, err := pubFactory.CreatePublisher(cfg.Broker.RabbitMQ.URL, cfg.Broker.RabbitMQ.Exchange)
	if err != nil {
		return nil, fmt.Errorf("failed to create publisher: %w", err)
	}

	subFactory := messaging.NewRabbitMQSubscriberFactory()
	subscriber, err := subFactory.CreateSubscriber(cfg.Broker.RabbitMQ.URL, cfg.Broker.RabbitMQ.Queue, cfg.Broker.RabbitMQ.Exchange)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscriber: %w", err)
	}

	vehicleSvc, err := vehicle.NewService(&VehicleRepoAdapter{queries: queries})
	if err != nil {
		return nil, fmt.Errorf("failed to create vehicle service: %w", err)
	}
	// alert service
	alertSvc, err := alert.NewService(&AlertRepoAdapter{queries: queries})
	if err != nil {
		return nil, fmt.Errorf("failed to create alert service: %w", err)
	}

	// telemetry service
	telemetrySvc, err := telemetry.NewService(&TelemetryRepoAdapter{queries: queries}, publisher)
	if err != nil {
		return nil, fmt.Errorf("failed to create telemetry service: %w", err)
	}

	router := chi.NewRouter()
	app := &App{
		cfg:              cfg,
		logger:           l,
		metrics:          m,
		router:           router,
		pool:             pool,
		vehicleService:   vehicleSvc,
		publisher:        publisher,
		subscriber:       subscriber,
		telemetryService: telemetrySvc,
		alertService:     alertSvc,
	}

	// start consumer for telemetry using a cancelable context so shutdown can stop it
	subCtx, subCancel := context.WithCancel(context.Background())
	app.consumerCancel = subCancel

	subscribeErr := app.subscriber.Subscribe(subCtx, "raw", messaging.CreateTelemetryHandler(func(ctx context.Context, tm messaging.TelemetryMessage) error {
		// use stable event id from the message for idempotency
		id := tm.EventID
		if id == "" {
			// if absent, generate one (backwards compatibility)
			id = uuid.New().String()
		}

		ev := &telemetry.TelemetryEvent{
			ID:        id,
			VehicleID: tm.VehicleID,
			Timestamp: tm.Timestamp.Format(time.RFC3339),
			Latitude:  tm.Latitude,
			Longitude: tm.Longitude,
			Speed:     tm.Speed,
			FuelLevel: tm.FuelLevel,
		}

		// Persist telemetry. If it fails with a constraint violation it already exists.
		if err := app.telemetryService.Persist(ctx, ev); err != nil {
			if database.IsConstraintViolation(err) {
				// already exists — do not create alerts again
				app.metrics.TelemetryEventsProcessedTotal.Inc()
				return nil
			}
			app.metrics.TelemetryEventsFailedTotal.Inc()
			return err
		}

		// Only create alert when this was a fresh insert and speed exceeded threshold
		if tm.Speed > 120 {
			_, err := app.alertService.CreateAlert(ctx, tm.VehicleID, "speeding", fmt.Sprintf("Speed %.1f km/h exceeded limit", tm.Speed))
			if err != nil {
				app.metrics.TelemetryEventsFailedTotal.Inc()
				return err
			}
			app.metrics.AlertsGeneratedTotal.Inc()
		}

		app.metrics.TelemetryEventsProcessedTotal.Inc()
		return nil
	}))

	if subscribeErr != nil {
		// ensure resources are cleaned up on failure
		publisher.Close()
		subscriber.Close()
		return nil, fmt.Errorf("failed to subscribe to telemetry topic: %w", subscribeErr)
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

    a.router.Post("/telemetry", a.ingestTelemetryHandler)
    a.router.Get("/vehicles/{id}/telemetry", a.getVehicleTelemetryHandler)
    a.router.Get("/vehicles/{id}/location", a.getLastLocationHandler)
    a.router.Get("/alerts", a.listAlertsHandler)
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
	// cancel consumer context to stop subscriber goroutines
	if a.consumerCancel != nil {
		a.consumerCancel()
	}

	// close messaging resources
	if a.publisher != nil {
		if err := a.publisher.Close(); err != nil {
			a.logger.Error("failed to close publisher", "error", err)
		}
	}
	if a.subscriber != nil {
		if err := a.subscriber.Close(); err != nil {
			a.logger.Error("failed to close subscriber", "error", err)
		}
	}

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

// HTTP Handlers for telemetry and alerts
func (a *App) ingestTelemetryHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		VehicleID string  `json:"vehicle_id"`
		Timestamp string  `json:"timestamp"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		Speed     float64 `json:"speed"`
		FuelLevel float64 `json:"fuel_level"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// validate timestamp
	if req.Timestamp == "" {
		http.Error(w, "timestamp is required", http.StatusBadRequest)
		return
	}
	if _, err := time.Parse(time.RFC3339, req.Timestamp); err != nil {
		http.Error(w, "invalid timestamp format; use RFC3339", http.StatusBadRequest)
		return
	}

	ev := &telemetry.TelemetryEvent{
		ID:        "",
		VehicleID: req.VehicleID,
		Timestamp: req.Timestamp,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
		Speed:     req.Speed,
		FuelLevel: req.FuelLevel,
	}

	if err := a.telemetryService.Ingest(r.Context(), ev); err != nil {
		a.handleServiceError(w, err)
		return
	}

	a.metrics.TelemetryEventsProcessedTotal.Inc()

	w.WriteHeader(http.StatusAccepted)
}

func (a *App) getVehicleTelemetryHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	// require both from and to and validate format
	if from == "" || to == "" {
		http.Error(w, "from and to query parameters are required (RFC3339)", http.StatusBadRequest)
		return
	}
	if _, err := time.Parse(time.RFC3339, from); err != nil {
		http.Error(w, "invalid 'from' timestamp; use RFC3339", http.StatusBadRequest)
		return
	}
	if _, err := time.Parse(time.RFC3339, to); err != nil {
		http.Error(w, "invalid 'to' timestamp; use RFC3339", http.StatusBadRequest)
		return
	}

	data, err := a.telemetryService.GetVehicleTelemetry(r.Context(), id, from, to)
	if err != nil {
		a.handleServiceError(w, err)
		return
	}

	a.metrics.TelemetryQueriesTotal.Inc()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (a *App) getLastLocationHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	data, err := a.telemetryService.GetLastLocation(r.Context(), id)
	if err != nil {
		a.handleServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (a *App) listAlertsHandler(w http.ResponseWriter, r *http.Request) {
	alerts, err := a.alertService.GetAllAlerts(r.Context())
	if err != nil {
		a.handleServiceError(w, err)
		return
	}

	a.metrics.AlertsQueriedTotal.Inc()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alerts)
}

// TelemetryRepoAdapter adapts SQLc queries to telemetry.Repository
type TelemetryRepoAdapter struct {
	queries *sqlc.Queries
}

func (a *TelemetryRepoAdapter) Create(ctx context.Context, e *telemetry.TelemetryEvent) error {
	id := pgtype.UUID{}
	if err := id.Scan(e.ID); err != nil {
		return database.WrapError(err)
	}

	vid := pgtype.UUID{}
	if err := vid.Scan(e.VehicleID); err != nil {
		return database.WrapError(err)
	}

	ts, err := time.Parse(time.RFC3339, e.Timestamp)
	if err != nil {
		return database.WrapError(err)
	}

	lat := pgtype.Numeric{}
	_ = lat.Scan(strconv.FormatFloat(e.Latitude, 'f', -1, 64))
	lon := pgtype.Numeric{}
	_ = lon.Scan(strconv.FormatFloat(e.Longitude, 'f', -1, 64))
	sp := pgtype.Numeric{}
	_ = sp.Scan(strconv.FormatFloat(e.Speed, 'f', -1, 64))
	fl := pgtype.Numeric{}
	_ = fl.Scan(strconv.FormatFloat(e.FuelLevel, 'f', -1, 64))

	params := sqlc.CreateTelemetryEventParams{
		ID:        id,
		VehicleID: vid,
		Timestamp: pgtype.Timestamptz{Time: ts, Valid: true},
		Latitude:  lat,
		Longitude: lon,
		Speed:     sp,
		FuelLevel: fl,
		CreatedAt: pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
	}

	if err := a.queries.CreateTelemetryEvent(ctx, params); err != nil {
		return database.WrapError(err)
	}

	return nil
}

func parseNumeric(n pgtype.Numeric) float64 {
	s := fmt.Sprintf("%v", n)
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

func (a *TelemetryRepoAdapter) GetByVehicleID(ctx context.Context, vehicleID string, from, to string) ([]*telemetry.TelemetryEvent, error) {
	vid := pgtype.UUID{}
	if err := vid.Scan(vehicleID); err != nil {
		return nil, database.WrapError(err)
	}

	fromT, err := time.Parse(time.RFC3339, from)
	if err != nil {
		return nil, database.WrapError(err)
	}
	toT, err := time.Parse(time.RFC3339, to)
	if err != nil {
		return nil, database.WrapError(err)
	}

	params := sqlc.GetTelemetryEventsByVehicleAndTimeRangeParams{
		VehicleID:   vid,
		Timestamp:   pgtype.Timestamptz{Time: fromT, Valid: true},
		Timestamp_2: pgtype.Timestamptz{Time: toT, Valid: true},
	}

	rows, err := a.queries.GetTelemetryEventsByVehicleAndTimeRange(ctx, params)
	if err != nil {
		return nil, database.WrapError(err)
	}

	out := make([]*telemetry.TelemetryEvent, len(rows))
	for i, r := range rows {
		out[i] = &telemetry.TelemetryEvent{
			ID:        uuidToHex(r.ID),
			VehicleID: uuidToHex(r.VehicleID),
			Timestamp: r.Timestamp.Time.Format(time.RFC3339),
			Latitude:  parseNumeric(r.Latitude),
			Longitude: parseNumeric(r.Longitude),
			Speed:     parseNumeric(r.Speed),
			FuelLevel: parseNumeric(r.FuelLevel),
		}
	}

	return out, nil
}

func (a *TelemetryRepoAdapter) GetLastByVehicleID(ctx context.Context, vehicleID string) (*telemetry.TelemetryEvent, error) {
	vid := pgtype.UUID{}
	if err := vid.Scan(vehicleID); err != nil {
		return nil, database.WrapError(err)
	}

	r, err := a.queries.GetLastTelemetryEventByVehicleID(ctx, vid)
	if err != nil {
		return nil, database.WrapError(err)
	}

	return &telemetry.TelemetryEvent{
		ID:        uuidToHex(r.ID),
		VehicleID: uuidToHex(r.VehicleID),
		Timestamp: r.Timestamp.Time.Format(time.RFC3339),
		Latitude:  parseNumeric(r.Latitude),
		Longitude: parseNumeric(r.Longitude),
		Speed:     parseNumeric(r.Speed),
		FuelLevel: parseNumeric(r.FuelLevel),
	}, nil
}

// AlertRepoAdapter adapts SQLc queries to alert.Repository
type AlertRepoAdapter struct {
	queries *sqlc.Queries
}

func (a *AlertRepoAdapter) Create(ctx context.Context, al *alert.Alert) error {
	id := pgtype.UUID{}
	if err := id.Scan(al.ID); err != nil {
		return database.WrapError(err)
	}
	vid := pgtype.UUID{}
	if err := vid.Scan(al.VehicleID); err != nil {
		return database.WrapError(err)
	}

	ts, err := time.Parse(time.RFC3339, al.Timestamp)
	if err != nil {
		return database.WrapError(err)
	}

	params := sqlc.CreateAlertParams{
		ID:          id,
		VehicleID:   vid,
		Type:        al.Type,
		Description: pgtype.Text{String: al.Description, Valid: true},
		Timestamp:   pgtype.Timestamptz{Time: ts, Valid: true},
		CreatedAt:   pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
	}

	if err := a.queries.CreateAlert(ctx, params); err != nil {
		return database.WrapError(err)
	}
	return nil
}

func (a *AlertRepoAdapter) GetByVehicleID(ctx context.Context, vehicleID string) ([]*alert.Alert, error) {
	vid := pgtype.UUID{}
	if err := vid.Scan(vehicleID); err != nil {
		return nil, database.WrapError(err)
	}

	rows, err := a.queries.GetAlertsByVehicleID(ctx, sqlc.GetAlertsByVehicleIDParams{VehicleID: vid, Limit: 100, Offset: 0})
	if err != nil {
		return nil, database.WrapError(err)
	}

	out := make([]*alert.Alert, len(rows))
	for i, r := range rows {
		out[i] = &alert.Alert{
			ID:          uuidToHex(r.ID),
			VehicleID:   uuidToHex(r.VehicleID),
			Type:        r.Type,
			Description: r.Description.String,
			Timestamp:   r.Timestamp.Time.Format(time.RFC3339),
		}
	}

	return out, nil
}

func (a *AlertRepoAdapter) GetAll(ctx context.Context) ([]*alert.Alert, error) {
	rows, err := a.queries.GetAllAlerts(ctx, sqlc.GetAllAlertsParams{Limit: 100, Offset: 0})
	if err != nil {
		return nil, database.WrapError(err)
	}

	out := make([]*alert.Alert, len(rows))
	for i, r := range rows {
		out[i] = &alert.Alert{
			ID:          uuidToHex(r.ID),
			VehicleID:   uuidToHex(r.VehicleID),
			Type:        r.Type,
			Description: r.Description.String,
			Timestamp:   r.Timestamp.Time.Format(time.RFC3339),
		}
	}

	return out, nil
}
