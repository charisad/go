package otelsdk

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
)

type Telemetry struct {
	TracerProvider *sdktrace.TracerProvider
	MeterProvider  *metric.MeterProvider
}

func NewTelemetry(ctx context.Context, cfg *Config) (*Telemetry, error) {
	if cfg == nil {
		return nil, errors.New("config is required")
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	res, err := buildResource(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("build resource: %w", err)
	}

	spanProcessors, err := buildSpanProcessors(ctx, cfg)
	if err != nil {
		return nil, err
	}

	traceOptions := []sdktrace.TracerProviderOption{sdktrace.WithResource(res)}
	for _, sp := range spanProcessors {
		traceOptions = append(traceOptions, sdktrace.WithSpanProcessor(sp))
	}

	tracerProvider := sdktrace.NewTracerProvider(traceOptions...)

	otel.SetTracerProvider(tracerProvider)

	var meterProvider *metric.MeterProvider
	if cfg.ShowConsoleMetrics || cfg.Protocol != "" {
		metricReaders, err := buildMetricReaders(ctx, cfg)
		if err != nil {
			return nil, err
		}
		if len(metricReaders) > 0 {
			providerOptions := make([]metric.Option, 0, len(metricReaders)+1)
			providerOptions = append(providerOptions, metric.WithResource(res))
			for _, reader := range metricReaders {
				providerOptions = append(providerOptions, metric.WithReader(reader))
			}
			meterProvider = metric.NewMeterProvider(providerOptions...)
			otel.SetMeterProvider(meterProvider)
		}
	}

	return &Telemetry{
		TracerProvider: tracerProvider,
		MeterProvider:  meterProvider,
	}, nil
}

func buildResource(ctx context.Context, cfg *Config) (*resource.Resource, error) {
	attrs := []attribute.KeyValue{
		semconv.ServiceNameKey.String(cfg.ServiceNameResourceAttribute()),
	}
	for key, value := range cfg.ResourceAttributes() {
		attrs = append(attrs, attribute.Key(key).String(value))
	}

	return resource.New(ctx,
		resource.WithAttributes(attrs...),
		resource.WithSchemaURL(semconv.SchemaURL),
	)
}

func buildSpanProcessors(ctx context.Context, cfg *Config) ([]sdktrace.SpanProcessor, error) {
	var processors []sdktrace.SpanProcessor

	traceExporter, err := buildTraceExporter(ctx, cfg)
	if err != nil {
		return nil, err
	}
	processors = append(processors, sdktrace.NewBatchSpanProcessor(traceExporter))

	if cfg.ShowConsoleTrace {
		consoleExporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
		if err != nil {
			return nil, fmt.Errorf("build console trace exporter: %w", err)
		}
		processors = append(processors, sdktrace.NewSimpleSpanProcessor(consoleExporter))
	}

	return processors, nil
}

func buildTraceExporter(ctx context.Context, cfg *Config) (sdktrace.SpanExporter, error) {
	endpoint := cfg.Endpoint()
	if cfg.Protocol == "grpc" {
		exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithEndpoint(endpoint), otlptracegrpc.WithInsecure())
		if err != nil {
			return nil, fmt.Errorf("create grpc trace exporter: %w", err)
		}
		return exporter, nil
	}

	exporter, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpoint(endpoint), otlptracehttp.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("create http trace exporter: %w", err)
	}
	return exporter, nil
}

func buildMetricReaders(ctx context.Context, cfg *Config) ([]metric.Reader, error) {
	var readers []metric.Reader

	if cfg.ShowConsoleMetrics {
		console, err := stdoutmetric.New(stdoutmetric.WithPrettyPrint())
		if err != nil {
			return nil, fmt.Errorf("build console metrics exporter: %w", err)
		}
		readers = append(readers, metric.NewPeriodicReader(console, metric.WithInterval(15*time.Second)))
	}

	metricExporter, err := buildMetricExporter(ctx, cfg)
	if err != nil {
		return nil, err
	}
	readers = append(readers, metric.NewPeriodicReader(metricExporter, metric.WithInterval(15*time.Second)))

	return readers, nil
}

func buildMetricExporter(ctx context.Context, cfg *Config) (metric.Exporter, error) {
	endpoint := cfg.Endpoint()
	if cfg.Protocol == "grpc" {
		exporter, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithEndpoint(endpoint), otlpmetricgrpc.WithInsecure())
		if err != nil {
			return nil, fmt.Errorf("create grpc metric exporter: %w", err)
		}
		return exporter, nil
	}

	exporter, err := otlpmetrichttp.New(ctx, otlpmetrichttp.WithEndpoint(endpoint), otlpmetrichttp.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("create http metric exporter: %w", err)
	}
	return exporter, nil
}

func (t *Telemetry) Shutdown(ctx context.Context) error {
	var firstErr error
	if t.MeterProvider != nil {
		if err := t.MeterProvider.Shutdown(ctx); err != nil {
			firstErr = fmt.Errorf("meter shutdown: %w", err)
		}
	}
	if t.TracerProvider != nil {
		if err := t.TracerProvider.Shutdown(ctx); err != nil {
			if firstErr != nil {
				return fmt.Errorf("%v; tracer shutdown: %w", firstErr, err)
			}
			return fmt.Errorf("tracer shutdown: %w", err)
		}
	}
	return firstErr
}
