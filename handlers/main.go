package handlers

import (
	"musthave-metrics/internal/storage"
	"net/http"
)

type Metric struct {
	metric_type  string
	metric_name  string
	metric_value string
}

func (m Metric) Update(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	m.setValue(r)
	if !m.isValid() || m.add() != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (m *Metric) setValue(r *http.Request) {
	m.metric_type = r.PathValue("metric_type")
	m.metric_name = r.PathValue("metric_name")
	m.metric_value = r.PathValue("metric_value")
}

func (m Metric) isValid() bool {
	if m.metric_type == "" || m.metric_name == "" || m.metric_value == "" {
		return false
	}
	if !(m.metric_type == "counter" || m.metric_type == "gauge") {
		return false
	}
	return true
}

func (m Metric) add() error {
	var Repository storage.Repository
	var err error
	if m.metric_type == "gauge" {
		Repository = storage.GaugeMetric{Name: m.metric_name, Value: m.metric_value}
	} else if m.metric_type == "counter" {
		Repository = storage.CounterMetric{Name: m.metric_name, Value: m.metric_value}
	}
	err = Repository.Add()
	return err
}
