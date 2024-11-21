// Package config предназначен для методов конфигурации
package config

import (
	"bytes"
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/caarlos0/env"
)

type ServerFlags struct {
	FlagRunAddr         string `json:"address"`
	FlagStoreInterval   int    `json:"store_interval"`
	FlagFileStoragePath string `json:"store_file"`
	FlagRestore         bool   `json:"restore"`
	FlagDatabaseDSN     string `json:"database_dsn"`
	FlagHashKey         string
	FlagMemProfile      string
	FlagCryptoKey       string `json:"crypto_key"`
	EnvStoreInterval    int    `env:"STORE_INTERVAL"`
	FileStoragePath     string `env:"FILE_STORAGE_PATH"`
	EnvRestore          bool   `env:"RESTORE"`
	DatabaseDSN         string `env:"DATABASE_DSN"`
	EnvHashKey          string `env:"KEY"`
	MemProfile          string `env:"MEM_PROFILE"`
	envCryptoKey        string `env:"CRYPTO_KEY"`
	Config              string `env:"CONFIGSRV"`
}

// ParseFlags обрабатывает аргументы командной строки
// и сохраняет их значения в соответствующих переменных
func ParseFlags() ServerFlags {
	// для случаев, когда в переменных окружения присутствует непустое значение,
	// переопределим их, даже если они были переданы через аргументы командной строки
	cfg := readConfig()
	if err := env.Parse(cfg); err != nil {
		log.Fatal(err)
	}
	// регистрируем переменную flagRunAddr
	// как аргумент -a со значением :8080 по умолчанию
	flag.StringVar(&cfg.FlagRunAddr, "a", "localhost:8080", "address and port to run server")
	// регистрируем переменную FlagStoreInterval
	// интервал времени в секундах, по истечении которого текущие показания сервера сохраняются на диск
	// (по умолчанию 300 секунд, значение 0 делает запись синхронной)
	flag.IntVar(&cfg.FlagStoreInterval, "i", 300, "store interval")
	// регистрируем переменную FlagFileStoragePath
	// полное имя файла, куда сохраняются текущие значения (по умолчанию /tmp/metrics-db.json, пустое значение отключает функцию записи на диск)
	flag.StringVar(&cfg.FlagFileStoragePath, "f", "/tmp/metrics-db.json", "file storage path")
	// регистрируем переменную FlagRestore
	// булево значение (true/false), определяющее, загружать или нет ранее сохранённые значения из указанного файла при старте сервера (по умолчанию true).
	flag.BoolVar(&cfg.FlagRestore, "r", true, "flag restore")
	// Строка с адресом подключения к БД должна получаться из переменной окружения DATABASE_DSN или флага командной строки -d.
	flag.StringVar(&cfg.FlagDatabaseDSN, "d", "postgres://postgres:pos111@localhost:5432/postgres?sslmode=disable", "Database DSN")
	// регистрируем переменную FlagHashKey
	// как аргумент -k со значением "" по умолчанию
	flag.StringVar(&cfg.FlagHashKey, "k", "", "hash key")
	// регистрируем переменную FlagMemProfile
	// как аргумент -mem со значением "profiles/base.pprof" по умолчанию
	flag.StringVar(&cfg.FlagMemProfile, "mem", "profiles/base.pprof", "mem profile path")
	// регистрируем переменную FlagCryptoKey
	// как аргумент -crypto-key со значением локального каталога по умолчанию
	flag.StringVar(&cfg.FlagCryptoKey, "crypto-key", "D:/_learning/YaP_workspace/go-musthave-metrics/cmd/cryptokeys/key", "path to private key")
	// парсим переданные серверу аргументы в зарегистрированные переменные
	flag.Parse()

	// для случаев, когда в переменной окружения ADDRESS присутствует непустое значение,
	// переопределим адрес запуска сервера,
	// даже если он был передан через аргумент командной строки
	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		cfg.FlagRunAddr = envRunAddr
	}
	if cfg.EnvStoreInterval != 0 {
		cfg.FlagStoreInterval = cfg.EnvStoreInterval
	}
	if cfg.FileStoragePath != "" {
		cfg.FlagFileStoragePath = cfg.FileStoragePath
	}
	if cfg.EnvRestore {
		cfg.FlagRestore = cfg.EnvRestore
	}
	if cfg.DatabaseDSN != "" {
		cfg.FlagDatabaseDSN = cfg.DatabaseDSN
	}
	if cfg.EnvHashKey != "" {
		cfg.FlagHashKey = cfg.EnvHashKey
	} else if EnvHashKey := os.Getenv("KEY"); EnvHashKey != "" {
		cfg.FlagHashKey = EnvHashKey
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

func readConfig() ServerFlags {

	cfg := ServerFlags{}

	// регистрируем переменную Config
	// как аргумент -config со значением локального каталога
	flag.StringVar(&cfg.Config, "config", "", "path to config file")

	// парсим переданные серверу аргументы в зарегистрированные переменные
	flag.Parse()

	if cfg.Config == "" {
		cfg.Config, _ = os.LookupEnv("CONFIGSRV")
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
