package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func init() {
	cfg := zap.NewDevelopmentConfig()
	cfg.EncoderConfig.TimeKey = ""
	cfg.Sampling = &zap.SamplingConfig{
		Initial:    3,
		Thereafter: 3,
		Hook: func(e zapcore.Entry, d zapcore.SamplingDecision) {
			if d == zapcore.LogDropped {
				fmt.Println("event dropped...")
			}
		},
	}
	logger, _ := cfg.Build()
	zap.ReplaceGlobals(logger)
}

func main() {
	r := mux.NewRouter()
	r.Handle("/", middleware(http.HandlerFunc(fibonacciHandler))).Queries("n", "{n}")
	log.Fatal(http.ListenAndServe(":3000", r))
}

func middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		for i := 0; i < 10; i++ {
			zap.S().Infow("requested host", "remoteHost", r.RemoteAddr, "i", i)
		}
		next.ServeHTTP(rw, r)
	})
}

func fibonacciHandler(w http.ResponseWriter, r *http.Request) {
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
