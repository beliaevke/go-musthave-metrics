package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os/signal"
	"reflect"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"

	"musthave-metrics/cmd/agent/client"
	"musthave-metrics/handlers"
	"musthave-metrics/internal/logger"
	"musthave-metrics/internal/postgres"
)

var (
	buildVersion, buildDate, buildCommit string = "N/A", "N/A", "N/A"
)

type agent struct {
	CounterMetrics map[string]int64
	GaugeMetrics   map[string]string
	client         client.Locallink
	notifyCtx      context.Context
	shutdown       context.CancelFunc
}

func (agent *agent) run() {
	agent.printAgentLog("Start")
	agent.initMetrics()
	go agent.pollMetrics()
	go agent.pollUtilMetrics()
	go agent.reportMetrics()

	for i := 0; i < 30; i++ {
		select {
		case <-agent.notifyCtx.Done():
			time.Sleep(time.Duration(agent.client.PollInterval) * time.Second * 3)
			agent.printAgentLog("Stop (on signal)")
			agent.shutdown()
			return
		default:
			time.Sleep(1 * time.Second)
		}
	}

	agent.printAgentLog("Stop")
}

func newAgent() (*agent, error) {
	agent := &agent{
		client: client.Locallink{},
	}
	return agent, agent.client.Run()
}

func main() {

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	logger.BuildInfo(buildVersion, buildDate, buildCommit)
	agent, err := newAgent()
	if err != nil {
		log.Fatal(err)
	}

	agent.notifyCtx = ctx
	agent.shutdown = cancel
	agent.run()

	/*
		fmem, err := os.Create("profiles/res.pprof")
		if err != nil {
			panic(err)
		}
		defer fmem.Close()
		runtime.GC() // получаем статистику по использованию памяти
		if err := rpprof.WriteHeapProfile(fmem); err != nil {
			panic(err)
		}
	*/
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
	select {
	case <-agent.notifyCtx.Done():
		logger.Infof("Получен сигнал отмены, завершаем операции pollMetrics")
	default:
		time.AfterFunc(time.Duration(agent.client.PollInterval)*time.Second, f)
	}
}

func (agent *agent) pollUtilMetrics() {
	f := func() {
		agent.setUtilMetrics()
		agent.pollUtilMetrics()
		agent.printMetricsLog("<= Util")
	}
	select {
	case <-agent.notifyCtx.Done():
		logger.Infof("Получен сигнал отмены, завершаем операции pollUtilMetrics")
	default:
		time.AfterFunc(time.Duration(agent.client.PollInterval)*time.Second, f)
	}
}

func (agent *agent) reportMetrics() {
	f := func() {
		agent.pushMetrics()
		agent.pushBatchMetricsWithWorkers()
		agent.reportMetrics()
		agent.printMetricsLog("=> Push")
	}
	select {
	case <-agent.notifyCtx.Done():
		logger.Infof("Получен сигнал отмены, завершаем операции reportMetrics")
	default:
		time.AfterFunc(time.Duration(agent.client.ReportInterval)*time.Second, f)
	}
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

func (agent *agent) pushBatchMetrics() {
	var metrics []postgres.Metrics
	var err error
	for name, val := range agent.CounterMetrics {
		metrics = append(metrics,
			postgres.Metrics{
				ID:    name,
				MType: "counter",
				Delta: &val,
			},
		)
	}
	for name, val := range agent.GaugeMetrics {
		gaugeValue, errprs := strconv.ParseFloat(val, 64)
		if errprs != nil {
			agent.printErrorLog(errprs)
			continue
		}
		metrics = append(metrics,
			postgres.Metrics{
				ID:    name,
				MType: "gauge",
				Value: &gaugeValue,
			},
		)
	}
	err = handlers.UpdateBatchMetrics(agent.client, metrics)
	if err != nil {
		agent.printErrorLog(err)
	}
}

func (agent *agent) pushBatchMetricsWithWorkers() {
	numJobs := agent.client.RateLimit
	// создаем буферизованный канал для принятия задач в воркер
	jobs := make(chan int, numJobs)
	// создаем буферизованный канал для отправки результатов
	results := make(chan int, numJobs)
	// создаем и запускаем 3 воркера, это и есть пул,
	// передаем id, это для наглядности, канал задач и канал результатов
	for w := 1; w <= 3; w++ {
		go pushWorkerBatchMetrics(agent, w, jobs, results)
	}
	// в канал задач отправляем какие-то данные
	// задач у нас 5, а воркера 3, значит одновременно решается только 3 задачи
	for j := 1; j <= numJobs; j++ {
		jobs <- j
	}
	// как вы помните, закрываем канал на стороне отправителя
	close(jobs)
	// забираем из канала результатов результаты
	// можно присваивать переменной, или выводить на экран, но мы не будем
	for a := 1; a <= numJobs; a++ {
		<-results
	}
}

func pushWorkerBatchMetrics(agent *agent, id int, jobs <-chan int, results chan<- int) {
	for j := range jobs {
		fmt.Println("рабочий процесс", id, "запущена задача", j)
		agent.pushBatchMetrics()
		fmt.Println("рабочий процесс", id, "закончена задача", j)
		results <- j + 1
	}
}

func (agent *agent) setMetrics() {
	var mu sync.Mutex
	mu.Lock()
	// counter
	agent.CounterMetrics["PollCount"] += 1
	// gauge
	agent.GaugeMetrics["RandomValue"] = strconv.FormatFloat(rand.Float64(), 'g', -1, 64)
	// memStats (gauge)
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	setGaugeMemStatsMetrics(memStats, agent)
	//setGaugeMemStatsMetricsNew(memStats, agent)
	mu.Unlock()
}

func (agent *agent) setUtilMetrics() {
	var mu sync.Mutex
	// Util gauge
	memstats, err := mem.VirtualMemory()
	if err != nil {
		agent.printErrorLog(err)
		return
	}
	cpustat, err := cpu.Percent(0, false)
	if err != nil {
		agent.printErrorLog(err)
		return
	}
	mu.Lock()
	agent.GaugeMetrics["TotalMemory"] = strconv.FormatUint(memstats.Total, 10)
	agent.GaugeMetrics["FreeMemory"] = strconv.FormatUint(memstats.Free, 10)
	for i := 0; i < len(cpustat); i++ {
		agent.GaugeMetrics["CPUutilization"+strconv.Itoa(i)] = strconv.FormatFloat(cpustat[i], 'g', -1, 64)
	}
	mu.Unlock()
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

func setGaugeMemStatsMetricsNew(s interface{}, agent *agent) {
	valOf := reflect.ValueOf(s)
	for i := 0; i < valOf.NumField(); i++ {
		var value string
		valField := valOf.Field(i)
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
		agent.GaugeMetrics[valOf.Type().Field(i).Name] = value
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
