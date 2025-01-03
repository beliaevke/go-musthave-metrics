package main

import (
	"context"
	"expvar"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	rpprof "runtime/pprof"
	"syscall"
	"time"

	"musthave-metrics/cmd/server/config"
	"musthave-metrics/handlers"
	"musthave-metrics/internal/compress"
	"musthave-metrics/internal/logger"
	"musthave-metrics/internal/postgres"
	"musthave-metrics/internal/service"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

var (
	buildVersion, buildDate, buildCommit string = "N/A", "N/A", "N/A"
)

func main() {
	// через этот канал сообщим основному потоку, что соединения закрыты
	idleConnsClosed := make(chan struct{})
	// канал для перенаправления прерываний
	// поскольку нужно отловить всего одно прерывание,
	// ёмкости 1 для канала будет достаточно
	sigs := make(chan os.Signal, 1)
	// регистрируем перенаправление прерываний
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	logger.BuildInfo(buildVersion, buildDate, buildCommit)
	cfg := config.ParseFlags()
	if cfg.FlagRestore {
		handlers.RestoreMetrics(cfg.FlagFileStoragePath)
	}
	if cfg.FlagStoreInterval != 0 {
		storeMetrics(cfg)
	}

	HTTPServer := run(cfg)

	// запускаем горутину обработки пойманных прерываний
	go func() {
		// читаем из канала прерываний
		// поскольку нужно прочитать только одно прерывание,
		// можно обойтись без цикла
		<-sigs
		// получили сигнал os.Interrupt, запускаем процедуру graceful shutdown
		if err := HTTPServer.Shutdown(context.Background()); err != nil {
			// ошибки закрытия Listener
			logger.Warnf("HTTP server Shutdown: " + err.Error())
		}
		// сообщаем основному потоку,
		// что все сетевые соединения обработаны и закрыты
		close(idleConnsClosed)
	}()

	if err := HTTPServer.ListenAndServe(); err != http.ErrServerClosed {
		// ошибки старта или остановки Listener
		logger.Warnf("HTTP server ListenAndServe: " + err.Error())
	}

	// ждём завершения процедуры graceful shutdown
	<-idleConnsClosed
	// получили оповещение о завершении
	// здесь можно освобождать ресурсы перед выходом,
	// например закрыть соединение с базой данных,
	// закрыть открытые файлы
	logger.Infof("Server Shutdown gracefully")

	fmem, err := os.Create(cfg.FlagMemProfile)
	if err != nil {
		panic(err)
	}
	defer fmem.Close()
	runtime.GC() // получаем статистику по использованию памяти
	if err := rpprof.WriteHeapProfile(fmem); err != nil {
		panic(err)
	}
}

func run(cfg config.ServerFlags) *http.Server {
	logger.ServerRunningInfo(cfg.FlagRunAddr)
	mux := chi.NewMux()
	if cfg.FlagHashKey != "" {
		hd := service.NewHashData(cfg.FlagHashKey)
		mux.Use(hd.WithHashVerification)
	}
	if cfg.FlagCryptoKey != "" {
		kd := service.NewKeyData(cfg.FlagCryptoKey)
		mux.Use(kd.WithEncrypt)
	}
	mux.Use(logger.WithLogging, compress.WithGzipEncoding)
	mux.Handle("/update/{metricType}/{metricName}/{metricValue}", handlers.UpdateHandler())
	mux.Handle("/update/", updateHandler(cfg))
	mux.Handle("/updates/", handlers.UpdateBatchDBHandler(cfg.FlagDatabaseDSN))
	mux.Handle("/value/{metricType}/{metricName}", handlers.GetValueHandler())
	mux.Handle("/value/", valueHandler(cfg))
	mux.Handle("/ping", handlers.PingDBHandler(cfg.FlagDatabaseDSN))
	mux.Handle("/", handlers.AllMetricsHandler())
	mux.Mount("/debug", middleware.Profiler())

	HTTPServer := &http.Server{
		Addr:    cfg.FlagRunAddr,
		Handler: mux,
	}

	return HTTPServer
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

func Profiler() http.Handler {
	r := chi.NewRouter()
	//r.Use(NoCache)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, r.RequestURI+"/pprof/", http.StatusMovedPermanently)
	})
	r.HandleFunc("/pprof", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, r.RequestURI+"/", http.StatusMovedPermanently)
	})

	r.HandleFunc("/pprof/*", pprof.Index)
	r.HandleFunc("/pprof/cmdline", pprof.Cmdline)
	r.HandleFunc("/pprof/profile", pprof.Profile)
	r.HandleFunc("/pprof/symbol", pprof.Symbol)
	r.HandleFunc("/pprof/trace", pprof.Trace)
	r.Handle("/vars", expvar.Handler())

	r.Handle("/pprof/goroutine", pprof.Handler("goroutine"))
	r.Handle("/pprof/threadcreate", pprof.Handler("threadcreate"))
	r.Handle("/pprof/mutex", pprof.Handler("mutex"))
	r.Handle("/pprof/heap", pprof.Handler("heap"))
	r.Handle("/pprof/block", pprof.Handler("block"))
	r.Handle("/pprof/allocs", pprof.Handler("allocs"))

	return r
}
