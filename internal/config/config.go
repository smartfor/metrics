package config

import "time"

type Config struct {
	PollInterval   time.Duration
	ReportInterval time.Duration
	UpdateURL      string
}

var DefaultConfig = Config{
	PollInterval:   2 * time.Second,
	ReportInterval: 10 * time.Second,
	UpdateURL:      "http://localhost:8080/update",
}
