package telemetryingestion

import (
	"context"
	"github.com/sirupsen/logrus"
	"sync"
	"turiontakehome/telemetryingestion/pkg/anomaly"
)

type telemetryValidator struct {
	validatorChannel chan TIData
	alertChannel     chan TIData
	packetChannel    chan TIData
	numberOfWorkers  int
	log              *logrus.Logger
}

func newValidator(validatorChannel chan TIData, alertChan chan TIData, packetChan chan TIData, workerNum int, logger *logrus.Logger) *telemetryValidator {
	return &telemetryValidator{validatorChannel: validatorChannel, alertChannel: alertChan, numberOfWorkers: workerNum, packetChannel: packetChan, log: logger}
}

func (t *telemetryValidator) run(ctx context.Context) {
	t.log.Infof("starting %d validator workers", t.numberOfWorkers)
	wg := &sync.WaitGroup{}
	for i := 0; i < t.numberOfWorkers; i++ {
		wg.Add(1)
		go func(workerNum int) {
			defer wg.Done()
			t.validatorWorker(ctx, workerNum)
		}(i)
	}

	<-ctx.Done()
	t.log.Infof("stopping %d validator workers", t.numberOfWorkers)
	wg.Wait()
}

func (t *telemetryValidator) validatorWorker(ctx context.Context, workerNum int) {
	t.log.Infof("starting validator worker #%v", workerNum)
	for {
		select {
		case payload := <-t.validatorChannel:
			t.log.Infof("validating payload: %v", payload.TelemetryPayload)

			//For now, given the anomaly requirements, we'll just check to see if the packet has anomalous Payload data
			//in the future, we can just expand on other validations for things like out of Normal (warnings), etc
			checkForAnomaliesAndSet(&payload)
			if payload.AnomalyFlags != 0 {
				t.alertChannel <- payload
			}
			//send to packet channel to be written to database
			t.packetChannel <- payload
		case <-ctx.Done():
			t.log.Infof("validator worker #%v cancelled", workerNum)
			return
		}
	}
}

func setAnomaly(anomaly *uint32, flag uint32) {
	*anomaly |= flag
}

func checkForAnomaliesAndSet(payload *TIData) {
	if payload.TelemetryPayload.Temperature > 35.0 {
		setAnomaly(&payload.AnomalyFlags, anomaly.TemperatureAnomalyFlag)
	}
	if payload.TelemetryPayload.Battery < 40.0 {
		setAnomaly(&payload.AnomalyFlags, anomaly.BatteryAnomalyFlag)
	}
	if payload.TelemetryPayload.Altitude < 400.0 {
		setAnomaly(&payload.AnomalyFlags, anomaly.AltitudeAnomalyFlag)
	}
	if payload.TelemetryPayload.Signal < -80.0 {
		setAnomaly(&payload.AnomalyFlags, anomaly.SignalAnomalyFlag)
	}
}
