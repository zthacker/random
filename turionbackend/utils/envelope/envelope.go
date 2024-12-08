package envelope

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"

	"go.opentelemetry.io/otel/sdk/trace"
	oteltracer "go.opentelemetry.io/otel/trace"
	"os"
)

// ServiceEnvelope holds a Logger and Tracer. We can dangle methods onto this that can
// be helper functions for tracing and logging, as well as add anything else to this Envelope that's
// being passed around to others
type ServiceEnvelope struct {
	Logger *logrus.Logger
	Tracer oteltracer.Tracer
}

func NewEnvelope(ctx context.Context, name string) (*ServiceEnvelope, error) {
	tracerProvider, err := initTracing(ctx, name)
	if err != nil {
		return nil, err
	}

	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	logger.SetFormatter(&logrus.JSONFormatter{})

	return &ServiceEnvelope{
		Tracer: tracerProvider.Tracer(name),
		Logger: logger,
	}, nil

}

// FiberMiddleware setups up a fiber.Handler for requests
func (se *ServiceEnvelope) FiberMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()

		rootCtx, rootSpan := se.Tracer.Start(ctx, c.Method()+" "+c.Path())
		defer rootSpan.End()

		c.SetUserContext(rootCtx)

		se.LogWithContext(rootCtx, "root span started for request")

		return c.Next()
	}
}

// LogWithContext logs with the trace_id and span_id, along with a message provided
func (se *ServiceEnvelope) LogWithContext(ctx context.Context, message string) {
	span := oteltracer.SpanFromContext(ctx)
	se.Logger.WithFields(logrus.Fields{
		"trace_id": span.SpanContext().TraceID().String(),
		"span_id":  span.SpanContext().SpanID().String(),
	}).Info(message)
}

func initTracing(ctx context.Context, serviceName string) (*trace.TracerProvider, error) {
	// Configure OTLP Exporter -- using Tempo here
	exporter, err := otlptrace.New(
		ctx,
		otlptracehttp.NewClient(
			otlptracehttp.WithEndpoint("tempo:14268"),
			otlptracehttp.WithInsecure(),
		),
	)
	if err != nil {
		return nil, err
	}

	// Create Tracer Provider
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewSchemaless(
			semconv.ServiceNameKey.String(serviceName),
		)),
	)

	//sets the global provider
	otel.SetTracerProvider(tp)
	return tp, nil
}
