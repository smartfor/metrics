// Сервер хранения метрик
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/smartfor/metrics/internal/build"
	"github.com/smartfor/metrics/internal/core"
	"github.com/smartfor/metrics/internal/logger"
	"github.com/smartfor/metrics/internal/server/config"
	"github.com/smartfor/metrics/internal/server/handlers"
	"github.com/smartfor/metrics/internal/server/storage"
	"go.uber.org/zap"
)

func main() {
	build.PrintGlobalVars()

	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %s", err)
	}

	zlog, err := logger.MakeLogger(cfg.LogLevel)
	if err != nil {
		log.Fatalf("Error initialize logger: %s", err)
	}

	zlog.Sugar().Infof("Server config: %+v", cfg)

	backupStorage, err := storage.NewFileStorage(cfg.FileStoragePath)
	if err != nil {
		zlog.Fatal("Error creating backup storage: ", zap.Error(err))
	}

	memStorage, err := storage.NewMemStorage(backupStorage, cfg.Restore, cfg.StoreIntervalDuration == 0)
	if err != nil {
		zlog.Fatal("Error creating metric storage: ", zap.Error(err))
	}

	var postgresStorage *storage.PostgresStorage
	if cfg.DatabaseDSN != "" {
		postgresStorage, err = storage.NewPostgresStorage(context.Background(), cfg.DatabaseDSN)
		if err != nil {
			zlog.Fatal("Error creatingPostgresStorage: ", zap.Error(err))
		}
	}

	var privateKey []byte
	if cfg.CryptoKey != "" {
		zlog.Info("Crypto key is set")
		privateKey, err = os.ReadFile(cfg.CryptoKey)
		if err != nil {
			zlog.Fatal("Error reading crypto key file: ", zap.Error(err))
		}
	}

	var router chi.Router
	if postgresStorage != nil {
		router = handlers.Router(postgresStorage, zlog, cfg.Secret, privateKey)
	} else {
		router = handlers.Router(memStorage, zlog, cfg.Secret, privateKey)
		if cfg.StoreIntervalDuration > 0 {
			go func(
				storage core.Storage,
				backup core.Storage,
				interval time.Duration,
			) {
				time.Sleep(interval)
				ticker := time.NewTicker(interval)
				defer ticker.Stop()

				for range ticker.C {
					if err := core.Sync(context.Background(), storage, backup); err != nil {
						fmt.Println(err)
						zlog.Error("Error sync metrics: ", zap.Error(err))
					}
				}
			}(memStorage, backupStorage, cfg.StoreIntervalDuration)
		}
	}
	server := &http.Server{
		Addr:              cfg.Addr,
		ReadHeaderTimeout: 10 * time.Second,
		Handler:           router,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-done
		zlog.Info("Shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			zlog.Fatal("Server Shutdown Failed: ", zap.Error(err))
		}

		zlog.Info("Server gracefully stopped.")
	}()

	log.Printf("Server is ready to handle requests at %s", cfg.Addr)
	if err := server.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
		if err := core.Sync(context.Background(), memStorage, backupStorage); err != nil {
			zlog.Fatal("Memstorage Backup Failed: ", zap.Error(err))
		}
		if err := memStorage.Close(); err != nil {
			zlog.Fatal("Memstorage Close Failed: ", zap.Error(err))
		}
	}
}
