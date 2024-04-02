package main

import (
	"github.com/smartfor/metrics/internal"
	"github.com/smartfor/metrics/internal/config"
)

func main() {
	cfg := config.ParseConfig()
	s := internal.NewService(&cfg)
	s.Run()
}
