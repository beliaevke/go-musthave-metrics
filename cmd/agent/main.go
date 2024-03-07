package main

import (
	"fmt"
	"math/rand"
	"musthave-metrics/cmd/agent/client"
	"reflect"
	"runtime"
	"strconv"
	"time"
)

type agent struct {
	CounterMetrics map[string]int64
	GaugeMetrics   map[string]string
	pollInterval   int
	reportInterval int
}

func (agent *agent) run() {
	fmt.Printf("%s === Start agent ===\n", time.Now().Format(time.DateTime))
	agent.CounterMetrics = make(map[string]int64, 1)
	agent.GaugeMetrics = make(map[string]string)
	agent.PollMetrics()
	agent.ReportMetrics()
	time.Sleep(20 * time.Second)
	fmt.Printf("%s === Stop agent ===\n", time.Now().Format(time.DateTime))
}

func newAgent(pollInterval int, reportInterval int) (*agent, error) {
	return &agent{
		pollInterval:   pollInterval,
		reportInterval: reportInterval,
	}, nil
}

func main() {
	agent, err := newAgent(2, 5)
	if err != nil {
		panic(err)
	}
	agent.run()
}

func (agent *agent) PollMetrics() {
	f := func() {
		agent.setMetrics()
		agent.PollMetrics()
		agent.printMetricsLog("<= Read")
	}
	time.AfterFunc(time.Duration(agent.pollInterval)*time.Second, f)
}

func (agent *agent) ReportMetrics() {
	f := func() {
		agent.PushMetrics()
		agent.ReportMetrics()
		agent.printMetricsLog("=> Push")
	}
	time.AfterFunc(time.Duration(agent.reportInterval)*time.Second, f)
}

func (agent *agent) PushMetrics() {
	lh := client.Localhost{}
	for name, val := range agent.CounterMetrics {
		lh.UpdateMetrics("counter", name, strconv.FormatInt(val, 10))
	}
	for name, val := range agent.GaugeMetrics {
		lh.UpdateMetrics("gauge", name, val)
	}
}

func (agent *agent) setMetrics() {
	// counter
	agent.CounterMetrics["PollCount"] += 1
	// gauge
	agent.GaugeMetrics["RandomValue"] = strconv.FormatFloat(rand.Float64(), 'g', -1, 64)
	// memStats (gauge)
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	SetGaugeMemStatsMetrics(memStats, agent)
}

func SetGaugeMemStatsMetrics(s interface{}, agent *agent) {
	valOf := reflect.ValueOf(s)
	typOf := reflect.TypeOf(s)
	for i := 0; i < valOf.NumField(); i++ {
		var value string
		valField := valOf.Field(i)
		typField := typOf.Field(i)
		switch valField.Interface().(type) {
		case float64:
			value = strconv.FormatFloat(valField.Interface().(float64), 'g', -1, 64)
		case uint32:
			value = fmt.Sprint(valField.Interface().(uint32))
		case uint64:
			value = strconv.FormatUint(valField.Interface().(uint64), 10)
		default:
			value = "0"
		}
		agent.GaugeMetrics[typField.Name] = value
	}
}

func (agent *agent) printMetricsLog(operation string) {
	fmt.Printf(
		"%s %s metrics (Poll count: %d)\n",
		time.Now().Format(time.DateTime),
		operation,
		agent.CounterMetrics["PollCount"],
	)
}
