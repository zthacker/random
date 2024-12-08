package main

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"turiontakehome/turionbackend/internal/turionbackendv1/api"
	"turiontakehome/turionbackend/internal/turionbackendv1/requesthandlers"
	"turiontakehome/turionbackend/internal/turionbackendv1/telemetry/broadcaster"
	"turiontakehome/turionbackend/internal/turionbackendv1/telemetry/telemetrymodels"
	"turiontakehome/turionbackend/internal/turionbackendv1/telemetry/telemetrystorage/persistenttelemetry"
	"turiontakehome/turionbackend/service"
	"turiontakehome/turionbackend/utils/datasource"
	"turiontakehome/turionbackend/utils/envelope"
)

func main() {
	//pprof setup
	go func() {
		logrus.Info(http.ListenAndServe("0.0.0.0:6060", nil))
	}()
	logrus.Info("pprof started")
	mainCtx, mainCancel := context.WithCancel(context.Background())

	//setup telemetry envelope -- holds a Logger and Tracer
	//later, we can configure this for if we want tracing to be enabled or not so we don't have to fatal out
	telemetryEnvelope, err := envelope.NewEnvelope(mainCtx, "turion backend service")
	if err != nil {
		logrus.Fatalf("error creating telemetryEnvelope: %v", err)
	}

	//make a fiber app
	fiberApp := fiber.New()

	//setup middleware
	fiberApp.Use(telemetryEnvelope.FiberMiddleware())
	fiberApp.Use(cors.New())

	//sig calls setup
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		logrus.Info("received shutdown signal")
		if err := fiberApp.Shutdown(); err != nil {
			telemetryEnvelope.Logger.Errorf("error shutting down Fiber: %v", err)
		}
		//call the main ctx thread's cancel, which sends the ctx.Done() to those who are using it
		mainCancel()
	}()

	//create Postgress
	turionBackendPostgres, err := datasource.NewPostgres(mainCtx)
	if err != nil {
		telemetryEnvelope.Logger.Fatalf("failed to connect to postgres: %v", err)
	}

	//create the backend service
	turionBackendService := createService(mainCtx, telemetryEnvelope, "turion backend", turionBackendPostgres)

	//register the api routes from the service with the fiber app
	telemetryEnvelope.Logger.Info("registering routes")
	api.RegisterRoutes("", "", turionBackendService, fiberApp)

	//listen and service
	telemetryEnvelope.Logger.Info("Telemetry Backend started and listening on :4000")
	logrus.Fatal(fiberApp.Listen(":4000"))

}

func createService(mainCtx context.Context, telemetryEnvelope *envelope.ServiceEnvelope, serviceName string, postgres *pgxpool.Pool) *service.TurionBackendService {
	//events channel
	eventsCh := make(chan telemetrymodels.Telemetry, 100)

	//create and run the broadcaster
	broadCaster := broadcaster.NewTelemetryBroadcaster(eventsCh, telemetryEnvelope)
	go broadCaster.Run(mainCtx)

	//create persistent telemetrystorage for Telemetry
	storage := persistenttelemetry.New(postgres, "telemetry", telemetryEnvelope, eventsCh)

	//make request handlers
	handlers := requesthandlers.MakeRequestHandlers(storage, telemetryEnvelope, broadCaster)

	//start listener
	go storage.RunPostgresListener(mainCtx)
	return service.New(telemetryEnvelope, serviceName, handlers)
}
