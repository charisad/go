package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/charisad/otelsdk"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

func main() {
	ctx := context.Background()

	cfg, err := otelsdk.LoadConfig("../otelsdk/sample.json")
	if err != nil {
		log.Fatalf("load telemetry config: %v", err)
	}

	telemetry, err := otelsdk.NewTelemetry(ctx, cfg)
	if err != nil {
		log.Fatalf("initialize telemetry: %v", err)
	}
	defer func() {
		if err := telemetry.Shutdown(ctx); err != nil {
			log.Printf("telemetry shutdown: %v", err)
		}
	}()

	tracer := otel.Tracer(cfg.ServiceName)
	meter := otel.Meter(cfg.ServiceName)

	ctx, span := tracer.Start(ctx, "simpleapp.request")
	span.SetAttributes(attribute.String("app.version", "1.0.0"))
	defer span.End()

	counter, err := meter.Int64Counter("simpleapp.requests")
	if err != nil {
		log.Fatalf("create metric: %v", err)
	}

	counter.Add(ctx, 1, metric.WithAttributes(attribute.String("route", "/checkout")))

	fmt.Println("simpleapp completed one traced request")
	time.Sleep(100 * time.Millisecond)
}
