package config

import (
	"errors"
	"flag"
	"fmt"
	"github.com/smartfor/metrics/internal/utils"
	"os"
)

var (
	ErrInvalidAddress   = errors.New("invalid address")
	ErrUnknownArguments = errors.New(fmt.Sprint("unknown flags:", flag.Args()))
)

type Config struct {
	Addr string
}

func GetConfig() (*Config, error) {
	addr := flag.String("a", ":8080", "address and port to run server")
	flag.Parse()

	if err := utils.ValidateAddress(*addr); err != nil {
		return nil, ErrInvalidAddress
	}

	if len(flag.Args()) > 0 {
		return nil, ErrUnknownArguments
	}

	if a := os.Getenv("ADDRESS"); a != "" {
		*addr = a
	}

	return &Config{Addr: *addr}, nil
}
