-- name: CreateTelemetryEvent :exec
INSERT INTO telemetry_events (id, vehicle_id, timestamp, latitude, longitude, speed, fuel_level, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: GetTelemetryEventsByVehicleID :many
SELECT id, vehicle_id, timestamp, latitude, longitude, speed, fuel_level, created_at
FROM telemetry_events
WHERE vehicle_id = $1
ORDER BY timestamp DESC
LIMIT $2 OFFSET $3;

-- name: GetTelemetryEventsByVehicleAndTimeRange :many
SELECT id, vehicle_id, timestamp, latitude, longitude, speed, fuel_level, created_at
FROM telemetry_events
WHERE vehicle_id = $1 AND timestamp >= $2 AND timestamp <= $3
ORDER BY timestamp DESC;

-- name: GetLastTelemetryEventByVehicleID :one
SELECT id, vehicle_id, timestamp, latitude, longitude, speed, fuel_level, created_at
FROM telemetry_events
WHERE vehicle_id = $1
ORDER BY timestamp DESC
LIMIT 1;
