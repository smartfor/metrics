// Агент для сбора метрик и отправки их на сервер для хранения
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/smartfor/metrics/internal"
	"github.com/smartfor/metrics/internal/build"
	"github.com/smartfor/metrics/internal/config"
)

func main() {
	build.PrintGlobalVars()

	cfg, err := config.GetConfig()
	if err != nil {
		fmt.Printf("Error loading configuration: %s\n", err)
		return
	}

	fmt.Printf("Agent config :: \n %v\n", cfg)

	var privateKey []byte
	if cfg.CryptoKey != "" {
		if cfg.CryptoKey != "" {
			fmt.Println("Crypto key is set")
			pk, err := os.ReadFile(cfg.CryptoKey)
			if err != nil {
				fmt.Println("Public key not found")
				return
			}
			privateKey = pk
			fmt.Println(string(privateKey))
		}
	}

	s := internal.NewService(cfg, privateKey)
	s.Run(context.Background())
}
