// Package config Модуль config отвечает за определение конфигурации сервера
package config

import (
	"errors"
	"flag"
	"fmt"
	"time"

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
	Addr string
	// LogLevel уровень логирования
	LogLevel string
	// FileStoragePath путь к файловому хранилищу метрик
	FileStoragePath string
	// DatabaseDSN строка подключения к базе данных хранения метрик
	DatabaseDSN string
	// Secret секретный код для создания и идентификации ключа аутентификации клиентов
	Secret string
	// StoreInterval временной интервал, через который сервер сохраняет состояние метрик в постоянное хранилище
	StoreInterval time.Duration
	// Restore флаг включающий воостановление метрик в память после старта сервера
	Restore bool
}

// GetConfig Функция для получения конфигурации сервера.
// Если параметры не найдены в переменных окружения то берутся значения из флагов либо значения по умолчанию
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
