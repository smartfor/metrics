package main

import (
	"fmt"
	"github.com/smartfor/metrics/internal"
	"github.com/smartfor/metrics/internal/config"
)

func main() {
	cfg := config.GetConfig()
	fmt.Println("Agent config :: \n", cfg)

	s := internal.NewService(&cfg)
	s.Run()
}
