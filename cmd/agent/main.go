package main

import (
	"fmt"
	"log"
	"math/rand"
	"musthave-metrics/cmd/agent/client"
	"musthave-metrics/handlers"
	"reflect"
	"runtime"
	"strconv"
	"time"
)

type agent struct {
	CounterMetrics map[string]int64
	GaugeMetrics   map[string]string
	client         client.Locallink
}

func (agent *agent) run() {
	agent.printAgentLog("Start")
	agent.initMetrics()
	agent.pollMetrics()
	agent.reportMetrics()
	time.Sleep(30 * time.Second)
	agent.printAgentLog("Stop")
}

func newAgent() (*agent, error) {
	agent := &agent{
		client: client.Locallink{},
	}
	return agent, agent.client.Run()
}

func main() {
	agent, err := newAgent()
	if err != nil {
		log.Fatal(err)
	}
	agent.run()
}

func (agent *agent) initMetrics() {
	agent.CounterMetrics = make(map[string]int64, 1)
	agent.GaugeMetrics = make(map[string]string)
}

func (agent *agent) pollMetrics() {
	f := func() {
		agent.setMetrics()
		agent.pollMetrics()
		agent.printMetricsLog("<= Read")
	}
	time.AfterFunc(time.Duration(agent.client.PollInterval)*time.Second, f)
}

func (agent *agent) reportMetrics() {
	f := func() {
		agent.pushMetrics()
		agent.reportMetrics()
		agent.printMetricsLog("=> Push")
	}
	time.AfterFunc(time.Duration(agent.client.ReportInterval)*time.Second, f)
}

func (agent *agent) pushMetrics() {
	var err error
	for name, val := range agent.CounterMetrics {
		err = handlers.UpdateMetrics(agent.client, "counter", name, strconv.FormatInt(val, 10))
	}
	for name, val := range agent.GaugeMetrics {
		err = handlers.UpdateMetrics(agent.client, "gauge", name, val)
	}
	if err != nil {
		agent.printErrorLog(err)
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
	setGaugeMemStatsMetrics(memStats, agent)
}

func setGaugeMemStatsMetrics(s interface{}, agent *agent) {
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

func (agent *agent) printAgentLog(operation string) {
	fmt.Printf(
		"%s === %s agent ===\n",
		time.Now().Format(time.DateTime),
		operation,
	)
}

func (agent *agent) printErrorLog(error) {
	if err := recover(); err != nil {
		fmt.Printf(
			"%s xxx Error: %s \n",
			time.Now().Format(time.DateTime),
			err,
		)
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
