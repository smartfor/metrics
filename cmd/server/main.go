package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/smartfor/metrics/internal/core"
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

	backupStorage, err := storage.NewFileStorage(cfg.FileStoragePath)
	if err != nil {
		logger.Log.Fatal("Error creating backup storage: ", zap.Error(err))
	}

	memStorage, err := storage.NewMemStorage(backupStorage, cfg.Restore, cfg.StoreInterval == 0)
	if err != nil {
		logger.Log.Fatal("Error creating metric storage: ", zap.Error(err))
	}

	router := handlers.Router(memStorage, logger.Log)
	server := &http.Server{
		Addr:    cfg.Addr,
		Handler: router,
	}

	if cfg.StoreInterval > 0 {
		go func(
			storage core.Storage,
			backup core.Storage,
			interval time.Duration,
		) {
			time.Sleep(interval)
			ticker := time.NewTicker(interval)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					if err := core.Sync(storage, backup); err != nil {
						fmt.Println(err)
						logger.Log.Error("Error sync metrics: ", zap.Error(err))
					}
				}
			}
		}(memStorage, backupStorage, cfg.StoreInterval)
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
		if err := core.Sync(memStorage, backupStorage); err != nil {
			logger.Log.Fatal("Memstorage Backup Failed: ", zap.Error(err))
		}
		if err := memStorage.Close(); err != nil {
			logger.Log.Fatal("Memstorage Close Failed: ", zap.Error(err))
		}
	}
}
