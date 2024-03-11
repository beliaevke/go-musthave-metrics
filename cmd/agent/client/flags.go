package client

import (
	"flag"
)

var flagRunAddr string
var flagReportInterval int
var flagPollInterval int

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
}
