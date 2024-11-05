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

	"github.com/smartfor/metrics/internal/agent"
	"github.com/smartfor/metrics/internal/build"
	"github.com/smartfor/metrics/internal/config"
	"github.com/smartfor/metrics/internal/metric_sender"
)

func main() {
	build.PrintGlobalVars()

	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %s\n", err)
	}

	fmt.Printf("Agent config :: \n %v\n", cfg)

	var publicKey []byte
	if cfg.CryptoKey != "" {
		if cfg.CryptoKey != "" {
			pk, err := os.ReadFile(cfg.CryptoKey)
			if err != nil {
				log.Fatalf("Public key not found")
			}
			publicKey = pk
		}
	}

	var sender metric_sender.MetricSender
	if cfg.Transport == config.HttpTransport {
		sender, err = metric_sender.NewHttpMetricSender(cfg, publicKey)
		if err != nil {
			log.Fatalf("Error creating metric sender: %v", err)
		}
	} else if cfg.Transport == config.GrpcTransport {
		sender, err = metric_sender.NewGrpcMetricSender(cfg, publicKey)
		if err != nil {
			log.Fatalf("Error creating metric sender: %v", err)
		}
	} else {
		log.Fatalf("Unknown transport type: %v", cfg.Transport)
	}

	s := agent.NewService(cfg, sender, publicKey)

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

	if err := s.Run(context.Background()); err != nil && !errors.Is(err, agent.ErrAgentClosed) {
		log.Fatalf("Agent Run failed: %v", err)
	}

	<-waitShutdown
	fmt.Println("Agent Shutdown gracefully!")
}
