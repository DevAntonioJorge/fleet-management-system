-- name: CreateAlert :exec
INSERT INTO alerts (id, vehicle_id, type, description, timestamp, created_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetAlertsByVehicleID :many
SELECT id, vehicle_id, type, description, timestamp, created_at
FROM alerts
WHERE vehicle_id = $1
ORDER BY timestamp DESC
LIMIT $2 OFFSET $3;

-- name: GetAlertsByVehicleAndTimeRange :many
SELECT id, vehicle_id, type, description, timestamp, created_at
FROM alerts
WHERE vehicle_id = $1 AND timestamp >= $2 AND timestamp <= $3
ORDER BY timestamp DESC;

-- name: GetAllAlerts :many
SELECT id, vehicle_id, type, description, timestamp, created_at
FROM alerts
ORDER BY timestamp DESC
LIMIT $1 OFFSET $2;
