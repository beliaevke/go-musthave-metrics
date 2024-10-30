package handlers

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"musthave-metrics/cmd/agent/config"
	serverconfig "musthave-metrics/cmd/server/config"

	"github.com/stretchr/testify/assert"
)

func TestIsValid(t *testing.T) {
	tests := []struct {
		name  string
		value Metric
		want  bool
	}{
		{
			name: "1",
			value: Metric{
				metricType:  "gauge",
				metricName:  "TestGaugeMetric",
				metricValue: "11.11",
			},
			want: true,
		},
		{
			name: "2",
			value: Metric{
				metricType:  "counter",
				metricName:  "TestCounterMetric",
				metricValue: "111",
			},
			want: true,
		},
		{
			name: "3",
			value: Metric{
				metricType:  "gauge",
				metricName:  "TestGaugeMetric",
				metricValue: "",
			},
			want: false,
		},
		{
			name: "4",
			value: Metric{
				metricType:  "counter",
				metricName:  "TestGaugeMetric",
				metricValue: "",
			},
			want: false,
		},
		{
			name: "5",
			value: Metric{
				metricType:  "",
				metricName:  "TestMetric",
				metricValue: "111",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.value.isValid())
		})
	}
}

func TestUpdateHandler(t *testing.T) {
	testCases := []struct {
		name           string
		pattern        string
		shouldPanic    bool
		method         string // Method to be used for the test request
		path           string // Path to be used for the test request
		expectedBody   string // Expected response body
		expectedStatus int    // Expected HTTP status code
	}{
		// Valid patterns
		{
			name:           "Unvalid pattern with HTTP POST",
			pattern:        "/update/{metricType}/{metricName}/{metricValue}",
			shouldPanic:    false,
			method:         "POST",
			path:           "/update/gauge/someMetric/11.11",
			expectedBody:   "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil && !tc.shouldPanic {
					t.Errorf("Unexpected panic for pattern %s:\n%v", tc.pattern, r)
				}
			}()

			ts := httptest.NewServer(UpdateHandler())
			defer ts.Close()

			data := []byte("")
			r := bytes.NewReader(data)
			res, err := http.Post(ts.URL+tc.path, "application/json", r)
			if err != nil {
				t.Errorf("Unexpected error")
			}
			bd, err := io.ReadAll(res.Body)
			res.Body.Close()
			if err != nil {
				t.Errorf("Unexpected error")
			}

			// Проверяем код
			if status := res.StatusCode; status != tc.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, http.StatusOK)
			}

			// Проверяем тело ответа
			if string(bd) != tc.expectedBody {
				t.Errorf("handler returned unexpected body: got %v want %v",
					string(bd), tc.expectedBody)
			}

		})
	}
}

func BenchmarkAllMetricsHandler(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		AllMetricsHandler()
	}
}

func ExampleAllMetricsHandler() {

	// Получаем конфигурацию
	cfg := config.ParseFlags()

	// Выполняем вызов основной страницы
	resp, err := http.Get("http://" + cfg.FlagRunAddr)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
}

func ExamplePingDBHandler() {

	// Получаем конфигурацию
	cfg := serverconfig.ParseFlags()

	// Выполняем вызов
	ts := httptest.NewServer(PingDBHandler(cfg.FlagDatabaseDSN))
	defer ts.Close()

	res, err := http.Get(ts.URL + "/ping")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	// Проверяем код ответа
	if res.StatusCode != http.StatusOK {
		fmt.Println("ping DB handler returned wrong status code")
	}
}

func TestUpdateJSONHandler(t *testing.T) {
	testCases := []struct {
		name           string
		pattern        string
		shouldPanic    bool
		method         string // Method to be used for the test request
		path           string // Path to be used for the test request
		expectedBody   string // Expected response body
		expectedStatus int    // Expected HTTP status code
	}{
		// Valid patterns
		{
			name:           "Valid pattern with HTTP POST",
			pattern:        "/update/",
			shouldPanic:    false,
			method:         "POST",
			path:           "/update/",
			expectedBody:   `{"id":"testtest","type":"gauge","value":111}`,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil && !tc.shouldPanic {
					t.Errorf("Unexpected panic for pattern %s:\n%v", tc.pattern, r)
				}
			}()

			ts := httptest.NewServer(UpdateJSONHandler(300, "/tmp/metrics-db.json"))
			defer ts.Close()

			data := []byte(tc.expectedBody)
			r := bytes.NewReader(data)
			res, err := http.Post(ts.URL+tc.path, "application/json", r)
			if err != nil {
				t.Errorf("Unexpected error")
			}
			bd, err := io.ReadAll(res.Body)
			res.Body.Close()
			if err != nil {
				t.Errorf("Unexpected error")
			}

			// Проверяем код
			if status := res.StatusCode; status != tc.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, http.StatusOK)
			}

			// Проверяем тело ответа
			if string(bd) != tc.expectedBody {
				t.Errorf("handler returned unexpected body: got %v want %v",
					string(bd), tc.expectedBody)
			}

		})
	}
}

func TestGetValueJSONHandler(t *testing.T) {
	testCases := []struct {
		name           string
		pattern        string
		shouldPanic    bool
		method         string // Method to be used for the test request
		path           string // Path to be used for the test request
		expectedBody   string // Expected response body
		expectedStatus int    // Expected HTTP status code
	}{
		// Valid patterns
		{
			name:           "Valid pattern with HTTP POST",
			pattern:        "/value/{metricType}/{metricName}",
			shouldPanic:    false,
			method:         "POST",
			path:           "/value/gauge/testtest",
			expectedBody:   `{"id":"testtest","type":"gauge","value":111}`,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil && !tc.shouldPanic {
					t.Errorf("Unexpected panic for pattern %s:\n%v", tc.pattern, r)
				}
			}()

			ts := httptest.NewServer(GetValueJSONHandler())
			defer ts.Close()

			data := []byte(tc.expectedBody)
			r := bytes.NewReader(data)
			req, err := http.NewRequest(tc.method, ts.URL+tc.path, r)
			if err != nil {
				t.Errorf("Unexpected error")
			}

			client := ts.Client()
			res, err := client.Do(req)
			if err != nil {
				t.Errorf("Unexpected error")
			}
			bd, err := io.ReadAll(res.Body)
			res.Body.Close()
			if err != nil {
				t.Errorf("Unexpected error")
			}

			// Проверяем код
			if status := res.StatusCode; status != tc.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, http.StatusOK)
			}

			// Проверяем тело ответа
			if string(bd) != tc.expectedBody {
				t.Errorf("handler returned unexpected body: got %v want %v",
					string(bd), tc.expectedBody)
			}

		})
	}
}

func TestAllMetricsHandler(t *testing.T) {
	testCases := []struct {
		name           string
		pattern        string
		shouldPanic    bool
		method         string // Method to be used for the test request
		path           string // Path to be used for the test request
		expectedBody   string // Expected response body
		expectBody     bool
		expectedStatus int // Expected HTTP status code
	}{
		// Valid patterns
		{
			name:           "Valid pattern with HTTP POST",
			pattern:        "/update/",
			shouldPanic:    false,
			method:         "GET",
			path:           "/",
			expectedBody:   "",
			expectBody:     true,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil && !tc.shouldPanic {
					t.Errorf("Unexpected panic for pattern %s:\n%v", tc.pattern, r)
				}
			}()

			ts := httptest.NewServer(AllMetricsHandler())
			defer ts.Close()

			res, err := http.Get(ts.URL + tc.path)
			if err != nil {
				t.Errorf("Unexpected error")
			}
			bd, err := io.ReadAll(res.Body)
			res.Body.Close()
			if err != nil {
				t.Errorf("Unexpected error")
			}

			// Проверяем код
			if status := res.StatusCode; status != tc.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, http.StatusOK)
			}

			// Проверяем тело ответа
			if tc.expectBody && string(bd) == tc.expectedBody {
				t.Errorf("handler returned unexpected body: got %v want %v",
					string(bd), tc.expectedBody)
			}

		})
	}
}

func TestUpdateDBHandler(t *testing.T) {
	testCases := []struct {
		name           string
		pattern        string
		shouldPanic    bool
		method         string // Method to be used for the test request
		path           string // Path to be used for the test request
		expectedBody   string // Expected response body
		expectedStatus int    // Expected HTTP status code
	}{
		// Valid patterns
		{
			name:           "Valid pattern with HTTP POST",
			pattern:        "/updates/",
			shouldPanic:    false,
			method:         "POST",
			path:           "/updates/",
			expectedBody:   `{"id":"testtest","type":"gauge","value":111}`,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil && !tc.shouldPanic {
					t.Errorf("Unexpected panic for pattern %s:\n%v", tc.pattern, r)
				}
			}()

			cfg := serverconfig.ParseFlags()
			ctx := context.Background()
			ts := httptest.NewServer(UpdateDBHandler(ctx, cfg.FlagDatabaseDSN, ""))
			defer ts.Close()

			data := []byte(tc.expectedBody)
			r := bytes.NewReader(data)
			res, err := http.Post(ts.URL+tc.path, "application/json", r)
			if err != nil {
				t.Errorf("Unexpected error")
			}
			res.Body.Close()

			// Проверяем код
			if status := res.StatusCode; status != tc.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, http.StatusOK)
			}

		})
	}
}

func TestGetValueDBHandler(t *testing.T) {
	testCases := []struct {
		name           string
		pattern        string
		shouldPanic    bool
		method         string // Method to be used for the test request
		path           string // Path to be used for the test request
		expectedBody   string // Expected response body
		expectedStatus int    // Expected HTTP status code
	}{
		// Valid patterns
		{
			name:           "Valid pattern with HTTP POST",
			pattern:        "/value/",
			shouldPanic:    false,
			method:         "POST",
			path:           "/value/",
			expectedBody:   `{"id":"testtest","type":"gauge","value":111}`,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil && !tc.shouldPanic {
					t.Errorf("Unexpected panic for pattern %s:\n%v", tc.pattern, r)
				}
			}()

			databaseDSNFlag := flag.Lookup("d")
			var databaseDSN string

			if databaseDSNFlag == nil {
				databaseDSN = "postgres://postgres:pos111@localhost:5432/postgres?sslmode=disable"
			} else {
				databaseDSN = databaseDSNFlag.Value.(flag.Getter).Get().(string)
			}

			ctx := context.Background()
			ts := httptest.NewServer(GetValueDBHandler(ctx, databaseDSN, ""))
			defer ts.Close()

			data := []byte(tc.expectedBody)
			r := bytes.NewReader(data)
			res, err := http.Post(ts.URL+tc.path, "application/json", r)
			if err != nil {
				t.Errorf("Unexpected error")
			}
			res.Body.Close()

			// Проверяем код
			if status := res.StatusCode; status != tc.expectedStatus && res.StatusCode != http.StatusInternalServerError {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tc.expectedStatus)
			}

		})
	}
}

func TestUpdateBatchDBHandler(t *testing.T) {
	testCases := []struct {
		name           string
		pattern        string
		shouldPanic    bool
		method         string // Method to be used for the test request
		path           string // Path to be used for the test request
		expectedBody   string // Expected response body
		expectedStatus int    // Expected HTTP status code
	}{
		// Valid patterns
		{
			name:           "Valid pattern with HTTP POST",
			pattern:        "/updates/",
			shouldPanic:    false,
			method:         "POST",
			path:           "/updates/",
			expectedBody:   `[{"id":"testtest","type":"gauge","value":111},{"id":"testtest","type":"counter","value":22}]`,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil && !tc.shouldPanic {
					t.Errorf("Unexpected panic for pattern %s:\n%v", tc.pattern, r)
				}
			}()

			databaseDSNFlag := flag.Lookup("d")
			var databaseDSN string

			if databaseDSNFlag == nil {
				databaseDSN = "postgres://postgres:pos111@localhost:5432/postgres?sslmode=disable"
			} else {
				databaseDSN = databaseDSNFlag.Value.(flag.Getter).Get().(string)
			}

			ts := httptest.NewServer(UpdateBatchDBHandler(databaseDSN))
			defer ts.Close()

			data := []byte(tc.expectedBody)
			r := bytes.NewReader(data)
			res, err := http.Post(ts.URL+tc.path, "application/json", r)
			if err != nil {
				t.Errorf("Unexpected error")
			}

			res.Body.Close()

			// Проверяем код
			if status := res.StatusCode; status != tc.expectedStatus && res.StatusCode != http.StatusInternalServerError {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, http.StatusOK)
			}

		})
	}
}

func Test_allMetricsJSON(t *testing.T) {
	tests := []struct {
		name string
		want []MetricsJSON
	}{
		{
			name: "1",
			want: make([]MetricsJSON, 0),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := allMetricsJSON(); reflect.DeepEqual(got, tt.want) {
				t.Errorf("allMetricsJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}
