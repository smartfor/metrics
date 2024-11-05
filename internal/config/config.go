// Package config содержит структуру конфигурации и функцию для ее получения.
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/smartfor/metrics/internal/cfgutils"
)

const (
	HTTPProto     = "http://"
	HTTPSProto    = "https://"
	HTTPTransport = "http"
	GRPCTransport = "grpc"
)

type Config struct {
	HostEndpoint            string `json:"address"`
	Secret                  string `json:"secret"`
	CryptoKey               string `json:"crypto_key"`
	RateLimit               int    `json:"rate_limit"`
	PollInterval            string `json:"poll_interval"`
	ReportInterval          string `json:"report_interval"`
	ResponseTimeout         string `json:"response_timeout"`
	Transport               string `json:"transport"`
	PollIntervalDuration    time.Duration
	ReportIntervalDuration  time.Duration
	ResponseTimeoutDuration time.Duration
}

func GetConfig() (*Config, error) {
	config := &Config{
		HostEndpoint:    "localhost:8080",
		PollInterval:    "2s",
		ReportInterval:  "10s",
		ResponseTimeout: "3s",
		RateLimit:       1,
		Transport:       HTTPTransport,
	}

	// resolve config path
	configFile := flag.String("config", "", "path to config file")
	flag.StringVar(configFile, "c", "", "path to config file (shorthand)")
	flag.Parse()
	cfgutils.TryTakeStringFromEnv("CONFIG", configFile)
	// Load from JSON config file if specified
	if *configFile != "" {
		file, err := os.Open(*configFile)
		if err != nil {
			return nil, fmt.Errorf("error opening config file: %w", err)
		}
		defer file.Close()
		decoder := json.NewDecoder(file)
		if err := decoder.Decode(config); err != nil {
			return nil, fmt.Errorf("error decoding JSON config: %w", err)
		}
	}

	if len(flag.Args()) > 0 {
		return nil, fmt.Errorf("unknown flags: %v", flag.Args())
	}

	cfgutils.ParseString("a", "ADDRESS", "host endpoint", &config.HostEndpoint)
	cfgutils.ParseString("k", "KEY", "secret key", &config.Secret)
	cfgutils.ParseInt("l", "RATE_LIMIT", "rate limit", &config.RateLimit)
	cfgutils.ParseString("crypto-key", "CRYPTO_KEY", "crypto key", &config.CryptoKey)

	cfgutils.ParseString("p", "POLL_INTERVAL", "poll interval", &config.PollInterval)
	val, err := time.ParseDuration(config.PollInterval)
	if err != nil {
		return nil, fmt.Errorf("error parsing poll interval: %w", err)
	}
	config.PollIntervalDuration = val

	cfgutils.ParseString("r", "REPORT_INTERVAL", "report interval", &config.ReportInterval)
	val, err = time.ParseDuration(config.ReportInterval)
	if err != nil {
		return nil, fmt.Errorf("error parsing report interval: %w", err)
	}
	config.ReportIntervalDuration = val

	cfgutils.ParseString("t", "RESPONSE_TIMEOUT", "response timeout", &config.ResponseTimeout)
	val, err = time.ParseDuration(config.ResponseTimeout)
	if err != nil {
		return nil, fmt.Errorf("error parsing response timeout: %w", err)
	}
	config.ResponseTimeoutDuration = val

	if !strings.HasPrefix(config.HostEndpoint, HTTPProto) &&
		!strings.HasPrefix(config.HostEndpoint, HTTPSProto) &&
		config.Transport == HTTPTransport {
		config.HostEndpoint = HTTPProto + config.HostEndpoint
	}

	return config, nil
}
