package storage

import (
	"fmt"
	"strconv"
)

type Repository interface {
	Add() error
	GetValue() (string, error)
	AllValuesHTML() string
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

func (metric GaugeMetric) GetValue() (value string, err error) {
	value = "0"
	val, ok := storage.Gauges[metric.Name]
	if ok {
		value = strconv.FormatFloat(val, 'g', -1, 64)
	}
	return
}

func (metric GaugeMetric) AllValuesHTML() (rows string) {
	for name, val := range storage.Gauges {
		rows += fmt.Sprintf("<tr><th>%v</th><th>%v</th></tr>", name, val)
	}
	return
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

func (metric CounterMetric) GetValue() (value string, err error) {
	value = "0"
	val, ok := storage.Counters[metric.Name]
	if ok {
		value = strconv.FormatInt(val, 10)
	}
	return
}

func (metric CounterMetric) AllValuesHTML() (rows string) {
	for name, val := range storage.Counters {
		rows += fmt.Sprintf("<tr><th>%v</th><th>%v</th></tr>", name, val)
	}
	return
}
