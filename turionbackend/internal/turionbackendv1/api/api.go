package api

import "github.com/gofiber/fiber/v2"

type OffsetLimitParams struct {
	Offset int
	Limit  int
}

type RequestHandlers interface {
	TelemetryRequestHandlers
	//add other handlers here as the backend grows to handle other requests...
}

func RegisterRoutes(basePath string, serviceName string, handlers RequestHandlers, fiberApp *fiber.App) {

	/*
		Consider including more meaningful groupings if your API expands in the future
		(e.g., /api/v1/telemetry could later become /api/v1/telemetry/sensors, /api/v1/telemetry/analytics, etc.).
	*/

	//groupings for compatability
	v1 := fiberApp.Group("api/v1")

	addTelemetryRoutes(handlers, v1)
}
