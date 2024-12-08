package service

import (
	"turiontakehome/turionbackend/internal/turionbackendv1/api"
	"turiontakehome/turionbackend/utils/envelope"
)

type TurionBackendService struct {
	Name     string
	Envelope *envelope.ServiceEnvelope
	//TODO add in a config
	//TODO Add in metrics even
	api.RequestHandlers
}

func New(telemetryEnvelope *envelope.ServiceEnvelope, name string, requestHandlers api.RequestHandlers) *TurionBackendService {
	return &TurionBackendService{Envelope: telemetryEnvelope, Name: name, RequestHandlers: requestHandlers}
}
