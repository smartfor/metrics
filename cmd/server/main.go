package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/smartfor/metrics/internal/server/config"
	"github.com/smartfor/metrics/internal/server/handlers"
	"github.com/smartfor/metrics/internal/server/storage"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg, err := config.GetConfig()
	if err != nil {
		fmt.Println("Error loading configuration:", err)
		os.Exit(1)
	}

	log.Println("Server config:", cfg)

	metricStorage := storage.NewMemStorage()
	router := handlers.Router(metricStorage)

	server := &http.Server{
		Addr:    cfg.Addr,
		Handler: router,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-done
		log.Println("Shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("Server Shutdown Failed:%+v", err)
		}
		log.Println("Server gracefully stopped.")
	}()

	log.Printf("Server is ready to handle requests at %s", cfg.Addr)
	if err := server.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("could not listen on %s: %v", cfg.Addr, err)
	}
}
