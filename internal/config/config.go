package config

import (
	"flag"
	"fmt"
	"github.com/smartfor/metrics/internal/utils"
	"os"
	"strings"
	"time"
)

const (
	HTTPProto  = "http://"
	HTTPSProto = "https://"
)

type Config struct {
	PollInterval    time.Duration
	ReportInterval  time.Duration
	ResponseTimeout time.Duration
	HostEndpoint    string
}

func GetConfig() Config {
	pollInterval := flag.Int("p", 2, "Poll Interval")
	reportInterval := flag.Int("r", 10, "Report Interval")
	responseTimeout := flag.Int("t", 3, "Response Timeout")
	hostEndpoint := flag.String("a", "http://localhost:8080", "Host Endpoint")

	flag.Parse()

	if len(flag.Args()) > 0 {
		fmt.Println("Error: unknown flags:", flag.Args())
		flag.PrintDefaults()
		os.Exit(1)
	}

	utils.TryTakeIntFromEnv("POLL_INTERVAL", pollInterval)
	utils.TryTakeIntFromEnv("REPORT_INTERVAL", reportInterval)
	utils.TryTakeIntFromEnv("RESPONSE_TIMEOUT", responseTimeout)

	if a := os.Getenv("ADDRESS"); a != "" {
		*hostEndpoint = a
	}

	if !strings.HasPrefix(*hostEndpoint, HTTPProto) && !strings.HasPrefix(*hostEndpoint, HTTPSProto) {
		*hostEndpoint = HTTPProto + *hostEndpoint
	}

	return Config{
		PollInterval:    time.Duration(*pollInterval) * time.Second,
		ReportInterval:  time.Duration(*reportInterval) * time.Second,
		ResponseTimeout: time.Duration(*responseTimeout) * time.Second,
		HostEndpoint:    *hostEndpoint,
	}
}
