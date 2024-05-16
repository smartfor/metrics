package config

import (
	"errors"
	"flag"
	"fmt"
	"github.com/smartfor/metrics/internal/utils"
	"time"
)

var (
	ErrInvalidAddress   = errors.New("invalid address")
	ErrUnknownArguments = errors.New(fmt.Sprint("unknown flags:", flag.Args()))
)

type Config struct {
	Addr            string
	LogLevel        string
	StoreInterval   time.Duration
	FileStoragePath string
	Restore         bool
	DatabaseDSN     string
	Secret          string
}

func GetConfig() (*Config, error) {
	var (
		addr            = flag.String("a", ":8080", "address and port to run server")
		logLevel        = flag.String("l", "info", "log level")
		fileStoragePath = flag.String("f", "/tmp/metrics-db.json", "file storage path")
		storeInterval   = flag.Int("i", 300, "metrics store interval")
		restore         = flag.Bool("r", true, "restore metrics when server starts")
		dbDsn           = flag.String("d", "", "database DSN")
		key             = flag.String("k", "", "very very very secret key")
	)
	flag.Parse()

	if err := utils.ValidateAddress(*addr); err != nil {
		return nil, ErrInvalidAddress
	}

	if len(flag.Args()) > 0 {
		return nil, ErrUnknownArguments
	}

	utils.TryTakeStringFromEnv("ADDRESS", addr)
	utils.TryTakeStringFromEnv("LOG_LEVEL", logLevel)
	utils.TryTakeStringFromEnv("FILE_STORAGE_PATH", fileStoragePath)
	utils.TryTakeIntFromEnv("STORE_INTERVAL", storeInterval)
	utils.TryGetBoolFromEnv("RESTORE", restore)
	utils.TryTakeStringFromEnv("DATABASE_DSN", dbDsn)
	utils.TryTakeStringFromEnv("KEY", key)

	return &Config{
		Addr:            *addr,
		LogLevel:        *logLevel,
		StoreInterval:   time.Second * time.Duration(*storeInterval),
		FileStoragePath: *fileStoragePath,
		Restore:         *restore,
		DatabaseDSN:     *dbDsn,
		Secret:          *key,
	}, nil
}
