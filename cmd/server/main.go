package main

import (
	"log"
	"musthave-metrics/handlers"
	"musthave-metrics/internal/compress"
	"musthave-metrics/internal/logger"
	"net/http"

	"github.com/go-chi/chi"
)

func main() {
	parseFlags()
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	logger.ServerRunningInfo(flagRunAddr)
	mux := chi.NewMux()
	mux.Use(logger.WithLogging, compress.WithGzipEncoding)
	mux.Handle("/update/{metricType}/{metricName}/{metricValue}", handlers.UpdateHandler())
	mux.Handle("/update/", handlers.UpdateJSONHandler())
	mux.Handle("/value/{metricType}/{metricName}", handlers.GetValueHandler())
	mux.Handle("/value/", handlers.GetValueJSONHandler())
	mux.Handle("/", handlers.AllMetricsHandler())
	return http.ListenAndServe(flagRunAddr, mux)
}
