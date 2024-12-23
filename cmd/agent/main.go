// Агент для сбора метрик и отправки их на сервер для хранения
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/smartfor/metrics/internal"
	"github.com/smartfor/metrics/internal/build"
	"github.com/smartfor/metrics/internal/config"
)

func main() {
	build.PrintGlobalVars()

	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %s\n", err)
	}

	fmt.Printf("Agent config :: \n %v\n", cfg)

	var privateKey []byte
	if cfg.CryptoKey != "" {
		if cfg.CryptoKey != "" {
			pk, err := os.ReadFile(cfg.CryptoKey)
			if err != nil {
				log.Fatalf("Public key not found")
			}
			privateKey = pk
		}
	}

	s := internal.NewService(cfg, privateKey)

	waitShutdown := make(chan struct{})
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	go func() {
		<-done
		log.Println("Shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := s.Shutdown(ctx); err != nil {
			log.Fatalf("Agent Shutdown Failed: %v", err)
		}

		close(waitShutdown)
	}()

	if err := s.Run(context.Background()); err != nil && !errors.Is(err, internal.ErrAgentClosed) {
		log.Fatalf("Agent Run failed: %v", err)
	}

	<-waitShutdown
	fmt.Println("Agent Shutdown gracefully!")
}
