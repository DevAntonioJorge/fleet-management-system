-- name: CreateVehicle :exec
INSERT INTO vehicles (id, plate, model, created_at) 
VALUES ($1, $2, $3, $4);

-- name: GetVehicleByID :one
SELECT id, plate, model, created_at 
FROM vehicles 
WHERE id = $1;

-- name: GetAllVehicles :many
SELECT id, plate, model, created_at 
FROM vehicles 
ORDER BY created_at DESC;

-- name: GetVehicleByPlate :one
SELECT id, plate, model, created_at 
FROM vehicles 
WHERE plate = $1;
