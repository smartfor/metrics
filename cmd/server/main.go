// Сервер хранения метрик
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/smartfor/metrics/api/metricapi"
	"github.com/smartfor/metrics/internal/build"
	"github.com/smartfor/metrics/internal/core"
	crypto_codec "github.com/smartfor/metrics/internal/crypto_codec"
	"github.com/smartfor/metrics/internal/logger"
	"github.com/smartfor/metrics/internal/server/config"
	"github.com/smartfor/metrics/internal/server/handlers"
	"github.com/smartfor/metrics/internal/server/storage"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
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

	if postgresStorage != nil && cfg.StoreIntervalDuration > 0 {
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

	var g errgroup.Group

	// ---------------------------- HTTP TRANSPORT ----------------------------- //
	g.Go(func() error {
		var (
			router chi.Router
			store  core.Storage
		)

		if postgresStorage != nil {
			store = postgresStorage
		} else {
			store = memStorage
		}

		router = handlers.Router(store, zlog, cfg.Secret, privateKey, cfg.TrustedSubnet)

		server := &http.Server{
			Addr:              cfg.Addr,
			ReadHeaderTimeout: 10 * time.Second,
			Handler:           router,
		}

		errsCh := make(chan error, 1)
		go func() {
			defer close(errsCh)

			zlog.Info("Http Server is ready to handle requests at ", zap.String("address", cfg.Addr))
			if err := server.ListenAndServe(); err == nil || errors.Is(err, http.ErrServerClosed) {
				return
			}

			errsCh <- fmt.Errorf("HTTP server ListenAndServe failed: %w", err)
		}()

		go func() {
			signalCh := make(chan os.Signal, 1)
			defer close(signalCh)

			signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
			<-signalCh
			zlog.Info("Shutting down server...")

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := server.Shutdown(ctx); err != nil {
				errsCh <- fmt.Errorf("server shutdown failed: %w", err)
			}
		}()

		return <-errsCh
	})

	// ---------------------------------- GRPC TRANSPORT ------------------------------------ //

	g.Go(func() error {
		listener, err := net.Listen("tcp", cfg.GrpcAddr)
		if err != nil {
			return err
		}

		server := grpc.NewServer(
			grpc.RPCCompressor(grpc.NewGZIPCompressor()),
			grpc.RPCDecompressor(grpc.NewGZIPDecompressor()),
			grpc.ChainUnaryInterceptor(
				handlers.MakeGrpcAuthInterceptor(cfg),
			),
			// Нужна помощь!
			// я не понимаю как быть? в кодеке я не имею доступа к открытому ключу из метаданных
			//  а в интерцепторе я имею доступ к метаданным но уже поздно так как он вызывается после кодека,
			//  и уже будет ошибка так как данные зашифрованы!!!!!!
			grpc.ForceServerCodec(crypto_codec.MakeCryptoCodec()),
		)

		metricapi.RegisterMetricsServer(
			server,
			handlers.NewGRPCServer(memStorage, zlog, cfg.Secret, privateKey, cfg.TrustedSubnet),
		)

		reflection.Register(server)

		errsCh := make(chan error, 1)

		go func() {
			signalCh := make(chan os.Signal, 1)
			defer close(signalCh)

			signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
			<-signalCh

			zlog.Info("Shutting down server...")
			server.GracefulStop()
		}()
		zlog.Info("Запуск gRPC сервера", zap.String("address", cfg.GrpcAddr))

		go func() {
			defer close(errsCh)

			zlog.Info("Grpc Server is ready to handle requests at ", zap.String("address", cfg.GrpcAddr))
			if err := server.Serve(listener); err == nil || errors.Is(err, grpc.ErrServerStopped) {
				return
			}

			errsCh <- fmt.Errorf("GRPC server Serve failed: %w", err)
		}()

		return <-errsCh
	})

	g.Wait()

	if err := core.Sync(context.Background(), memStorage, backupStorage); err != nil {
		zlog.Fatal("memstorage backup failed:", zap.Error(err))
	}

	if err := memStorage.Close(); err != nil {
		zlog.Fatal("memstorage close failed:", zap.Error(err))
	}

	zlog.Info("Server stopped")
}
