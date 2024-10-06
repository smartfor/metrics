// Агент для сбора метрик и отправки их на сервер для хранения
package main

import (
	"context"
	"fmt"

	"github.com/smartfor/metrics/internal"
	"github.com/smartfor/metrics/internal/config"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	printGlobalVars()

	cfg := config.GetConfig()
	fmt.Printf("Agent config :: \n %v\n", cfg)

	s := internal.NewService(&cfg)
	s.Run(context.Background())
}

func printGlobalVars() {
	const NA = "N/A"

	if buildVersion == "" {
		buildVersion = NA
	}

	if buildDate == "" {
		buildDate = NA
	}

	if buildCommit == "" {
		buildCommit = NA
	}

	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}
