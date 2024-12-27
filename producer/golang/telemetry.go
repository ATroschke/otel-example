package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

var tracer trace.Tracer
var logger *zap.Logger

func newResource(serviceName string, serviceVersion string) (*resource.Resource, error) {
	return resource.Merge(resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
		),
	)
}

func newTraceProvider(ctx context.Context, r *resource.Resource, traceEndpoint string) (*sdktrace.TracerProvider, error) {
	fmt.Printf("Setting up trace provider with endpoint %s\n", traceEndpoint)
	// Setup exporter
	tExp, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(traceEndpoint))
	if err != nil {
		return nil, err
	}

	// Return provider
	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(tExp),
		sdktrace.WithResource(r),
	), nil
}

func newLoggerProvider(ctx context.Context, r *resource.Resource, logEndpoint string) (*sdklog.LoggerProvider, error) {
	fmt.Printf("Setting up log provider with endpoint %s\n", logEndpoint)
	// Setup exporter
	lExp, err := otlploghttp.New(
		ctx,
		otlploghttp.WithInsecure(),
		otlploghttp.WithEndpoint(logEndpoint))
	if err != nil {
		return nil, err
	}
	processor := sdklog.NewBatchProcessor(lExp)
	provider := sdklog.NewLoggerProvider(
		sdklog.WithResource(r),
		sdklog.WithProcessor(processor),
	)

	return provider, nil
}

func newMeterProvider(ctx context.Context, r *resource.Resource) (*metric.MeterProvider, error) {
	metricExporter, err := stdoutmetric.New()
	if err != nil {
		return nil, err
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithResource(r),
		metric.WithReader(metric.NewPeriodicReader(metricExporter,
			metric.WithInterval(3*time.Second))),
	)

	return meterProvider, nil
}

func SetupTelemetry(ctx context.Context) {
	// Setup our service resource (Name, Version, ect)
	r, err := newResource("GO-ExampleProducer", "0.1.0")
	if err != nil {
		log.Fatalf("failed to initialize resources: %v", err)
	}
	// Create a new tracer provider with a batch span processor and the given exporter.
	tp, err := newTraceProvider(ctx, r, "localhost:4317")
	if err != nil {
		log.Fatalf("failed to initialize tracer provider: %v", err)
	}

	otel.SetTracerProvider(tp)
	tracer = tp.Tracer("al.de/andy/monitoring-stack/producer/golang")

	// Create a logger provider (We need a fully qualified URL here becauase we aren't using an ingester)
	// This isn't recommended!
	loggerProvider, err := newLoggerProvider(ctx, r, "localhost:4318")
	if err != nil {
		log.Fatalf("failed to initialize log provider: %v", err)
	}

	logger = zap.New(otelzap.NewCore("al.de/andy/monitoring-stack/producer/golang", otelzap.WithLoggerProvider(loggerProvider)))

	// Create a metric provider
	meterProvider, err := newMeterProvider(ctx, r)
	if err != nil {
		log.Fatalf("failed to initialize meter provider: %v", err)
	}

	otel.SetMeterProvider(meterProvider)
}
