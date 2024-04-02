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
	envStoreInterval    int    `env:"STORE_INTERVAL"`
	envFileStoragePath  string `env:"FILE_STORAGE_PATH"`
	envRestore          bool   `env:"RESTORE"`
}

// parseFlags обрабатывает аргументы командной строки
// и сохраняет их значения в соответствующих переменных
func ParseFlags() ServerFlags {
	// для случаев, когда в переменных окружения присутствует непустое значение,
	// переопределим их, даже если они были переданы через аргументы командной строки
	cfg := new(ServerFlags)
	err := env.Parse(cfg)
	if err != nil {
		log.Fatal(err)
	}
	// регистрируем переменную flagRunAddr
	// как аргумент -a со значением :8080 по умолчанию
	flag.StringVar(&cfg.FlagRunAddr, "a", "localhost:8080", "address and port to run server")
	// регистрируем переменную FlagStoreInterval
	// интервал времени в секундах, по истечении которого текущие показания сервера сохраняются на диск
	// (по умолчанию 300 секунд, значение 0 делает запись синхронной)
	flag.IntVar(&cfg.FlagStoreInterval, "i", 300, "report interval")
	// регистрируем переменную FlagFileStoragePath
	// полное имя файла, куда сохраняются текущие значения (по умолчанию /tmp/metrics-db.json, пустое значение отключает функцию записи на диск)
	flag.StringVar(&cfg.FlagFileStoragePath, "f", "/tmp/metrics-db.json", "report interval")
	// регистрируем переменную FlagRestore
	// булево значение (true/false), определяющее, загружать или нет ранее сохранённые значения из указанного файла при старте сервера (по умолчанию true).
	flag.BoolVar(&cfg.FlagRestore, "r", true, "report interval")
	// парсим переданные серверу аргументы в зарегистрированные переменные
	flag.Parse()

	// для случаев, когда в переменной окружения ADDRESS присутствует непустое значение,
	// переопределим адрес запуска сервера,
	// даже если он был передан через аргумент командной строки
	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		cfg.FlagRunAddr = envRunAddr
	}
	if cfg.envStoreInterval != 0 {
		cfg.FlagStoreInterval = cfg.envStoreInterval
	}
	if cfg.envFileStoragePath != "" {
		cfg.FlagFileStoragePath = cfg.envFileStoragePath
	}
	if _, isSet := os.LookupEnv("RESTORE"); !isSet {
		cfg.FlagRestore = cfg.envRestore
	}
	return *cfg
}
