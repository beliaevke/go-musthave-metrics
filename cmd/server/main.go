package main

import (
	"fmt"
	"musthave-metrics/handlers"
	"net/http"

	"github.com/go-chi/chi"
)

func main() {
	parseFlags()
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	metric := handlers.Metric{}
	mux := chi.NewMux()
	mux.HandleFunc("/update/{metricType}/{metricName}/{metricValue}", metric.Update)
	mux.HandleFunc("/value/{metricType}/{metricName}", metric.GetValue)
	mux.HandleFunc("/", handlers.AllMetrics)
	fmt.Println("Running server on ", flagRunAddr)
	return http.ListenAndServe(flagRunAddr, mux)
}
