package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMemStorage(t *testing.T) {
	tests := []struct {
		name string
		want MemStorage
	}{
		{
			name: "1",
			want: MemStorage{
				Gauges:   make(map[string]float64),
				Counters: make(map[string]int64),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, newMemStorage())
		})
	}
}

func TestAddGaugeMetric(t *testing.T) {
	tests := []struct {
		name  string
		value GaugeMetric
		want  error
	}{
		{
			name: "1",
			value: GaugeMetric{
				Name:  "TestGaugeMetric",
				Value: "11.11",
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.value.Add())
		})
	}
}

func TestAddCounterMetric(t *testing.T) {
	tests := []struct {
		name  string
		value CounterMetric
		want  error
	}{
		{
			name: "1",
			value: CounterMetric{
				Name:  "TestCounterMetric",
				Value: "111",
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.value.Add())
		})
	}
}

func TestCounterMetric_AllValuesHTML(t *testing.T) {
	tests := []struct {
		name     string
		metric   CounterMetric
		wantRows string
	}{
		{
			name: "1",
			metric: CounterMetric{
				Name:  "TestCounterMetric",
				Value: "111",
			},
			wantRows: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotRows := tt.metric.AllValuesHTML(); gotRows == tt.wantRows {
				t.Errorf("CounterMetric.AllValuesHTML() = %v, want %v", gotRows, tt.wantRows)
			}
		})
	}
}
