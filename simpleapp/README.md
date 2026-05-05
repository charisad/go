# simpleapp

A minimal Go application that uses the shared `otelsdk` package for telemetry.

## Run

```bash
cd simpleapp
go run .
```

## What it does

- Loads telemetry settings from `../otelsdk/sample.json`
- Initializes trace and metric exporters via `otelsdk.NewTelemetry`
- Starts a sample span and emits a simple counter metric
- Shuts down telemetry cleanly at program exit
