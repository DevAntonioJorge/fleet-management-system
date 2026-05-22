module github.com/fms/fms/services/telemetry

go 1.26.2

require (
	github.com/fms/fms/shared/messaging v0.0.0
	github.com/google/uuid v1.6.0
)

replace github.com/fms/fms/shared/messaging => ../../shared/messaging
