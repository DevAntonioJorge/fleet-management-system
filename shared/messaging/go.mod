module github.com/fms/fms/shared/messaging

go 1.26.2

require (
	github.com/ThreeDotsLabs/watermill v1.5.1
	github.com/ThreeDotsLabs/watermill-amqp/v2 v2.1.3
	github.com/fms/fms/shared/logger v0.0.0-00010101000000-000000000000
)

require (
	github.com/cenkalti/backoff/v3 v3.2.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/lithammer/shortuuid/v3 v3.0.7 // indirect
	github.com/oklog/ulid v1.3.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rabbitmq/amqp091-go v1.10.0 // indirect
)

replace github.com/fms/fms/shared/logger => ../logger
