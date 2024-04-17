package main

import (
	"flag"
	"fmt"
	"github.com/smartfor/metrics/internal/server/config"
	"github.com/smartfor/metrics/internal/server/handlers"
	"github.com/smartfor/metrics/internal/server/storage"
	"net/http"
	"os"
)

func main() {
	cfg, err := config.GetConfig()
	if err != nil {
		flag.PrintDefaults()
		os.Exit(1)
	}

	fmt.Println("Server config :: \n", cfg)

	if err := run(cfg); err != nil {
		panic(err)
	}
}

func run(cfg *config.Config) error {
	metricStorage := storage.NewMemStorage()
	r := handlers.Router(metricStorage)

	fmt.Printf("Server started on '%s'...\n", cfg.Addr)
	return http.ListenAndServe(cfg.Addr, r)
}
