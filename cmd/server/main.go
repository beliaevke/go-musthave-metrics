package main

import (
	"log"
	"musthave-metrics/cmd/server/config"
	"musthave-metrics/handlers"
	"musthave-metrics/internal/compress"
	"musthave-metrics/internal/logger"
	"net/http"
	"time"

	"github.com/go-chi/chi"
)

func main() {
	cfg := config.ParseFlags()
	if cfg.FlagRestore {
		handlers.RestoreMetrics(cfg.FlagFileStoragePath)
	}
	if cfg.FlagStoreInterval != 0 {
		storeMetrics(cfg)
	}
	if err := run(cfg); err != nil {
		log.Fatal(err)
	}
}

func run(cfg config.ServerFlags) error {
	logger.ServerRunningInfo(cfg.FlagRunAddr)
	mux := chi.NewMux()
	mux.Use(logger.WithLogging, compress.WithGzipEncoding)
	mux.Handle("/update/{metricType}/{metricName}/{metricValue}", handlers.UpdateHandler())
	mux.Handle("/update/", handlers.UpdateJSONHandler(cfg.FlagStoreInterval, cfg.FlagFileStoragePath))
	mux.Handle("/value/{metricType}/{metricName}", handlers.GetValueHandler())
	mux.Handle("/value/", handlers.GetValueJSONHandler())
	mux.Handle("/", handlers.AllMetricsHandler())
	return http.ListenAndServe(cfg.FlagRunAddr, mux)
}

func storeMetrics(cfg config.ServerFlags) {
	f := func() {
		handlers.StoreMetrics(cfg.FlagFileStoragePath)
		storeMetrics(cfg)
	}
	time.AfterFunc(time.Duration(cfg.FlagStoreInterval)*time.Second, f)
}
