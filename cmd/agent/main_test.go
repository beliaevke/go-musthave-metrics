package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAgent(t *testing.T) {
	tests := []struct {
		name string
		want error
	}{
		{
			name: "1",
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := newAgent()
			assert.Equal(t, nil, err)
		})
	}
}

func TestInitMetrics(t *testing.T) {
	tests := []struct {
		name               string
		agent              agent
		wantCounterMetrics map[string]int64
		wantGaugeMetrics   map[string]string
	}{
		{
			name:               "1",
			agent:              agent{},
			wantCounterMetrics: make(map[string]int64, 1),
			wantGaugeMetrics:   make(map[string]string),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.agent.initMetrics()
			assert.Equal(t, tt.agent.CounterMetrics, tt.wantCounterMetrics)
			assert.Equal(t, tt.agent.GaugeMetrics, tt.wantGaugeMetrics)
		})
	}
}
