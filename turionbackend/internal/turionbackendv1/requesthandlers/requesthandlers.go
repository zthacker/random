package requesthandlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"turiontakehome/turionbackend/internal/turionbackendv1/telemetry/broadcaster"
	"turiontakehome/turionbackend/internal/turionbackendv1/telemetry/telemetrystorage"
	"turiontakehome/turionbackend/utils/envelope"
)

// TurionBackendServiceRequestHandlers implements RequestHandlers and will hold objects needed to successfully fulfill requests
type TurionBackendServiceRequestHandlers struct {
	envelope      *envelope.ServiceEnvelope
	storage       telemetrystorage.TelemetryBackendStorage
	tmBroadCaster *broadcaster.TelemetryBroadcaster
	//we can add other storages here as the service grows
}

func MakeRequestHandlers(storage telemetrystorage.TelemetryBackendStorage, telemetryEnvelope *envelope.ServiceEnvelope, tmBroadCaster *broadcaster.TelemetryBroadcaster) *TurionBackendServiceRequestHandlers {
	return &TurionBackendServiceRequestHandlers{
		envelope:      telemetryEnvelope,
		storage:       storage,
		tmBroadCaster: tmBroadCaster,
	}
}

func (t TurionBackendServiceRequestHandlers) HandleGet() fiber.Handler {
	return func(c *fiber.Ctx) error {
		t.envelope.Logger.Info("HandleGet called")

		//add child traces here if we add more logic than just calling the function
		return t.storage.GetTelemetry(c)
	}
}

func (t TurionBackendServiceRequestHandlers) HandleGetCurrent() fiber.Handler {
	return func(c *fiber.Ctx) error {
		t.envelope.Logger.Info("HandleGetCurrent called")

		//add child traces here if we add more logic than just calling the function
		return t.storage.GetCurrentTelemetry(c)
	}
}

func (t TurionBackendServiceRequestHandlers) HandleGetAnomalies() fiber.Handler {
	return func(c *fiber.Ctx) error {
		t.envelope.Logger.Info("HandleAnomalies called")

		//add child traces here if we add more logic than just calling the function
		return t.storage.GetAnomalies(c)
	}
}

func (t TurionBackendServiceRequestHandlers) HandleGetAggregations() fiber.Handler {
	return func(c *fiber.Ctx) error {
		t.envelope.Logger.Info("HandleGetAggregations called")

		//add child traces here if we add more logic than just calling the function
		return t.storage.GetAggregations(c)
	}
}

func (t TurionBackendServiceRequestHandlers) HandleWebsocket() fiber.Handler {
	return websocket.New(func(conn *websocket.Conn) {
		t.envelope.Logger.Info("HandleWebsocket called")
		t.tmBroadCaster.Register(conn)
		defer t.tmBroadCaster.Unregister(conn)

		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				t.envelope.Logger.Infof("WebSocket connection closed: %v", err)
				break
			}
		}
	})
}
