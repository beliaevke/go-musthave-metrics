package handlers

import (
	"testing"

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
