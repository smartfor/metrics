package main

import (
	"context"
	"fmt"

	"github.com/smartfor/metrics/internal"
	"github.com/smartfor/metrics/internal/config"
)

func main() {
	cfg := config.GetConfig()
	fmt.Printf("Agent config :: \n %v\n", cfg)

	s := internal.NewService(&cfg)
	s.Run(context.Background())
}
