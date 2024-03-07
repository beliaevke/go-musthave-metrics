package main

import (
	"musthave-metrics/handlers"
	"net/http"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	metric := handlers.Metric{}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /update/{metricType}/{metricName}/{metricValue}", metric.Update)
	mux.HandleFunc(`/`, badrequest)
	return http.ListenAndServe(`:8080`, mux)
}

func badrequest(w http.ResponseWriter, r *http.Request) {
	// unknown request
	w.WriteHeader(http.StatusNotFound)
}
