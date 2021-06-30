package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"google.golang.org/grpc"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"
)

var widget int64

func main() {

	ctx := context.Background()
	metricExporter, err := otlpmetricgrpc.New(
		ctx,
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithEndpoint("otel:4317"),
		otlpmetricgrpc.WithDialOption(grpc.WithBlock()),
	)
	if err != nil {
		log.Fatalf("failed to initialize metric export pipeline: %v", err)
	}

	pusher := controller.New(
		processor.New(
			simple.NewWithExactDistribution(),
			metricExporter,
		),
		controller.WithExporter(metricExporter),
	)

	err = pusher.Start(ctx)
	if err != nil {
		log.Fatalf("failed to initialize metric controller: %v", err)
	}

	// Handle this error in a sensible manner where possible
	defer func() { _ = pusher.Stop(ctx) }()

	global.SetMeterProvider(pusher.MeterProvider())

	stop := make(chan (struct{}))
	defer close(stop)
	incr := func(stop <-chan struct{}) {
		for {
			select {
			case <-stop:
				return
			case <-time.After(1 * time.Second):
				v := atomic.AddInt64(&widget, 1)
				log.Printf("widget is now %d", v)
			}
		}
	}
	meter := global.Meter("dtucker.co.uk/bugz")
	callback := func(ctx context.Context, result metric.Int64ObserverResult) {
		v := atomic.LoadInt64(&widget)
		result.Observe(v)
	}
	_ = metric.Must(meter).NewInt64ValueObserver(
		"widget",
		callback,
		metric.WithDescription("a widget"),
	)

	log.Print("Running. Press CTRL+C to stop")
	ctrlC := make(chan os.Signal, 1)
	signal.Notify(ctrlC, os.Interrupt)
	go incr(stop)
	<-ctrlC
}
