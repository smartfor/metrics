package config

import (
	"errors"
	"flag"
	"fmt"
	"github.com/smartfor/metrics/internal/utils"
	"os"
)

var (
	InvalidAddressError   = errors.New("invalid address")
	UnknownArgumentsError = errors.New(fmt.Sprint("unknown flags:", flag.Args()))
)

type Config struct {
	Addr string
}

func GetConfig() (*Config, error) {
	addr := flag.String("a", ":8080", "address and port to run server")
	flag.Parse()

	if err := utils.ValidateAddress(*addr); err != nil {
		return nil, InvalidAddressError
	}

	if len(flag.Args()) > 0 {
		return nil, UnknownArgumentsError
	}

	if a := os.Getenv("ADDRESS"); a != "" {
		*addr = a
	}

	return &Config{Addr: *addr}, nil
}
