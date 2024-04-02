package main

import (
	"fmt"
	"github.com/smartfor/metrics/cmd/server/config"
	"github.com/smartfor/metrics/cmd/server/handlers"
	"github.com/smartfor/metrics/cmd/server/storage"
	"net/http"
)

func main() {
	cfg := config.GetConfig()
	fmt.Println("Server config :: \n", cfg)

	if err := run(cfg); err != nil {
		panic(err)
	}
}

func run(cfg config.Config) error {
	metricStorage := storage.NewMemStorage()
	r := handlers.Router(metricStorage)

	fmt.Printf("Server started on '%s'...\n", cfg.Addr)
	return http.ListenAndServe(cfg.Addr, r)
}
