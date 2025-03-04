package logger

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWithLogging(t *testing.T) {
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
			pattern:        "/",
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

			ts := httptest.NewServer(WithLogging(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, "Hello, client")
			})))
			defer ts.Close()

			res, err := http.Get(ts.URL)
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

func TestServerRunningInfo(t *testing.T) {
	type args struct {
		RunAddr string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "1",
			args: args{
				RunAddr: "localhost",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ServerRunningInfo(tt.args.RunAddr)
		})
	}
}

func TestBuildInfo(t *testing.T) {
	type args struct {
		buildVersion string
		buildDate    string
		buildCommit  string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "1",
			args: args{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			BuildInfo(tt.args.buildVersion, tt.args.buildDate, tt.args.buildCommit)
		})
	}
}
