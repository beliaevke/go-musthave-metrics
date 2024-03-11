package client

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
)

var flagRunAddr string
var flagReportInterval int
var flagPollInterval int

type Config struct {
	envRunAddr        string `env:"ADDRESS"`
	envReportInterval int    `env:"REPORT_INTERVAL"`
	envPollInterval   int    `env:"POLL_INTERVAL"`
}

// parseFlags обрабатывает аргументы командной строки
// и сохраняет их значения в соответствующих переменных
func parseFlags() {
	// регистрируем переменную flagRunAddr
	// как аргумент -a со значением :8080 по умолчанию
	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "address and port to run server")
	// регистрируем переменную flagReportInterval
	// как аргумент -r со значением 10 по умолчанию
	flag.IntVar(&flagReportInterval, "r", 10, "report interval")
	// регистрируем переменную flagPollInterval
	// как аргумент -p со значением 2 по умолчанию
	flag.IntVar(&flagPollInterval, "p", 2, "poll interval")
	// парсим переданные серверу аргументы в зарегистрированные переменные
	flag.Parse()

	// для случаев, когда в переменных окружения присутствует непустое значение,
	// переопределим их, даже если они были переданы через аргументы командной строки
	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	if cfg.envRunAddr != "" {
		flagRunAddr = cfg.envRunAddr
	}
	if cfg.envReportInterval != 0 {
		flagReportInterval = cfg.envReportInterval
	}
	if cfg.envPollInterval != 0 {
		flagPollInterval = cfg.envPollInterval
	}
}
