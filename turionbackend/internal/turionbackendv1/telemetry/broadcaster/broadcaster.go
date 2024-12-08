package broadcaster

import (
	"context"
	"encoding/json"
	"github.com/gofiber/websocket/v2"
	"sync"
	"turiontakehome/turionbackend/internal/turionbackendv1/telemetry/telemetrymodels"
	"turiontakehome/turionbackend/utils/envelope"
)

type TelemetryBroadcaster struct {
	clients  map[*websocket.Conn]bool
	mu       sync.Mutex
	envelope *envelope.ServiceEnvelope
	events   chan telemetrymodels.Telemetry
}

func NewTelemetryBroadcaster(eventsChan chan telemetrymodels.Telemetry, telemEnvelope *envelope.ServiceEnvelope) *TelemetryBroadcaster {
	return &TelemetryBroadcaster{clients: make(map[*websocket.Conn]bool), envelope: telemEnvelope, events: eventsChan}
}

func (b *TelemetryBroadcaster) Run(ctx context.Context) {
	b.envelope.Logger.Info("starting telemetry broadcaster")
	for {
		select {
		case <-ctx.Done():
			b.envelope.Logger.Info("Stopping Telemetry Broadcaster and closing all connections")
			b.mu.Lock()
			for conn := range b.clients {
				conn.Close()
				delete(b.clients, conn)
			}
			b.mu.Unlock()
			return
		case message := <-b.events:
			b.mu.Lock()
			msg := message
			for client := range b.clients {
				b.envelope.Logger.Info("sending telemetry to client")
				byteMsg, err := json.Marshal(msg)
				if err != nil {
					b.envelope.Logger.Error("Error marshalling telemetry to bytes: %s", err.Error())
					continue
				}
				err = client.WriteMessage(websocket.TextMessage, byteMsg)
				if err != nil {
					client.Close()
					delete(b.clients, client)
					b.envelope.Logger.Info("Error sending message, client disconnected")
				}
			}
			b.mu.Unlock()
		}
	}
}
func (b *TelemetryBroadcaster) Register(client *websocket.Conn) {
	b.envelope.Logger.Info("Registering client")
	b.mu.Lock()
	b.clients[client] = true
	b.mu.Unlock()
	b.envelope.Logger.Info("Client registered")
}
func (b *TelemetryBroadcaster) Unregister(client *websocket.Conn) {
	b.envelope.Logger.Info("Unregistering client")
	b.mu.Lock()
	if _, ok := b.clients[client]; ok {
		delete(b.clients, client)
		client.Close()
		b.envelope.Logger.Info("Client unregistered")
	}
	b.mu.Unlock()
}
