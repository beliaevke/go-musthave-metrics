package service

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHashData(t *testing.T) {
	tests := []struct {
		name    string
		hashKey string
		want    error
	}{
		{
			name:    "1",
			hashKey: "test",
			want:    nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashData := NewHashData(tt.hashKey)
			assert.Equal(t, hashData.Key, tt.hashKey)
		})
	}
}

func TestMakeURL(t *testing.T) {
	tests := []struct {
		name    string
		runAddr string
		method  string
		mtype   string
		mname   string
		mvalue  string
		want    string
	}{
		{
			name:    "1",
			runAddr: "localhost",
			method:  "/update/",
			mtype:   "type",
			mname:   "name",
			mvalue:  "0",
			want:    "http://localhost/update/type/name/0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := MakeURL(tt.runAddr, tt.method, tt.mtype, tt.mname, tt.mvalue)
			assert.Equal(t, url, tt.want)
		})
	}
}

func TestMakeBatchUpdatesURL(t *testing.T) {
	tests := []struct {
		name    string
		runAddr string
		want    string
	}{
		{
			name:    "1",
			runAddr: "localhost",
			want:    "http://localhost/updates/",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := MakeBatchUpdatesURL(tt.runAddr)
			assert.Equal(t, url, tt.want)
		})
	}
}

func TestHashData_WithHashVerification(t *testing.T) {
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

			hd := NewHashData("HashKey")
			ts := httptest.NewServer(hd.WithHashVerification(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

func Example_getHash() {

	// Получаем hash data
	hd := NewHashData("hashKey")

	// Получаем данные
	data := []byte("data")

	// Формируем hash
	getHash(data, hd.Key)

}
