# otelsdk

`otelsdk` is a Go telemetry bootstrap SDK for your microservices. It creates an OpenTelemetry trace and metric pipeline, exports to a remote OTLP collector, and optionally prints traces and metrics to the console.

## Features

- Reads a JSON configuration file from the `otelsdk` folder
- Sends traces and metrics to an OTLP collector
- Supports both `http` and `grpc` transport protocols
- Adds service resource attributes and user-defined extra attributes
- Optional console output for trace and metric debugging

## Configuration

Create a JSON configuration file like `otelsdk/sample.json`:

```json
{
  "ServiceName": "charisad.checkout.api",
  "Host": "otel.internal.charisad.com",
  "Port": 4317,
  "Protocol": "http",
  "ShowConsoleMetrics": false,
  "ShowConsoleTrace": true,
  "ExtraResourceAttributes": {
    "product.category": "data.services",
    "product.team": "platform.product"
  }
}
```

### Field reference

- `ServiceName`: required service identifier
- `Host`: OTLP collector host
- `Port`: OTLP collector port
- `Protocol`: `http` or `grpc`
- `ShowConsoleMetrics`: optional console meter output
- `ShowConsoleTrace`: optional console trace output
- `ExtraResourceAttributes`: extra resource attributes attached to every span and metric

## Usage

Import the package and initialize telemetry in your microservice `main.go`:

```go
package main

import (
  "context"
  "fmt"
  "log"

  "github.com/charisad/otelsdk"
  "go.opentelemetry.io/otel"
  "go.opentelemetry.io/otel/metric"
)

func main() {
  ctx := context.Background()

  cfg, err := otelsdk.LoadConfig("otelsdk/sample.json")
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

  tracer := otel.Tracer("charisad.checkout.api")
  meter := otel.Meter("charisad.checkout.api")

  ctx, span := tracer.Start(ctx, "checkout.process")
  defer span.End()

  counter := metric.Must(meter).NewInt64Counter("checkout.orders.processed")
  counter.Add(ctx, 1)

  fmt.Println("checkout flow complete")
}
```

## Extending for your microservices

- Use `otelsdk.LoadConfig` in every service to load the shared telemetry settings
- Create a service-specific `Tracer` and `Meter` using `otel.Tracer(...)` and `otel.Meter(...)`
- Wrap important operations in spans and record metrics using the OTel API
- Keep `otelsdk/sample.json` as a platform SDK config example
