package handlers

import (
	"fmt"
	"io"
	"musthave-metrics/cmd/agent/config"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
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
			name:           "Valid pattern with HTTP POST",
			pattern:        "/update/{metricType}/{metricName}/{metricValue}",
			shouldPanic:    false,
			method:         "POST",
			path:           "/update/gauge/someMetric/11.11",
			expectedBody:   "with-prefix POST",
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

			r1 := chi.NewRouter()
			r1.Handle(tc.pattern, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				m := Metric{}
				w.Header().Set("Content-Type", "text/plain")
				m.setValue(r)
				if !m.isValid() || m.add() != nil {
					w.WriteHeader(http.StatusBadRequest)
				} else {
					w.WriteHeader(http.StatusOK)
					_, err := w.Write([]byte(tc.expectedBody))
					if err != nil {
						t.Errorf("Write failed: %v", err)
					}
				}
			}))

			// Test that HandleFunc also handles method patterns
			r2 := chi.NewRouter()
			r2.HandleFunc(tc.pattern, func(w http.ResponseWriter, r *http.Request) {
				m := Metric{}
				w.Header().Set("Content-Type", "text/plain")
				m.setValue(r)
				if !m.isValid() || m.add() != nil {
					w.WriteHeader(http.StatusBadRequest)
				} else {
					w.WriteHeader(http.StatusOK)
					_, err := w.Write([]byte(tc.expectedBody))
					if err != nil {
						t.Errorf("Write failed: %v", err)
					}
				}
			})

			if !tc.shouldPanic {
				for _, r := range []chi.Router{r1, r2} {
					// Use testRequest for valid patterns
					ts := httptest.NewServer(r)
					defer ts.Close()
					resp, body := testRequest(t, ts, tc.method, tc.path, nil)
					defer resp.Body.Close()
					if body != tc.expectedBody || resp.StatusCode != tc.expectedStatus {
						t.Errorf("Expected status %d and body %s; got status %d and body %s for pattern %s",
							tc.expectedStatus, tc.expectedBody, resp.StatusCode, body, tc.pattern)
					}
				}
			}
		})
	}
}

func testRequest(t *testing.T, ts *httptest.Server, method, path string, body io.Reader) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, body)
	if err != nil {
		t.Fatal(err)
		return nil, ""
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
		return nil, ""
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
		return nil, ""
	}
	defer resp.Body.Close()

	return resp, string(respBody)
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
