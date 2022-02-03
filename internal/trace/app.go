package trace

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	serviceName = "fibonacci_2"
	environment = "development"
	id          = 1
)

func main() {
	tp, err := traceProvider("http://localhost:14268/api/traces")
	if err != nil {
		log.Fatal(err)
	}
	otel.SetTracerProvider(tp)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	defer func(ctx context.Context) {
		ctx, cancel := context.WithTimeout(ctx, time.Second*5)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	}(ctx)

	r := mux.NewRouter()
	r.Handle("/", otelhttp.NewHandler(http.HandlerFunc(fibonacciHandler), "root")).Queries("n", "{n}")
	log.Fatal(http.ListenAndServe(":8009", r))
}

func traceProvider(url string) (*tracesdk.TracerProvider, error) {
	exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
	if err != nil {
		return nil, err
	}

	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exporter),
		tracesdk.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(serviceName),
				attribute.String("environment", environment),
				attribute.Int64("id", id),
			),
		),
	)
	return tp, nil
}

func fibonacciHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	n, _ := strconv.Atoi(vars["n"])
	ctx := r.Context()
	result := <-fibonacci(ctx, n)
	sp := trace.SpanFromContext(ctx)
	defer sp.End()
	if sp != nil {
		sp.SetAttributes(
			attribute.Int("parameter", n),
			attribute.Int("result", result),
		)
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"value": result,
	})
}

func fibonacci(ctx context.Context, n int) chan int {
	ch := make(chan int)
	go func() {
		tr := otel.GetTracerProvider().Tracer(serviceName)
		cctx, sp := tr.Start(
			ctx,
			fmt.Sprintf("Fibonacci(%d)", n))
		defer sp.End()
		result := 1
		if n > 1 {
			a := fibonacci(cctx, n-1)
			b := fibonacci(cctx, n-2)
			result = <-a + <-b
		}
		ch <- result
	}()
	return ch
}
