package main

import (
	"context"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"net/http"
	_ "net/http/pprof" // Enables pprof profiling via HTTP
	"os"
	"os/signal"
	"syscall"
	"turiontakehome/telemetryingestion/internal/telemetryingestion"
)

func main() {
	go func() {
		logrus.Info(http.ListenAndServe("0.0.0.0:6060", nil))
	}()
	logrus.Info("pprof started")

	//main setups to pass on
	ctx, cancel := context.WithCancel(context.Background())
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	logger.SetFormatter(&logrus.JSONFormatter{})

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	runErrCh := make(chan error)
	go func() {
		err := telemetryingestion.Run(ctx, logger)
		runErrCh <- err
	}()

	select {
	case sig := <-sigCh:
		logrus.Infof("received signal to exit: %s", sig)
		cancel()
		<-runErrCh
	case err := <-runErrCh:
		logrus.Fatalf("telemetry ingestion exited unexpectedly: %s", err)
	}

	logrus.Infof("TI Terminated")
}
