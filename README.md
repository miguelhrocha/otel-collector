# OTEL collector

OTEL collector for receiving, processing and exporting telemetry data.

## How to use

To run the OTEL collector for local validation, use the following command:

```bash
make run-example
```

This will start the OTEL collector using the port `4317` for receiving telemetry data in OTLP format.

It disables all system-level exports and uses a simple logging exporter to print the received telemetry data to the console.

To run normally, you can use the following command:

```bash
make run
```

### Configuration

The OTEL collector service can be configured via environment variables. For documentation on available configuration, 
please refer to the [config.go file](./config/config.go) in the source code.

## Sending data to the collector

You can use the following make command to send sample logs to the OTEL collector:

```bash
make run-client
```

This command sends example logs every second to the OTEL collector running on `localhost:4317`.
