package api

import (
	"github.com/gofiber/fiber/v2"
)

type TelemetryRequestHandlers interface {
	//have telemetry otel be apart of these
	HandleGet() fiber.Handler
	HandleGetCurrent() fiber.Handler
	HandleGetAnomalies() fiber.Handler
	HandleGetAggregations() fiber.Handler
	HandleWebsocket() fiber.Handler
}

// looks like /api/v1/telemetry/...
func addTelemetryRoutes(handlers RequestHandlers, router fiber.Router) {
	router.Get("/telemetry", handlers.HandleGet())
	router.Get("/telemetry/current", handlers.HandleGetCurrent())
	router.Get("/telemetry/anomalies", handlers.HandleGetAnomalies())
	router.Get("/telemetry/aggregations", handlers.HandleGetAggregations())
	router.Get("/telemetry/ws", handlers.HandleWebsocket())
}
