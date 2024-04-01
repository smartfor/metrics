package main

import (
	"github.com/smartfor/metrics/cmd/server/handlers"
	"github.com/smartfor/metrics/cmd/server/storage"
	"net/http"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	metricStorage := storage.NewMemStorage()

	mux := http.NewServeMux()
	mux.HandleFunc("/update/", handlers.MakeUpdateHandler(metricStorage))

	return http.ListenAndServe(`:8080`, mux)
}
