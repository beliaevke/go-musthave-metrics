// Package config предназначен для методов конфигурации.
package config

import (
	"bytes"
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/caarlos0/env/v6"
)

type ClientFlags struct {
	FlagRunAddr        string `json:"address"`
	FlagReportInterval int    `json:"report_interval"`
	FlagPollInterval   int    `json:"poll_interval"`
	FlagHashKey        string
	FlagRateLimit      int
	FlagMemProfile     string
	FlagCryptoKey      string `json:"crypto_key"`
	envRunAddr         string `env:"ADDRESS"`
	envReportInterval  int    `env:"REPORT_INTERVAL"`
	envPollInterval    int    `env:"POLL_INTERVAL"`
	envHashKey         string `env:"KEY"`
	envRateLimit       int    `env:"RATE_LIMIT"`
	MemProfile         string `env:"MEM_PROFILE"`
	envCryptoKey       string `env:"CRYPTO_KEY"`
	Config             string `env:"CONFIG"`
}

// ParseFlags обрабатывает аргументы командной строки
// и сохраняет их значения в соответствующих переменных
func ParseFlags() ClientFlags {
	// для случаев, когда в переменных окружения присутствует непустое значение,
	// переопределим их, даже если они были переданы через аргументы командной строки
	cfg := readConfig()
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}
	// регистрируем переменную flagRunAddr
	// как аргумент -a со значением :8080 по умолчанию
	flag.StringVar(&cfg.FlagRunAddr, "a", "127.0.0.1:8080", "address and port to run server")
	// регистрируем переменную flagReportInterval
	// как аргумент -r со значением 10 по умолчанию
	flag.IntVar(&cfg.FlagReportInterval, "r", 10, "report interval")
	// регистрируем переменную flagPollInterval
	// как аргумент -p со значением 2 по умолчанию
	flag.IntVar(&cfg.FlagPollInterval, "p", 2, "poll interval")
	// регистрируем переменную FlagHashKey
	// как аргумент -k со значением "" по умолчанию
	flag.StringVar(&cfg.FlagHashKey, "k", "", "hash key")
	// регистрируем переменную FlagRateLimit
	// как аргумент -l со значением 1 по умолчанию
	flag.IntVar(&cfg.FlagRateLimit, "l", 1, "rate limit")
	// регистрируем переменную FlagMemProfile
	// как аргумент -mem со значением "profiles/base.pprof" по умолчанию
	flag.StringVar(&cfg.FlagMemProfile, "mem", "profiles/base.pprof", "mem profile path")
	// регистрируем переменную FlagCryptoKey
	// как аргумент -crypto-key со значением локального каталога по умолчанию
	flag.StringVar(&cfg.FlagCryptoKey, "crypto-key", "", "path to public key")
	// парсим переданные серверу аргументы в зарегистрированные переменные
	flag.Parse()
	if cfg.envRunAddr != "" {
		cfg.FlagRunAddr = cfg.envRunAddr
	} else if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		cfg.FlagRunAddr = envRunAddr
	}
	if cfg.envReportInterval != 0 {
		cfg.FlagReportInterval = cfg.envReportInterval
	}
	if cfg.envPollInterval != 0 {
		cfg.FlagPollInterval = cfg.envPollInterval
	}
	if cfg.envHashKey != "" {
		cfg.FlagHashKey = cfg.envHashKey
	} else if envHashKey := os.Getenv("KEY"); envHashKey != "" {
		cfg.FlagHashKey = envHashKey
	}
	if cfg.envRateLimit != 0 {
		cfg.FlagRateLimit = cfg.envRateLimit
	}
	if MemProfile := os.Getenv("MEM_PROFILE"); MemProfile != "" {
		cfg.FlagMemProfile = MemProfile
	}
	if cfg.envCryptoKey != "" {
		cfg.FlagCryptoKey = cfg.envCryptoKey
	} else if envCryptoKey := os.Getenv("CRYPTO_KEY"); envCryptoKey != "" {
		cfg.FlagCryptoKey = envCryptoKey
	}
	return cfg
}

func readConfig() ClientFlags {

	cfg := ClientFlags{}

	// регистрируем переменную Config
	// как аргумент -config со значением локального каталога
	flag.StringVar(&cfg.Config, "config", "", "path to config file")

	// парсим переданные серверу аргументы в зарегистрированные переменные
	flag.Parse()

	if cfg.Config == "" {
		cfg.Config, _ = os.LookupEnv("CONFIG")
	}

	if cfg.Config == "" {
		return cfg
	}

	data, err := os.ReadFile(cfg.Config)
	if err != nil {
		return cfg
	}
	reader := bytes.NewReader(data)
	if err := json.NewDecoder(reader).Decode(&cfg); err != nil {
		return cfg
	}

	return cfg
}
