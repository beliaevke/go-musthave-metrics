package handlers

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"musthave-metrics/cmd/agent/client"
	"musthave-metrics/internal/service"
	"musthave-metrics/internal/storage"
	"net/http"
	"os"
	"text/template"

	"github.com/go-chi/chi"
)

type Metric struct {
	metricType  string
	metricName  string
	metricValue string
}

type metricsContent struct {
	Rowsg string
	Rowsc string
}

func UpdateHandler() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		m := Metric{}
		w.Header().Set("Content-Type", "text/plain")
		m.setValue(r)
		if !m.isValid() || m.add() != nil {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}
	return http.HandlerFunc(fn)
}

func GetValueHandler() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		m := Metric{}
		m.setValue(r)
		val, err := m.getValue()
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte(val))
		if err != nil {
			log.Fatal(err)
		}
	}
	return http.HandlerFunc(fn)
}

func AllMetricsHandler() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		content := metricsContent{
			Rowsg: repo(Metric{metricType: "gauge"}).AllValuesHTML(),
			Rowsc: repo(Metric{metricType: "counter"}).AllValuesHTML(),
		}
		body, err := template.New("temp").Parse(metricstemplate())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
			err = body.Execute(w, content)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
		}
	}
	return http.HandlerFunc(fn)
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

func (m Metric) getValue() (string, error) {
	value, err := repo(m).GetValue()
	if value == "" && err == nil {
		err = fmt.Errorf("unknown metric")
	}
	return value, err
}

func repo(m Metric) (repository storage.Repository) {
	if m.metricType == "gauge" {
		repository = storage.GaugeMetric{Name: m.metricName, Value: m.metricValue}
	} else if m.metricType == "counter" {
		repository = storage.CounterMetric{Name: m.metricName, Value: m.metricValue}
	}
	return repository
}

func UpdateMetrics(locallink client.Locallink, mtype string, mname string, mvalue string) error {
	client := &http.Client{}
	url := service.MakeURL(locallink.RunAddr, locallink.Method, mtype, mname, mvalue)
	var body []byte
	request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", locallink.ContentType)
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	response.Body.Close()
	_, err = io.Copy(os.Stdout, response.Body)
	if err != nil {
		return err
	}
	return nil
}

func metricstemplate() string {
	return `<html>
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
				{{ .Rowsg }}
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
				{{ .Rowsc }}
			</tbody>
		</table>
	</body>
</html>`
}
