package telemetrystorage

import (
	"context"
	"github.com/gofiber/fiber/v2"
)

type TelemetryBackendStorage interface {
	telemetryStorage
}

type telemetryStorage interface {
	GetTelemetry(c *fiber.Ctx) error
	GetCurrentTelemetry(c *fiber.Ctx) error
	GetAnomalies(c *fiber.Ctx) error
	GetAggregations(c *fiber.Ctx) error
	RunPostgresListener(ctx context.Context)
}
