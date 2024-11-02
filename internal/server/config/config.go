// Package config Модуль config отвечает за определение конфигурации сервера
package config

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/smartfor/metrics/internal/cfgutils"
	"github.com/smartfor/metrics/internal/utils"
)

var (
	// ErrInvalidAddress Ошибка при некорректном формате адреса сервера передаваемом при зауске сервера
	ErrInvalidAddress = errors.New("invalid address")

	// ErrUnknownArguments Ошибка при передаче несуществующиего параметра конфигурации
	ErrUnknownArguments = errors.New(fmt.Sprint("unknown flags:", flag.Args()))
)

// Config Конфигурация сервера
type Config struct {
	// Addr адрес сервера
	Addr string `json:"address"`
	// LogLevel уровень логирования
	LogLevel string `json:"log_level"`
	// FileStoragePath путь к файловому хранилищу метрик
	FileStoragePath string `json:"file_storage_path"`
	// DatabaseDSN строка подключения к базе данных хранения метрик
	DatabaseDSN string `json:"database_dsn"`
	// Secret секретный код для создания и идентификации ключа аутентификации клиентов
	Secret string `json:"secret"`
	// StoreInterval  временной интервал (сек), через который сервер сохраняет состояние метрик в постоянное хранилище,
	StoreInterval string `json:"store_interval"` // as string 1s, 1m, 1h
	// Restore флаг включающий воостановление метрик в память после старта сервера
	Restore bool `json:"restore"`
	// CryptoKey путь к ключу для шифрования данных
	CryptoKey string `json:"crypto_key"`
	// StoreIntervalDuration - StoreInterval as time.Duration
	StoreIntervalDuration time.Duration
}

// GetConfig Функция для получения конфигурации сервера.
// Если параметры не найдены в переменных окружения то берутся значения из флагов либо значения по умолчанию
func GetConfig() (*Config, error) {
	config := &Config{
		Addr:            ":8080",
		LogLevel:        "info",
		FileStoragePath: "/tmp/metrics-db.json",
		StoreInterval:   "300s",
		Restore:         true,
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
		return nil, ErrUnknownArguments
	}

	cfgutils.ParseString("l", "LOG_LEVEL", "log level", &config.LogLevel)
	cfgutils.ParseString("f", "FILE_STORAGE_PATH", "file storage path", &config.FileStoragePath)
	cfgutils.ParseBool("r", "RESTORE", "restore metrics when server starts", &config.Restore)
	cfgutils.ParseString("d", "DATABASE_DSN", "database DSN", &config.DatabaseDSN)
	cfgutils.ParseString("k", "KEY", "very very very secret key", &config.Secret)
	cfgutils.ParseString("crypto-key", "CRYPTO_KEY", "Crypto key", &config.CryptoKey)
	err := cfgutils.ParseStringWithValidator("a", "ADDRESS", "address and port to run server", &config.Addr, utils.ValidateAddress)
	if err != nil {
		return nil, err
	}
	cfgutils.ParseString("i", "STORE_INTERVAL", "metrics store interval", &config.StoreInterval)

	val, err := time.ParseDuration(config.StoreInterval)
	if err != nil {
		return nil, err
	}
	config.StoreIntervalDuration = val

	return config, nil
}
