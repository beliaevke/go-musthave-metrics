package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"musthave-metrics/cmd/agent/client"
	"musthave-metrics/internal/logger"
	"musthave-metrics/internal/postgres"
	"musthave-metrics/internal/service"
	"musthave-metrics/internal/storage"
	"net/http"
	"os"
	"strconv"
	"text/template"
	"time"

	"github.com/go-chi/chi"
)

type Metric struct {
	metricType  string
	metricName  string
	metricValue string
}

type MetricsJSON struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
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

func UpdateJSONHandler(storeInterval int, fileStoragePath string) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			var metric MetricsJSON
			var buf bytes.Buffer
			// читаем тело запроса
			_, err := buf.ReadFrom(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			// десериализуем JSON в Visitor
			if err = json.Unmarshal(buf.Bytes(), &metric); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if metric.MType == "gauge" {
				err = repo(metric.ID, metric.MType, strconv.FormatFloat(*metric.Value, 'g', -1, 64)).Add()
			} else if metric.MType == "counter" {
				err = repo(metric.ID, metric.MType, strconv.FormatInt(*metric.Delta, 10)).Add()
			} else {
				http.Error(w, "unknown metric type", http.StatusInternalServerError)
				return
			}
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			resp, err := json.Marshal(metric)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			_, err = w.Write(resp)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			if storeInterval == 0 {
				storeMetric(metric, fileStoragePath)
			}
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

func GetValueJSONHandler() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			var metric MetricsJSON
			var buf bytes.Buffer
			// читаем тело запроса
			_, err := buf.ReadFrom(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			// десериализуем JSON в Visitor
			if err = json.Unmarshal(buf.Bytes(), &metric); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			val, err := repo(metric.ID, metric.MType, "").GetValue()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if val == "" {
				val = "0"
			}
			if metric.MType == "gauge" {
				gaugeValue, err := strconv.ParseFloat(val, 64)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				metric.Value = &gaugeValue
			} else if metric.MType == "counter" {
				counterValue, err := strconv.ParseInt(val, 10, 64)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				metric.Delta = &counterValue
			} else {
				http.Error(w, "unknown metric type", http.StatusInternalServerError)
				return
			}
			resp, err := json.Marshal(metric)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			_, err = w.Write(resp)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		}
	}
	return http.HandlerFunc(fn)
}

func AllMetricsHandler() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		content := metricsContent{
			Rowsg: repo("", "gauge", "").AllValuesHTML(),
			Rowsc: repo("", "counter", "").AllValuesHTML(),
		}
		body, err := template.New("temp").Parse(metricstemplate())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			err = body.Execute(w, content)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
		}
	}
	return http.HandlerFunc(fn)
}

func PingDBHandler(DatabaseDSN string) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		settings := postgres.NewPSQLStr(DatabaseDSN)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		err := settings.Ping(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
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
	err := repo(m.metricName, m.metricType, m.metricValue).Add()
	return err
}

func (m Metric) getValue() (string, error) {
	value, err := repo(m.metricName, m.metricType, m.metricValue).GetValue()
	if value == "" && err == nil {
		err = fmt.Errorf("unknown metric")
	}
	return value, err
}

func repo(metricName string, metricType string, metricValue string) (repository storage.Repository) {
	if metricType == "gauge" {
		repository = storage.GaugeMetric{Name: metricName, Value: metricValue}
	} else if metricType == "counter" {
		repository = storage.CounterMetric{Name: metricName, Value: metricValue}
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

func RestoreMetrics(fileStoragePath string) {
	m, err := readFile(fileStoragePath)
	if err != nil {
		logger.Warnf("Read file error: " + err.Error())
	}
	for i, v := range m {
		restoreMetric(v, i)
	}
}

func readFile(fileStoragePath string) ([]MetricsJSON, error) {
	data, err := os.ReadFile(fileStoragePath)
	if err != nil {
		return nil, err
	}
	m := make([]MetricsJSON, 0)
	reader := bytes.NewReader(data)
	if err := json.NewDecoder(reader).Decode(&m); err != nil {
		logger.Warnf("Read metric from file error: " + err.Error())
		return nil, err
	}
	return m, nil
}

func restoreMetric(metric MetricsJSON, line int) {
	var err error
	if metric.MType == "gauge" {
		err = repo(metric.ID, metric.MType, strconv.FormatFloat(*metric.Value, 'g', -1, 64)).Add()
	} else if metric.MType == "counter" {
		err = repo(metric.ID, metric.MType, strconv.FormatInt(*metric.Delta, 10)).Add()
	} else {
		logger.Warnf("Read file error: unknown metric type - " + metric.MType + ", line: " + strconv.Itoa(line))
	}
	if err != nil {
		logger.Warnf("Read file error: " + err.Error() + ", line: " + strconv.Itoa(line))
	}
}

func StoreMetrics(fileStoragePath string) {
	data, err := json.MarshalIndent(allMetricsJSON(), "", "   ")
	if err != nil {
		logger.Warnf("Write file error: " + err.Error())
	}
	// сохраняем данные в файл
	err = os.WriteFile(fileStoragePath, data, 0666)
	if err != nil {
		logger.Warnf("Write file error: " + err.Error())
	}
}

func storeMetric(m MetricsJSON, fileStoragePath string) {
	metric, err := json.MarshalIndent(m, "", "   ")
	if err != nil {
		logger.Warnf("Write file error: " + err.Error())
	}
	// сохраняем данные в файл
	err = os.WriteFile(fileStoragePath, metric, 0666)
	if err != nil {
		logger.Warnf("Write file error: " + err.Error())
	}
}

func allMetricsJSON() []MetricsJSON {
	var metrics []MetricsJSON
	storGauges := repo("", "gauge", "").GetValues().Gauges
	for name, val := range storGauges {
		metrics = append(metrics, MetricsJSON{ID: name, MType: "gauge", Value: &val})
	}
	storCounters := repo("", "counter", "").GetValues().Counters
	for name, del := range storCounters {
		metrics = append(metrics, MetricsJSON{ID: name, MType: "counter", Delta: &del})
	}
	return metrics
}
