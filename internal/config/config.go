package config

import "time"

type Config struct {
	PollInterval    time.Duration
	ReportInterval  time.Duration
	ResponseTimeout time.Duration
	UpdateURL       string
}

var DefaultConfig = Config{
	PollInterval:    2 * time.Second,
	ReportInterval:  10 * time.Second,
	ResponseTimeout: 3 * time.Second,
	UpdateURL:       "http://localhost:8080/update",
}
