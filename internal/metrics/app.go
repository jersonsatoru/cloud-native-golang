package metrics

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"runtime"
	"strconv"

	"github.com/gorilla/mux"

	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/sdk/export/metric/aggregation"
	"go.opentelemetry.io/otel/sdk/metric/aggregator/histogram"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	selector "go.opentelemetry.io/otel/sdk/metric/selector/simple"
)

const (
	serviceName = "fibonacci"
)

var requests metric.Int64Counter

func main() {
	config := prometheus.Config{
		DefaultHistogramBoundaries: []float64{1, 2, 5, 10, 20, 50},
	}
	c := controller.New(
		processor.NewFactory(
			selector.NewWithHistogramDistribution(
				histogram.WithExplicitBoundaries(config.DefaultHistogramBoundaries),
			),
			aggregation.CumulativeTemporalitySelector(),
			processor.WithMemory(true),
		),
	)
	exporter, err := prometheus.New(config, c)
	if err != nil {
		log.Panicf("failed to initalize prometheus exporter %v", err)
	}
	global.SetMeterProvider(exporter.MeterProvider())
	meter := global.Meter(serviceName)
	requests = metric.Must(meter).NewInt64Counter(
		"fibonacci_requests_counter",
		metric.WithDescription("Total number os Fibonacci requests."))

	m := runtime.MemStats{}
	metric.Must(meter).NewInt64CounterObserver(
		"memory_usage_bytes",
		func(c context.Context, ior metric.Int64ObserverResult) {
			runtime.ReadMemStats(&m)
			ior.Observe(int64(m.Sys))
		},
		metric.WithDescription("Amount of memory used"),
	)

	metric.Must(meter).NewInt64CounterObserver(
		"num_goroutines",
		func(c context.Context, ior metric.Int64ObserverResult) {
			ior.Observe(int64(runtime.NumGoroutine()))
		},
		metric.WithDescription("Number of running goroutines."),
	)

	r := mux.NewRouter()
	r.HandleFunc("/metrics", exporter.ServeHTTP)
	r.Handle("/", http.HandlerFunc(fibonacciHandler)).Queries("n", "{n}")
	log.Fatal(http.ListenAndServe(":3000", r))
}

func fibonacciHandler(w http.ResponseWriter, r *http.Request) {
	requests.Add(r.Context(), 1)
	vars := mux.Vars(r)
	n, _ := strconv.Atoi(vars["n"])
	ch := fibonacci(r.Context(), n)
	result := <-ch
	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"value": result,
	})
}

func fibonacci(ctx context.Context, n int) chan int {
	ch := make(chan int)
	go func() {
		result := 1
		if n > 1 {
			a := fibonacci(ctx, n-1)
			b := fibonacci(ctx, n-2)
			result = <-a + <-b
		}
		ch <- result
	}()
	return ch
}
