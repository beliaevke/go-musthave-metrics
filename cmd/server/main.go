package main

import (
	"context"
	"log"
	"musthave-metrics/cmd/server/config"
	"musthave-metrics/handlers"
	"musthave-metrics/internal/compress"
	"musthave-metrics/internal/logger"
	"musthave-metrics/internal/postgres"
	"musthave-metrics/internal/service"
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
	if cfg.FlagHashKey != "" {
		hd := service.NewHashData(cfg.FlagHashKey)
		mux.Use(hd.WithHashVerification)
	}
	mux.Handle("/update/{metricType}/{metricName}/{metricValue}", handlers.UpdateHandler())
	mux.Handle("/update/", updateHandler(cfg))
	mux.Handle("/updates/", handlers.UpdateBatchDBHandler(cfg.FlagDatabaseDSN))
	mux.Handle("/value/{metricType}/{metricName}", handlers.GetValueHandler())
	mux.Handle("/value/", valueHandler(cfg))
	mux.Handle("/ping", handlers.PingDBHandler(cfg.FlagDatabaseDSN))
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

func updateHandler(cfg config.ServerFlags) http.Handler {
	if cfg.FlagDatabaseDSN != "" {
		ctx := context.Background()
		postgres.SetDB(ctx, cfg.FlagDatabaseDSN)
		return handlers.UpdateDBHandler(ctx, cfg.FlagDatabaseDSN, cfg.FlagHashKey)
	}
	return handlers.UpdateJSONHandler(cfg.FlagStoreInterval, cfg.FlagFileStoragePath)
}

func valueHandler(cfg config.ServerFlags) http.Handler {
	if cfg.FlagDatabaseDSN != "" {
		ctx := context.Background()
		return handlers.GetValueDBHandler(ctx, cfg.FlagDatabaseDSN, cfg.FlagHashKey)
	}
	return handlers.GetValueJSONHandler()
}
