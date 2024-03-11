package main

import (
	"musthave-metrics/handlers"
	"net/http"

	"github.com/go-chi/chi"
)

func main() {
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
	return http.ListenAndServe(`:8080`, mux)
}
