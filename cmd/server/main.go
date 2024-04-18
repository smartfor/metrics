package main

import (
	"context"
	"errors"
	"github.com/smartfor/metrics/internal/logger"
	"github.com/smartfor/metrics/internal/server/config"
	"github.com/smartfor/metrics/internal/server/handlers"
	"github.com/smartfor/metrics/internal/server/storage"
	"go.uber.org/zap"
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
		log.Fatalf("Error loading configuration: %s", err)
	}

	if err := logger.Initialize(cfg.LogLevel); err != nil {
		log.Fatalf("Error initialize logger: %s", err)
	}

	logger.Log.Sugar().Infof("Server config: %+v", cfg)

	metricStorage := storage.NewMemStorage()
	router := handlers.Router(metricStorage, logger.Log)

	server := &http.Server{
		Addr:    cfg.Addr,
		Handler: router,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-done
		logger.Log.Info("Shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			logger.Log.Fatal("Server Shutdown Failed: ", zap.Error(err))
		}

		logger.Log.Info("Server gracefully stopped.")
	}()

	log.Printf("Server is ready to handle requests at %s", cfg.Addr)
	if err := server.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
		logger.Log.Fatal("Error not ")
	}
}
