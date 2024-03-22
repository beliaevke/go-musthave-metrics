package main

import (
	"fmt"
	"log"
	"musthave-metrics/handlers"
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
	mux := chi.NewMux()
	mux.HandleFunc("/update/{metricType}/{metricName}/{metricValue}", handlers.Update)
	mux.HandleFunc("/value/{metricType}/{metricName}", handlers.GetValue)
	mux.HandleFunc("/", handlers.AllMetrics)
	fmt.Println("Running server on ", flagRunAddr)
	return http.ListenAndServe(flagRunAddr, mux)
}
