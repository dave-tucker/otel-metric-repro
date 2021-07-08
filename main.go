package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"
)

var widget int64

func main() {
	ctx := context.Background()
	client := otlpmetricgrpc.NewClient(
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithEndpoint("otel:4317"),
	)
	exp, err := otlpmetric.New(ctx, client)
	if err != nil {
		log.Fatalf("Failed to create the collector exporter: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		if err := exp.Shutdown(ctx); err != nil {
			otel.Handle(err)
		}
	}()

	pusher := controller.New(
		processor.New(
			simple.NewWithExactDistribution(),
			exp,
		),
		controller.WithExporter(exp),
		controller.WithCollectPeriod(2*time.Second),
	)
	global.SetMeterProvider(pusher.MeterProvider())

	if err := pusher.Start(ctx); err != nil {
		log.Fatalf("could not start metric controoler: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		// pushes any last exports to the receiver
		if err := pusher.Stop(ctx); err != nil {
			otel.Handle(err)
		}
	}()

	meter := global.Meter("test-meter")

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
