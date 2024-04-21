package utils

import (
	"os"
	"strconv"
)

func TryTakeStringFromEnv(name string, target *string) {
	if fromEnv := os.Getenv(name); fromEnv != "" {
		*target = fromEnv
	}
}

func TryTakeIntFromEnv(name string, target *int) {
	if fromEnv := os.Getenv(name); fromEnv != "" {
		if v, err := strconv.Atoi(fromEnv); err == nil {
			*target = v
		}
	}
}

func TryGetBoolFromEnv(name string, target *bool) {
	if fromEnv := os.Getenv(name); fromEnv != "" {
		if v, err := strconv.ParseBool(fromEnv); err == nil {
			*target = v
		}
	}
}
