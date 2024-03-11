package handlers

import (
	"fmt"
	"musthave-metrics/internal/storage"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
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

func (m Metric) GetValue(w http.ResponseWriter, r *http.Request) {
	m.setValue(r)
	val, err := m.getValue()
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(val))
}

func AllMetrics(w http.ResponseWriter, r *http.Request) {
	body := allMetricsBody(
		repo(Metric{metricType: "gauge"}).AllValuesHTML(),
		repo(Metric{metricType: "counter"}).AllValuesHTML(),
	)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(body))
}

func (m *Metric) setValue(r *http.Request) {
	m.metricType = chi.URLParam(r, "metricType")
	m.metricName = chi.URLParam(r, "metricName")
	m.metricValue = chi.URLParam(r, "metricValue")
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
	err := repo(m).Add()
	return err
}

func (m Metric) getValue() (value string, err error) {
	value, err = repo(m).GetValue()
	if value == "" && err == nil {
		err = fmt.Errorf("unknown metric")
	}
	return
}

func repo(m Metric) (repository storage.Repository) {
	if m.metricType == "gauge" {
		repository = storage.GaugeMetric{Name: m.metricName, Value: m.metricValue}
	} else if m.metricType == "counter" {
		repository = storage.CounterMetric{Name: m.metricName, Value: m.metricValue}
	}
	return
}

func allMetricsBody(rowsg string, rowsc string) string {
	body :=
		`<html>
		<head>
		<title></title>
		</head>
		<body>
			<table border="1" cellpadding="1" cellspacing="1" style="width: 500px">
				<thead>
					<tr>
						<th scope="col">Gauge metric</th>
						<th scope="col">Value</th>
					</tr>
				</thead>
				<tbody>
					%rowsg
				</tbody>
			</table>
			<table border="1" cellpadding="1" cellspacing="1" style="width: 500px">
				<thead>
					<tr>
						<th scope="col">Counter metric</th>
						<th scope="col">Value</th>
					</tr>
				</thead>
				<tbody>
					%rowsc
				</tbody>
			</table>
		</body>
	</html>`
	body = strings.ReplaceAll(body, "%rowsg", rowsg)
	body = strings.ReplaceAll(body, "%rowsc", rowsc)
	return body
}
