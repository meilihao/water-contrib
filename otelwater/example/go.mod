module github.com/meilihao/water-contrib/otelwater/example

go 1.16

replace github.com/meilihao/water-contrib/otelwater => ../

require (
	github.com/meilihao/water v0.0.0-20210419113811-23fde735115e
	github.com/meilihao/water-contrib/otelwater v0.0.0-00010101000000-000000000000
	go.opentelemetry.io/otel v0.19.0
	go.opentelemetry.io/otel/exporters/stdout v0.19.0
	go.opentelemetry.io/otel/sdk v0.19.0
	go.opentelemetry.io/otel/trace v0.19.0
)
