package storage

import (
	"strconv"
)

type Repository interface {
	Add() error
}

type MemStorage struct {
	Gauges   map[string]float64
	Counters map[string]int64
}

var storage = newMemStorage()

func newMemStorage() MemStorage {
	return MemStorage{
		Gauges:   make(map[string]float64),
		Counters: make(map[string]int64),
	}
}

type GaugeMetric struct {
	Name  string
	Value string
}

func (metric GaugeMetric) Add() error {
	val, err := strconv.ParseFloat(metric.Value, 64)
	if err == nil {
		storage.Gauges[metric.Name] = val
	}
	return err
}

type CounterMetric struct {
	Name  string
	Value string
}

func (metric CounterMetric) Add() error {
	val, err := strconv.ParseInt(metric.Value, 10, 64)
	if err == nil {
		storage.Counters[metric.Name] += val
	}
	return err
}
