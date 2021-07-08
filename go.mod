module github.com/dave-tucker/otel-metric-repro

go 1.16

require (
	go.opentelemetry.io/otel v1.0.0-RC1
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric v0.21.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v0.21.0
	go.opentelemetry.io/otel/metric v0.21.0
	go.opentelemetry.io/otel/sdk/metric v0.21.0
	google.golang.org/grpc v1.39.0 // indirect
)
