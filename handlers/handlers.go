package handlers

import (
	"musthave-metrics/internal/storage"
	"net/http"
)

type Metric struct {
	metricType  string
	metricName  string
	metricValue string
}

func (m Metric) Update(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	m.setValue(r)
	if !m.isValid() || m.add() != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (m *Metric) setValue(r *http.Request) {
	m.metricType = r.PathValue("metricType")
	m.metricName = r.PathValue("metricName")
	m.metricValue = r.PathValue("metricValue")
}

func (m Metric) isValid() bool {
	if m.metricType == "" || m.metricName == "" || m.metricValue == "" {
		return false
	}
	if !(m.metricType == "counter" || m.metricType == "gauge") {
		return false
	}
	return true
}

func (m Metric) add() error {
	var Repository storage.Repository
	var err error
	if m.metricType == "gauge" {
		Repository = storage.GaugeMetric{Name: m.metricName, Value: m.metricValue}
	} else if m.metricType == "counter" {
		Repository = storage.CounterMetric{Name: m.metricName, Value: m.metricValue}
	}
	err = Repository.Add()
	return err
}
