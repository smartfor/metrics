package main

import (
	"flag"
	"fmt"
	"github.com/smartfor/metrics/cmd/server/handlers"
	"github.com/smartfor/metrics/cmd/server/storage"
	"net/http"
	"os"
)

var flagRunAddr string

func main() {
	parseFlags()

	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	metricStorage := storage.NewMemStorage()
	r := handlers.Router(metricStorage)
	return http.ListenAndServe(flagRunAddr, r)
}

func parseFlags() {
	flag.StringVar(&flagRunAddr, "a", ":8080", "address and port to run server")
	flag.Parse()

	if len(flag.Args()) > 0 {
		fmt.Println("Error: unknown flags:", flag.Args())
		flag.PrintDefaults()
		os.Exit(1)
	}
}
