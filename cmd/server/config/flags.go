package config

import (
	"flag"
	"log"
	"os"

	"github.com/caarlos0/env"
)

type ServerFlags struct {
	FlagRunAddr         string
	FlagStoreInterval   int
	FlagFileStoragePath string
	FlagRestore         bool
	FlagDatabaseDSN     string
	FlagHashKey         string
	EnvStoreInterval    int    `env:"STORE_INTERVAL"`
	FileStoragePath     string `env:"FILE_STORAGE_PATH"`
	EnvRestore          bool   `env:"RESTORE"`
	DatabaseDSN         string `env:"DATABASE_DSN"`
	envHashKey          string `env:"KEY"`
}

// parseFlags обрабатывает аргументы командной строки
// и сохраняет их значения в соответствующих переменных
func ParseFlags() ServerFlags {
	// для случаев, когда в переменных окружения присутствует непустое значение,
	// переопределим их, даже если они были переданы через аргументы командной строки
	cfg := new(ServerFlags)
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
	flag.StringVar(&cfg.FlagDatabaseDSN, "d", "", "Database DSN")
	// регистрируем переменную FlagHashKey
	// как аргумент -k со значением "" по умолчанию
	flag.StringVar(&cfg.FlagHashKey, "k", "", "hash key")
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
	if cfg.envHashKey != "" {
		cfg.FlagHashKey = cfg.envHashKey
	} else if envHashKey := os.Getenv("KEY"); envHashKey != "" {
		cfg.FlagHashKey = envHashKey
	}
	return *cfg
}
