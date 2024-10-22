package main

import (
	"log"
	"os"
	"runtime"
	rpprof "runtime/pprof"
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

func BenchmarkSetGaugeMemStatsMetrics(b *testing.B) {
	agent, err := newAgent()
	if err != nil {
		log.Fatal(err)
	}
	agent.initMetrics()
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	fmem, err := os.Create("profiles/base1.pprof")
	if err != nil {
		panic(err)
	}
	defer fmem.Close()
	runtime.GC() // получаем статистику по использованию памяти
	if err := rpprof.WriteHeapProfile(fmem); err != nil {
		panic(err)
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		setGaugeMemStatsMetrics(memStats, agent)
	}
}

func BenchmarkSetGaugeMemStatsMetricsNew(b *testing.B) {
	agent, err := newAgent()
	if err != nil {
		log.Fatal(err)
	}
	agent.initMetrics()
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	fmem, err := os.Create("profiles/res1.pprof")
	if err != nil {
		panic(err)
	}
	defer fmem.Close()
	runtime.GC() // получаем статистику по использованию памяти
	if err := rpprof.WriteHeapProfile(fmem); err != nil {
		panic(err)
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		setGaugeMemStatsMetricsNew(memStats, agent)
	}
}
