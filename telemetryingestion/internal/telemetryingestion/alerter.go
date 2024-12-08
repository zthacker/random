package telemetryingestion

import (
	"context"
	"github.com/sirupsen/logrus"
)

type telemetryAlerter struct {
	alerterChannel chan TIData
	log            *logrus.Logger
}

func newAlerter(alerterChannel chan TIData, logger *logrus.Logger) *telemetryAlerter {
	return &telemetryAlerter{alerterChannel, logger}
}

func (a *telemetryAlerter) run(ctx context.Context) {
	logrus.Info("starting alerter")

	//in the future, this is the place where  we'll send alerts to some sort of paging or alerting system. We can do
	//this in many forms that shouldn't be established in a vacuum; so for this test, we'll just simply log
	//and know we'll come back to this later
	for {
		select {
		case alert := <-a.alerterChannel:
			a.log.Infof("alert received: %v", alert)
		case <-ctx.Done():
			a.log.Info("alerter canceled")
			return
		}
	}
}
