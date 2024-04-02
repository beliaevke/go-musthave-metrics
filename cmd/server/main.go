package main

import (
	"log"
	"musthave-metrics/handlers"
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
	mux.Handle("/update/{metricType}/{metricName}/{metricValue}", logger.WithLogging(handlers.UpdateHandler()))
	mux.Handle("/update/", logger.WithLogging(handlers.UpdateJSONHandler()))
	mux.Handle("/value/{metricType}/{metricName}", logger.WithLogging(handlers.GetValueHandler()))
	mux.Handle("/value/", logger.WithLogging(handlers.GetValueJSONHandler()))
	mux.Handle("/", logger.WithLogging(handlers.AllMetricsHandler()))
	return http.ListenAndServe(flagRunAddr, mux)
}
